FROM moul/icecast

# Copy template + entrypoint into image
COPY icecast.xml.template /etc/icecast2/icecast.xml.template
COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT [ "/entrypoint.sh" ]
