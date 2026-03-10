# Stage 1: Build Vue app
FROM node:22-slim AS build
WORKDIR /app
COPY vue/package.json vue/package-lock.json ./
RUN npm ci
COPY vue/ ./
RUN npm run build

# Stage 2: Serve with nginx
FROM nginx:latest
RUN rm -rf /etc/nginx/html/* && \
    apt-get update && apt-get install -y gettext-base && \
    rm -rf /var/lib/apt/lists/*
COPY --from=build /app/dist /etc/nginx/html/
COPY nginx.conf.template /etc/nginx/nginx.conf.template
COPY nginx_setup.conf.template /etc/nginx/nginx_setup.conf.template
COPY nginx_dev.conf.template /etc/nginx/nginx_dev.conf.template
COPY robots.txt /etc/nginx/html/robots.txt
COPY entrypoint.sh /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
