# SilverBullet git credentials

Drop the deploy key for the notes repo here as `id_ed25519` (plus `known_hosts`
for the Gitea host). The directory is mounted read-only into the `silverbullet-git`
sidecar at `/ssh`, and referenced via `GIT_SSH_COMMAND`.

Generate and register one with:

    ssh-keygen -t ed25519 -N "" -f silverbullet/ssh/id_ed25519 -C silverbullet
    ssh-keyscan -p 2222 <gitea-host> > silverbullet/ssh/known_hosts

Then add the public key as a write-enabled deploy key on the notes repo in Gitea.
Key material is gitignored; only this README is committed.

The sidecar runs as uid 1000, so the private key must be readable by that uid.

To use an HTTPS token instead, skip the key and embed the token in the clone's
remote URL on the host — the sidecar just runs `git push` against whatever
`origin` the clone already has.
