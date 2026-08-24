# Builder Stage
FROM docker.io/library/golang:1.26-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /bin/web ./cmd/web

# Runtime Stage
FROM docker.io/library/golang:1.26-alpine AS runtime

WORKDIR /app

COPY --from=builder /bin/web ./web

RUN addgroup -S appuser && adduser -S -G appuser appuser
USER appuser

# The app reads PORT, METRICS_PORT and DSN from the environment and
# expects TLS certs at ./tls/cert.pem and ./tls/key.pem (mount them).
ENV PORT=8080 \
    METRICS_PORT=8081

EXPOSE 8080 8081

ENTRYPOINT ["./web"]
