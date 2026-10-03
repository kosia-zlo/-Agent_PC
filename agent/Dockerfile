FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY agent_linux /app/agent
# By default, the agent expects -server flag, but we can use ENV for convenience
ENV SERVER_URL=http://host.docker.internal:8080
ENTRYPOINT ["/app/agent"]
