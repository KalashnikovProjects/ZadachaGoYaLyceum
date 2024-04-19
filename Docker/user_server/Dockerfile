# Build stage
FROM golang:latest AS builder
WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY ./templates ./templates
COPY ./internal ./internal
COPY ./cmd/user_server ./cmd/user_server
COPY ../../proto ./proto

RUN CGO_ENABLED=0 go build -ldflags='-extldflags "-static"' -o user_server ./cmd/user_server

# Final stage
FROM scratch
COPY --from=builder ./app/user_server ./app/user_server

EXPOSE 80
CMD ["/app/user_server"]