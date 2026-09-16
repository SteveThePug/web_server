# =============================================================================
# Radio image — liquidsoap base with icecast2 added on top.
#
# BASE CHOICE: savonet/liquidsoap:v2.3.2, pinned. Liquidsoap is by far the
# harder of the two to install (OCaml toolchain, many audio codecs), so the
# image starts from the official liquidsoap build and adds icecast2 from apt,
# rather than the other way round. Both daemons then run in one container so
# liquidsoap can reach icecast on localhost — see icecast2/entrypoint.sh.
#
# PACKAGES
#   icecast2     the streaming server itself
#   gettext      provides `envsubst` for rendering the two config templates
#   media-types  the /etc/mime.types database; without it icecast serves the
#                stream and its status pages with wrong content types
#
# THE USER SWITCH: apt needs root, so the image switches to root, installs,
# creates an unprivileged `radio` user, chowns everything the two daemons will
# need to write (icecast config, its logs, the music dir, the liquidsoap config
# dir) and then drops back to `radio` for the rest. Neither daemon runs as
# root, which is also why icecast.xml sets <chroot>0</chroot> — an unprivileged
# process cannot chroot anyway.
#
# The chown of /music matters even though it is mounted read-only at runtime:
# the mount inherits ownership from the host, and the directory must exist
# before the mount lands.
#
# The apt install is Debian-based and multi-arch; no platform pin needed on the Pi.
# =============================================================================
FROM savonet/liquidsoap:v2.3.2
USER root
RUN apt-get update \
    && apt-get install --yes icecast2 gettext media-types \
    && rm -rf /var/lib/apt/lists/*
RUN useradd radio
RUN mkdir -p /music /etc/liquidsoap
RUN chown -R radio:radio /etc/icecast2 /var/log/icecast2 /music /etc/liquidsoap
USER radio
COPY icecast.xml.template /etc/icecast2/icecast.xml.template
COPY stream.liq.template /etc/liquidsoap/stream.liq.template
COPY entrypoint.sh /entrypoint.sh
ENTRYPOINT [ "/entrypoint.sh" ]
