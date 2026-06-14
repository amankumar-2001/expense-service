# syntax=docker/dockerfile:1

# ---- build stage ----
FROM golang:1.25-alpine AS build

WORKDIR /src

# Cache module downloads independently of source changes.
COPY go.mod go.sum ./
RUN go mod download

# Build a static binary (CGO off => no libc dependency at runtime).
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/expense-service ./cmd

# ---- runtime stage ----
FROM alpine:3.20

# ca-certificates: required for TLS to Neon (Postgres) and Upstash (Redis).
RUN apk add --no-cache ca-certificates && adduser -D -u 10001 appuser

WORKDIR /app

COPY --from=build /out/expense-service /app/expense-service
# Only the non-secret JSON config ships in the image. The JWT public key is
# provided at runtime via Render Secret Files (see render.yaml / DEPLOYMENT.md).
COPY assets/kharchibook /app/assets/kharchibook

USER appuser

# Matches SERVER_PORT in render.yaml (Render's default injected PORT is 10000).
EXPOSE 10000

ENTRYPOINT ["/app/expense-service"]
