# Build stage
FROM golang:latest AS builder
WORKDIR /app

COPY ../../go.mod .
COPY ../../go.sum .

RUN go mod download

COPY ../../internal ./internal
COPY ../../pkg ./pkg
COPY ../../cmd/orchestrator ./cmd/orchestrator
COPY ../../proto ./proto

RUN CGO_ENABLED=0 go build -ldflags='-extldflags "-static"' -o orchestrator ./cmd/orchestrator

# Final stage
FROM scratch
COPY --from=builder ./app/orchestrator ./app/orchestrator

EXPOSE 8080
CMD ["/app/orchestrator"]