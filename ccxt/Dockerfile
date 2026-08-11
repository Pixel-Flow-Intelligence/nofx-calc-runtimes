ARG NODE_VERSION=22-alpine

FROM node:${NODE_VERSION} AS ccxt-release

RUN apk add --no-cache ca-certificates tzdata busybox-extras

WORKDIR /app/backend/ccxt

COPY app/backend/ccxt/package.json app/backend/ccxt/package-lock.json ./
RUN npm install

COPY app/backend/ccxt ./
COPY proto /app/proto

EXPOSE 50061

HEALTHCHECK --interval=30s --timeout=10s --start-period=20s --retries=3 \
  CMD nc -z 127.0.0.1 50061 || exit 1

CMD ["npm", "run", "start"]
