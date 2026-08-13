FROM golang:1.26.5-alpine3.24 AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 \
    GOOS=linux \
    go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/server \
    ./cmd


FROM alpine:3.24.1

RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    && addgroup -S poliredi \
    && adduser -S -G poliredi poliredi

WORKDIR /app

COPY --from=builder --chown=poliredi:poliredi \
    /out/server \
    /app/server

USER poliredi

EXPOSE 3000

ENTRYPOINT ["/app/server"]
