# Stage 1: Builder
FROM golang:1.22-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w" -o /out/server .

# Stage 2: Runtime
FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /out/server /app/server
RUN mkdir -p /data /app/updates
ENV LISTEN_ADDR=0.0.0.0:8080 DB_PATH=/data/server.db UPDATE_DIR=/app/updates
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s \
  CMD wget -qO- http://127.0.0.1:8080/api/agents || exit 1
ENTRYPOINT ["/app/server"]
