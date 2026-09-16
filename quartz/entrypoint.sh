#!/bin/sh
# =============================================================================
# quartz entrypoint — render the config, then build-and-serve the notes site.
#
# WHAT IT DOES
#   Quartz turns the Obsidian vault into a static site. Its config is a
#   TypeScript file that has to know the public base URL, which is only known
#   at runtime, so quartz.config.ts is generated from a template on each start.
#
# THE envsubst ALLOW-LIST
#   Only ${DOMAIN} is substituted. This matters more here than elsewhere: the
#   template is TypeScript and full of `${...}` -shaped syntax and dollar signs
#   that must NOT be touched. Restricting envsubst to a single name guarantees
#   it can only ever rewrite the one intended line (baseUrl).
#
# WHY build --serve RATHER THAN A STATIC BUILD
#   --serve leaves Quartz running as an HTTP server AND watching the content
#   directory, so editing a note in Obsidian on the host re-renders the site
#   live with no container restart. The vault is bind-mounted READ-ONLY
#   (docker-compose.yml), so Quartz can watch it but never write to it.
#   The trade-off is that this is a long-running Node process on a Pi rather
#   than pre-built static files nginx could serve for free.
#
# ACCESS CONTROL
#   None here — Quartz is wide open on its port inside app-network. The site is
#   private because nginx gates /notes/ behind an `auth_request` admin check
#   before proxying. Never publish this container's port to the host.
#
# ENV VARS
#   DOMAIN       required; becomes baseUrl "www.<DOMAIN>/notes"
#   QUARTZ_PORT  optional, defaults to 8080; must match the QUARTZ_PORT nginx
#                builds its upstream from, or /notes/ 502s.
#
# `exec` so the Node process becomes PID 1 and receives docker's signals.
# =============================================================================
set -e

envsubst '${DOMAIN}' \
    </quartz/quartz.config.ts.template \
    >/quartz/quartz.config.ts

exec npx quartz build --serve --port "${QUARTZ_PORT:-8080}"
