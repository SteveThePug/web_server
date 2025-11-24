FROM debian:latest as builder

WORKDIR /app


RUN apt-get update \
    && apt-get install --yes icecast2 gettext media-types
# RUN apt-get install --yes liquidsoap

RUN useradd radio
RUN chown -R radio:radio /etc/icecast2 /var/log/icecast2
# RUN chown -R radio:radio /etc/liquidsoap /var/log/liquidsoap
USER radio

COPY icecast.xml.template /etc/icecast2/icecast.xml.template
# COPY stream.liq.template /etc/liquidsoap/stream.liq.template
COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT [ "/entrypoint.sh" ]
