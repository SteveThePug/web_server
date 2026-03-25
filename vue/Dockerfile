FROM node:22-slim
RUN apt-get update && apt-get install -y make git && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
CMD ["sh", "-c", "npm run build -- --outDir /output --emptyOutDir"]
