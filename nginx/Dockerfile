FROM nginx:latest


# Copy template config
COPY nginx.conf.template /etc/nginx/nginx.conf.template
COPY nginx_setup.conf.template /etc/nginx/nginx_setup.conf.template

COPY entrypoint.sh /entrypoint.sh

# Update package list and install gettext-base
RUN apt-get update && apt-get install -y gettext-base

# Replace variables at build time
ENV BACKEND_HOST=backend
ENV BACKEND_PORT=8080
ENV DOMAIN=example.com

ENTRYPOINT ["/entrypoint.sh"]
