#!/usr/bin/env python3
# =============================================================================
# apply.py — the host half of the admin panel's Environment tab.
#
# WHAT IT DOES
#   Takes the change the backend queued in env-apply/run/request.json, merges
#   it into ./.env, recreates the containers that read the changed variables,
#   and reports the outcome in env-apply/run/status.json for the panel to show.
#
# WHY IT RUNS ON THE HOST
#   The backend is the internet-facing process. It is given neither write
#   access to .env nor the Docker socket, so the most a hijacked admin session
#   can do is ask for a change to a variable this script agrees is editable.
#   The list of those is backend/services/env_registry.json, read here from
#   the checkout — the same file the backend embeds.
#
# WHEN IT RUNS
#   On the Pi, started by the systemd path unit that install.sh sets up, the
#   moment request.json appears. By hand anywhere else:
#     python3 env-apply/apply.py
#   On a dev machine pass the compose files the stack was started with:
#     python3 env-apply/apply.py \
#       --compose-cmd "docker compose -f docker-compose.yml -f docker-compose.dev.yml"
#
# SAFETY
#   - The request is claimed by rename before anything else, so it is consumed
#     exactly once and its secrets leave the shared directory even on failure.
#   - .env is replaced atomically (temp file + rename) and the previous
#     version kept as .env.bak. If compose then rejects the file, the backup
#     is put back before anything is recreated.
#   - Only the services that read a changed variable are recreated (plus the
#     backend, whose own environment is what the panel reports). A plain
#     `docker compose up -d` would restart Postgres and nginx for any edit,
#     because they load the whole file too; the next deploy catches them up.
#   - An flock on run/lock keeps this from overlapping itself or the deploy
#     workflow, which takes the same lock.
#
# Standard library only: it has to run on a bare Raspberry Pi OS.
# =============================================================================
import argparse
import datetime
import fcntl
import json
import os
import re
import shlex
import shutil
import subprocess
import sys
import tempfile

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))

# Values made only of these are written bare; anything else is single-quoted,
# which dotenv reads literally (no $ interpolation, no # comments).
BARE_VALUE = re.compile(r"^[A-Za-z0-9_./:@+,=-]*$")

# How much compose output a failed run reports back to the panel.
OUTPUT_TAIL_LINES = 40


def now():
    return datetime.datetime.now(datetime.timezone.utc).isoformat(timespec="seconds")


def format_line(name, value):
    if BARE_VALUE.match(value):
        return f"{name}={value}"
    return f"{name}='{value}'"


def merge(text, changes):
    """Return .env text with each changed variable's line replaced in place.

    Comments, blank lines and ordering are kept. A variable with no line yet
    is appended.
    """
    remaining = dict(changes)
    out = []
    for line in text.splitlines():
        match = re.match(r"^\s*(export\s+)?([A-Za-z_][A-Za-z0-9_]*)\s*=", line)
        if match and match.group(2) in changes:
            name = match.group(2)
            out.append((match.group(1) or "") + format_line(name, changes[name]))
            remaining.pop(name, None)
        else:
            out.append(line)
    if remaining:
        if out and out[-1].strip():
            out.append("")
        out.append("# Added from the admin panel")
        for name in sorted(remaining):
            out.append(format_line(name, remaining[name]))
    return "\n".join(out) + "\n"


def write_atomically(path, text):
    """Replace path, keeping its mode and owner."""
    stat = os.stat(path)
    fd, tmp = tempfile.mkstemp(dir=os.path.dirname(path), prefix=".env.tmp.")
    try:
        with os.fdopen(fd, "w") as handle:
            handle.write(text)
            handle.flush()
            os.fsync(handle.fileno())
        os.chmod(tmp, stat.st_mode & 0o7777)
        try:
            os.chown(tmp, stat.st_uid, stat.st_gid)
        except PermissionError:
            pass  # not root and not the owner; the rename below still works
        os.replace(tmp, path)
    except BaseException:
        if os.path.exists(tmp):
            os.unlink(tmp)
        raise


def load_registry(path):
    with open(path) as handle:
        return {entry["name"]: entry for entry in json.load(handle)}


def check(changes, registry):
    """Refuse anything the backend should never have asked for."""
    if not isinstance(changes, dict) or not changes:
        raise ValueError("the request changes nothing")
    for name, value in changes.items():
        entry = registry.get(name)
        if entry is None or not entry.get("editable"):
            raise ValueError(f"{name} is not editable")
        if not isinstance(value, str):
            raise ValueError(f"{name} is not a string")
        if "'" in value or any(ord(c) < 0x20 or ord(c) == 0x7F for c in value):
            raise ValueError(f"{name} holds a character .env cannot carry")


def services_for(names, registry):
    services = {"backend"}
    for name in names:
        services.update(registry[name].get("services", []))
    return sorted(services)


def run(cmd, cwd):
    result = subprocess.run(
        cmd, cwd=cwd, stdout=subprocess.PIPE, stderr=subprocess.STDOUT, text=True
    )
    return result.returncode, result.stdout


def tail(text):
    return "\n".join(text.strip().splitlines()[-OUTPUT_TAIL_LINES:])


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--env-file", default=os.path.join(ROOT, ".env"))
    parser.add_argument("--run-dir", default=os.path.join(ROOT, "env-apply", "run"))
    parser.add_argument(
        "--registry",
        default=os.path.join(ROOT, "backend", "services", "env_registry.json"),
    )
    parser.add_argument(
        "--compose-cmd",
        default="docker compose",
        help="compose invocation, e.g. to add -f files for a dev stack",
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="report what would change and leave everything untouched",
    )
    args = parser.parse_args()

    compose = shlex.split(args.compose_cmd)
    compose_dir = os.path.dirname(os.path.abspath(args.env_file))
    request_path = os.path.join(args.run_dir, "request.json")
    claimed_path = os.path.join(args.run_dir, "request.claimed")
    status_path = os.path.join(args.run_dir, "status.json")
    backup_path = args.env_file + ".bak"

    def report(state, request_id, **extra):
        body = {"state": state, "id": request_id, "at": now(), **extra}
        tmp = status_path + ".tmp"
        with open(tmp, "w") as handle:
            json.dump(body, handle)
        os.chmod(tmp, 0o644)
        os.replace(tmp, status_path)

    if args.dry_run:
        if not os.path.exists(request_path):
            print("No request pending.")
            return 0
        with open(request_path) as handle:
            request = json.load(handle)
        registry = load_registry(args.registry)
        check(request.get("set"), registry)
        names = sorted(request["set"])
        print("Would set:", ", ".join(names))
        print("Would recreate:", ", ".join(services_for(names, registry)))
        return 0

    with open(os.path.join(args.run_dir, "lock"), "w") as lock:
        fcntl.flock(lock, fcntl.LOCK_EX)

        # Claim first: after this rename the request cannot be applied twice,
        # and a crash below leaves no secrets where the backend can read them.
        try:
            os.replace(request_path, claimed_path)
        except FileNotFoundError:
            print("No request pending.")
            return 0
        try:
            with open(claimed_path) as handle:
                request = json.load(handle)
        finally:
            os.unlink(claimed_path)

        request_id = str(request.get("id", ""))
        try:
            registry = load_registry(args.registry)
            changes = request.get("set")
            check(changes, registry)
            names = sorted(changes)
            services = services_for(names, registry)
            report("applying", request_id, changed=names, services=services)

            with open(args.env_file) as handle:
                before = handle.read()
            shutil.copy2(args.env_file, backup_path)
            os.chmod(backup_path, 0o600)
            write_atomically(args.env_file, merge(before, changes))

            # Catch a file compose cannot parse before any container is
            # stopped for it.
            code, output = run(compose + ["config", "-q"], compose_dir)
            if code != 0:
                write_atomically(args.env_file, before)
                raise RuntimeError(
                    "compose rejected the new .env; it was restored.\n" + tail(output)
                )

            code, output = run(
                compose + ["up", "-d", "--no-deps", "--force-recreate"] + services,
                compose_dir,
            )
            if code != 0:
                # .env keeps the new values: they are valid, and the usual
                # cause here is a container that would not start, which the
                # old values would not fix.
                raise RuntimeError(
                    ".env was updated but recreating the containers failed.\n"
                    + tail(output)
                )
        except Exception as error:  # noqa: BLE001 - everything is reported
            report("failed", request_id, output=str(error))
            print(f"Failed: {error}", file=sys.stderr)
            return 1

        report("applied", request_id, changed=names, services=services)
        print(f"Applied {', '.join(names)}; recreated {', '.join(services)}.")
        return 0


if __name__ == "__main__":
    sys.exit(main())
