# Build stage
FROM golang:latest AS builder
WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY ./internal ./internal
COPY ./cmd/agents ./cmd/agents
COPY ./proto ./proto

RUN CGO_ENABLED=0 go build -ldflags='-extldflags "-static"' -o agents ./cmd/agents

# Final stage
FROM scratch
COPY --from=builder /app/agents /app/agents

EXPOSE 9090
ENTRYPOINT ["/app/agents"]