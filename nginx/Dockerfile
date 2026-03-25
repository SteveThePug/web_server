FROM nginx:latest
RUN rm -rf /etc/nginx/html/* && \
    apt-get update && apt-get install -y gettext-base && \
    rm -rf /var/lib/apt/lists/*
COPY nginx.conf.template /etc/nginx/nginx.conf.template
COPY nginx_setup.conf.template /etc/nginx/nginx_setup.conf.template
COPY nginx_dev.conf.template /etc/nginx/nginx_dev.conf.template
COPY entrypoint.sh /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
