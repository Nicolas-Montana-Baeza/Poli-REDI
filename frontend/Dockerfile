FROM docker.io/library/node:22-alpine3.22 AS builder

WORKDIR /src

COPY package*.json ./
RUN npm ci

COPY . .

RUN npm run build


FROM docker.io/library/caddy:2.11.4-alpine

COPY Caddyfile /etc/caddy/Caddyfile
COPY --from=builder /src/dist /srv
