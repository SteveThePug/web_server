FROM debian:latest as builder

WORKDIR /app

COPY icecast.xml.template /etc/icecast2/icecast.xml.template
COPY entrypoint.sh /entrypoint.sh

RUN apt-get update \
    && apt-get install --yes icecast2 gettext

RUN useradd radio
RUN chown -R radio:radio /etc/icecast2 /var/log/icecast2 /usr/local/icecast/
USER radio

ENTRYPOINT [ "/entrypoint.sh" ]
