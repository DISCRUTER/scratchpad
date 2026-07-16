# Builder Stage
FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Runtime Stage
FROM alpine:3.24 as runtime

WORKDIR /app

COPY --from=builder /app/main .

RUN addgroup --system appuser && adduser --system --group appuser
USER appuser

EXPOSE 8080

CMD [ "./main" ]