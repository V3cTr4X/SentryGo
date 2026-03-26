FROM golang:1.25 AS builder

WORKDIR /app

RUN git clone https://github.com/v3ctr4x/SentryGo.git .  # Cambia esto por tu URL del repositorio

RUN go mod tidy

RUN GOOS=linux GOARCH=amd64 go build -o sentrygo-server-linux cmd/server/main.go
RUN GOOS=linux GOARCH=amd64 go build -o sentrygo-agent-linux cmd/agent/main.go

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y \
    curl \
    sqlite3 \
    ca-certificates \
    git \
    libc6 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /app/sentrygo-server-linux /app/
COPY --from=builder /app/sentrygo-agent-linux /app/

COPY --from=builder /app/web /app/web/

COPY config.yaml /app/

EXPOSE 8080

CMD ["./sentrygo-server-linux"]