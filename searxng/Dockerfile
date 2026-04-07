FROM searxng/searxng:latest

RUN apk add --no-cache gettext

COPY settings.yml.template /searxng/settings.yml.template
COPY entrypoint.sh /searxng/entrypoint.sh
RUN chmod +x /searxng/entrypoint.sh

ENTRYPOINT ["/searxng/entrypoint.sh"]
