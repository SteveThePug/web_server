FROM pltnk/icecast2

# Copy template + entrypoint into image
COPY icecast.xml.template /etc/icecast2/icecast.xml.template
COPY entrypoint.sh /entrypoint.sh

# Ensure script is executable
RUN chmod +x /entrypoint.sh

ENTRYPOINT [ "entrypoint.sh" ]
