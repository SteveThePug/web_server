FROM debian:latest as builder

WORKDIR /app


RUN apt-get update \
    && apt-get install --yes icecast2 gettext

RUN useradd radio
RUN chown -R radio:radio /etc/icecast2 /var/log/icecast2
USER radio

COPY icecast.xml.template /etc/icecast2/icecast.xml.template
COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT [ "/entrypoint.sh" ]
