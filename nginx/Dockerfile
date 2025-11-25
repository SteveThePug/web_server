FROM nginx:latest

# Copy template config
COPY nginx.conf.template /etc/nginx/nginx.conf.template
COPY nginx_setup.conf.template /etc/nginx/nginx_setup.conf.template

COPY entrypoint.sh /entrypoint.sh

# Install dependencies
RUN apt-get update && apt-get install -y \
    gettext-base \
    curl \
    build-essential \
    git \
    nodejs\
    && npm install -g npm@latest

WORKDIR /app
COPY vue/ ./
RUN npm install
RUN npm run build

RUN rm -rf /etc/nginx/html/* \
    && cp -r ./dist/* /etc/nginx/html/

ENTRYPOINT ["/entrypoint.sh"]
