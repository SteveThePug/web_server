#!/usr/bin/env bash
# =============================================================================
# edge_test.sh — what the public edge does with a fixed set of requests.
#
# WHAT IT DOES
#   Sends each request below to the edge and prints one line per request
#   (status and Location, or one response header). Two kinds of line:
#     ok / FAIL  the privacy boundary and the gates, which have a fixed right
#                answer: the /notes/ denials, /py/internal/, every gated
#                prefix. Any FAIL makes the exit status 1.
#     show / hdr everything else (redirects, headers, carve-outs, uploads),
#                whose answer depends on the upstream's data. Compare two runs
#                instead (below).
#
# USAGE
#   caddy/edge_test.sh <base-url> [curl options...]
#     base-url      the www origin, e.g. https://www.example.com
#     curl options  added to every request: --resolve to reach a given
#                   container, -k for dev's internal CA.
#   Environment, all optional:
#     EDGE_APEX     the apex origin (production only): adds the apex cases.
#     EDGE_HTTP     the plain-http apex origin (production only), e.g.
#                   http://example.com: adds the http -> https case.
#     EDGE_COOKIE   an admin session's Cookie header value: adds the
#                   signed-in cases (each gated prefix must then NOT redirect
#                   to the login).
#     EDGE_UPLOAD   the name of a file under /uploads/ that exists.
#
#   On stpvm, pin the hostname to this machine, or curl goes out over the
#   internet and back (and hangs):
#     caddy/edge_test.sh https://www.$DOMAIN --resolve www.$DOMAIN:443:127.0.0.1
#   Dev stack:
#     caddy/edge_test.sh https://localhost -k
#
# COMPARING TWO EDGES (the nginx -> Caddy cutover)
#   Run it against each and diff the output:
#     caddy/edge_test.sh https://www.$DOMAIN --resolve www.$DOMAIN:443:127.0.0.1 >nginx.txt
#     caddy/edge_test.sh https://www.$DOMAIN:8443 --resolve www.$DOMAIN:8443:127.0.0.1 >caddy.txt
#     diff nginx.txt caddy.txt
#   Every remaining difference should be one that was intended.
#
# NOTHING HERE WRITES. Every request is a GET, apart from one POST to
# /py/internal/, which never gets past the edge.
# =============================================================================
set -u

if [ $# -lt 1 ]; then
    echo "usage: $0 <base-url> [curl options...]" >&2
    exit 2
fi
BASE=${1%/}
shift
CURL_OPTS=("$@")
FAILED=0
# The www host, without scheme or port.
BASE_HOST=$(printf '%s' "$BASE" | sed -E 's#^[a-z]+://([^/:]+).*#\1#')

# "<status> <location>" for one request. Arguments after the URL are extra
# curl options (method, headers, --path-as-is). A redirect to the www host is
# shown as its path alone: nginx sends an absolute Location built from its
# own port, Caddy a relative one, and either way the browser lands on the
# same page.
probe() {
    local url=$1 out
    shift
    out=$(curl -sS -o /dev/null --max-time 15 -w '%{http_code} %{redirect_url}' \
        "${CURL_OPTS[@]}" "$@" "$url" 2>/dev/null) || out=ERR
    printf '%s' "$out" | sed -E "s#^([0-9]+) [a-z]+://$BASE_HOST(:[0-9]+)?/#\\1 /#"
}

# The first value of one response header, or "-".
header_of() {
    local name=$1 url=$2 value
    shift 2
    value=$(curl -sS -o /dev/null -D - --max-time 15 "${CURL_OPTS[@]}" "$@" "$url" 2>/dev/null |
        tr -d '\r' | grep -i "^$name:" | head -n 1 | cut -d: -f2- | sed 's/^ //')
    printf '%s' "${value:--}"
}

# must <status> <location prefix, or -> <label> <path> [curl options]
must() {
    local want_status=$1 want_loc=$2 label=$3 path=$4 got status loc ok=yes
    shift 4
    got=$(probe "$BASE$path" "$@")
    status=${got%% *}
    loc=${got#* }
    [ "$status" = "$want_status" ] || ok=no
    if [ "$want_loc" != "-" ] && [ "${loc#"$want_loc"}" = "$loc" ]; then
        ok=no
    fi
    if [ $ok = yes ]; then
        printf 'ok    %-12s %-58s %s\n' "$label" "$path" "$got"
    else
        printf 'FAIL  %-12s %-58s %s   (want %s %s)\n' "$label" "$path" "$got" "$want_status" "$want_loc"
        FAILED=1
    fi
}

# not_login <label> <path> [curl options]: anything but a redirect to the login.
not_login() {
    local label=$1 path=$2 got
    shift 2
    got=$(probe "$BASE$path" "$@")
    case $got in
        ERR | "302 /admin/login"*)
            printf 'FAIL  %-12s %-58s %s   (want: let through)\n' "$label" "$path" "$got"
            FAILED=1 ;;
        *) printf 'ok    %-12s %-58s %s\n' "$label" "$path" "$got" ;;
    esac
}

# show <label> <url> [curl options]
show() {
    local label=$1 url=$2
    shift 2
    printf 'show  %-12s %-58s %s\n' "$label" "${url#"$BASE"}" "$(probe "$url" "$@")"
}

# hdr <url> <header...>
hdr() {
    local url=$1 h
    shift
    for h in "$@"; do
        printf 'hdr   %-12s %-58s %s\n' "$h" "${url#"$BASE"}" "$(header_of "$h" "$url")"
    done
}

LOGIN="/admin/login?redirect="

echo "# privacy boundary: must be 404"
for p in \
    /notes/private/x /notes/Private/x /notes/a/private/x /notes/private \
    /notes/templates/t /notes/.fs/private/x.md /notes/.fs/a/private/x.md \
    /notes/.fs/.git/HEAD /notes/a/.obsidian/workspace.json /notes/a/.git/config \
    /notes/.revisions/x.md /notes/private%2Fx /notes/%70rivate/x; do
    must 404 - notes "$p"
done
must 404 - notes /notes/a/../private/x --path-as-is
must 404 - notes /notes/a//private/x --path-as-is
must 404 - py-internal /py/internal/rowing/read -X POST
must 404 - py-internal /py/internal/rowing/read
must 404 - py-internal /py/x/../internal/rowing/read --path-as-is -X POST
must 404 - py-internal /py//internal/rowing/read --path-as-is -X POST

echo "# admin gate, signed out: must redirect to the login, with the original URI"
for p in \
    /sb/ /sb/index.html /jobs/ /jobs/api/graphql /tracker/ /tracker/dashboard \
    /dozzle/ /garage/ /radio/admin/stats.xsl /uploads/private/x.png \
    /gitea/adamf/web_server/commit/0000000 /gitea/adamf/web_server/src/commit/0000000/README.md \
    /gitea/adamf/web_server/archive/main.tar.gz /gitea/adamf/web_server/compare/main...x \
    /gitea/adamf/web_server/blame/commit/0000000/README.md \
    "/jobs/leads?x=1"; do
    must 302 "$LOGIN$p" gate "$p"
done
must 302 "${LOGIN}/uploads/x/../private/x.png" gate /uploads/x/../private/x.png --path-as-is
show gate "$BASE/SB/"
must 302 "${LOGIN}/receipt/" permission /receipt/

if [ -n "${EDGE_COOKIE:-}" ]; then
    echo "# admin gate, signed in: must be let through"
    for p in /sb/ /jobs/ /tracker/ /dozzle/ /garage/ /radio/admin/stats.xsl \
        /gitea/adamf/web_server/commit/0000000 /receipt/; do
        not_login gate "$p" -H "Cookie: $EDGE_COOKIE"
    done
fi

echo "# carve-outs and public routes (status depends on the upstream)"
for p in /jobs/api/capture /jobs/job-capture.user.js /tracker/api/v1/events \
    /tracker/tracker.user.js /tracker/readyz /usaco/ /notes/ /gitea/ \
    /gitea/adamf/web_server/src/branch/main/README.md /radio/status.xsl \
    /py/ /py/docs /api/ /api/me /; do
    show public "$BASE$p"
done

echo "# bare prefixes and moved URLs"
for p in /api /radio /dozzle /garage /gitea /notes /sb /py /jobs /usaco \
    /tracker /receipt /cv/jobs /cv/leads /job-capture.user.js "/gitea?x=1"; do
    show redirect "$BASE$p"
done

echo "# SPA"
for p in /admin /architecture /some-canvas-page /assets/does-not-exist.js; do
    show spa "$BASE$p"
done
hdr "$BASE/" Content-Type Strict-Transport-Security X-Frame-Options \
    X-Content-Type-Options Referrer-Policy Content-Security-Policy Server
hdr "$BASE/api/" Strict-Transport-Security X-Frame-Options Server
hdr "$BASE/gitea/" Strict-Transport-Security X-Frame-Options Server
hdr "$BASE/assets/does-not-exist.js" Cache-Control
hdr "$BASE/img/stamps/mine.gif" Access-Control-Allow-Origin Cache-Control

echo "# uploads"
show uploads "$BASE/uploads/does-not-exist.png"
hdr "$BASE/uploads/does-not-exist.png" X-Content-Type-Options Content-Security-Policy
if [ -n "${EDGE_UPLOAD:-}" ]; then
    show uploads "$BASE/uploads/$EDGE_UPLOAD"
    show uploads "$BASE/uploads/$EDGE_UPLOAD" -H "Range: bytes=0-9"
    hdr "$BASE/uploads/$EDGE_UPLOAD" Content-Type X-Content-Type-Options \
        Content-Disposition Content-Security-Policy Strict-Transport-Security
fi

if [ -n "${EDGE_APEX:-}" ]; then
    APEX=${EDGE_APEX%/}
    echo "# apex (production)"
    show apex "$APEX/x?y=1"
    show apex "$APEX/v2/"
    hdr "$APEX/x" Strict-Transport-Security
fi
if [ -n "${EDGE_HTTP:-}" ]; then
    echo "# plain http (production)"
    show http "${EDGE_HTTP%/}/x?y=1"
fi

echo "# result: $([ $FAILED = 0 ] && echo all MUST cases passed || echo SOME MUST CASES FAILED)"
exit $FAILED
