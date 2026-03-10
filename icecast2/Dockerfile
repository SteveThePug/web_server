FROM savonet/liquidsoap:v2.3.2
USER root
RUN apt-get update \
    && apt-get install --yes icecast2 gettext media-types \
    && rm -rf /var/lib/apt/lists/*
RUN useradd radio
RUN mkdir -p /music
RUN chown -R radio:radio /etc/icecast2 /var/log/icecast2 /music
USER radio
COPY icecast.xml.template /etc/icecast2/icecast.xml.template
COPY stream.liq.template /etc/liquidsoap/stream.liq.template
COPY entrypoint.sh /entrypoint.sh
ENTRYPOINT [ "/entrypoint.sh" ]
