FROM --platform=amd64/v8 pltnk/icecast2


RUN apk add --no-cache gettext

# Copy template + entrypoint into image
COPY icecast.xml.template /etc/icecast2/icecast.xml.template
COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT [ "/entrypoint.sh" ]
