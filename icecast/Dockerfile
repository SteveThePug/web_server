FROM debian:latest as builder

WORKDIR /app

RUN apt-get update \
    && apt-get install --yes icecast2

COPY icecast.xml.template /etc/icecast2/icecast.xml.template
COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT [ "/entrypoint.sh" ]
