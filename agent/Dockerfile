FROM golang:1.22-alpine AS builder
WORKDIR /src
COPY . .
RUN go mod init agent && go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w" -o /out/agent .

FROM alpine:3.19
COPY --from=builder /out/agent /app/agent
ENTRYPOINT ["/app/agent"]
