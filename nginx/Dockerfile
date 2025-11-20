FROM nginx:latest


# Copy template config
COPY nginx.conf.template /etc/nginx/nginx.conf.template
COPY nginx_setup.conf.template /etc/nginx/nginx_setup.conf.template

COPY entrypoint.sh /entrypoint.sh

# Update package list and install gettext-base
RUN apt-get update && apt-get install -y gettext-base

# Replace variables at build time
ENV BACKEND_HOST=backend
ENV DOMAIN=example.com
ENV CERT_PATH=/etc/letsencrypt/live/$DOMAIN


ENTRYPOINT ["/entrypoint.sh"]
