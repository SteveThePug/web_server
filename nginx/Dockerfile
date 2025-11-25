FROM nginx:latest

# Install dependencies needed to add NodeSource repo
RUN apt-get update && apt-get install -y \
    curl \
    build-essential \
    git \
    gettext-base \
    && rm -rf /var/lib/apt/lists/*

# Install Node.js LTS + npm
RUN curl -fsSL https://deb.nodesource.com/setup_lts.x | bash - \
    && apt-get install -y nodejs \
    && npm install -g npm@latest

COPY nginx.conf.template /etc/nginx/nginx.conf.template
COPY nginx_setup.conf.template /etc/nginx/nginx_setup.conf.template
COPY entrypoint.sh /entrypoint.sh

WORKDIR /app
COPY vue/ ./
RUN npm install
RUN npm run build

RUN mkdir -p /etc/nginx/html \
    && rm -rf /etc/nginx/html/* \
    && cp -r ./dist/* /etc/nginx/html/

ENTRYPOINT ["/entrypoint.sh"]
