# syntax=docker/dockerfile:1.7

FROM golang:1.23-alpine AS development

WORKDIR /src
RUN apk add --no-cache ca-certificates git tzdata

COPY go.mod go.sum ./
RUN go mod download
COPY . .

EXPOSE 8080
CMD ["go", "run", "./cmd/api"]

FROM golang:1.23-alpine AS build

WORKDIR /src
RUN apk add --no-cache ca-certificates git tzdata

COPY go.mod go.sum ./
RUN go mod download
COPY . .

ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w -X main.version=${VERSION}" \
    -o /out/dinar-api ./cmd/api

FROM alpine:3.20 AS runtime

RUN apk add --no-cache ca-certificates tzdata wget \
    && addgroup -S -g 10001 dinar \
    && adduser -S -D -H -u 10001 -G dinar dinar

WORKDIR /app
COPY --from=build /out/dinar-api /app/dinar-api
COPY --from=build /src/db/migrations /app/db/migrations

USER dinar
EXPOSE 8080
HEALTHCHECK --interval=15s --timeout=5s --start-period=20s --retries=5 \
  CMD wget -q -O - http://127.0.0.1:8080/healthz >/dev/null || exit 1

ENTRYPOINT ["/app/dinar-api"]
