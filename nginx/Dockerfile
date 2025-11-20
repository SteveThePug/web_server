FROM nginx:latest


# Copy template config
COPY nginx.conf.template /etc/nginx/nginx.conf.template

# Update package list and install gettext-base
RUN apt-get update && apt-get install -y gettext-base

# Replace variables at build time
ENV BACKEND_HOST=backend

CMD /bin/sh -c "envsubst '\$BACKEND_HOST' < /etc/nginx/nginx.conf.template > /etc/nginx/nginx.conf && nginx -g 'daemon off;'"
