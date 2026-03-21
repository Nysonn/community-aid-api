# Stage 1 — migrate: goose migrations runner
FROM golang:1.24-alpine AS migrate
RUN apk add --no-cache netcat-openbsd
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

# Stage 2 — dev: hot-reload with Air
FROM golang:1.25-alpine AS dev
RUN go install github.com/air-verse/air@latest
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
EXPOSE 3000
CMD ["air", "-c", ".air.toml"]

# Stage 3 — builder: compile the production binary
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ENV CGO_ENABLED=0
RUN go build -ldflags="-s -w" -o /app/cmd-server ./cmd/server

# Stage 4 — production: minimal runtime image
FROM alpine:latest AS production
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/cmd-server /app/cmd-server
RUN adduser -D app && chown app:app /app/cmd-server
USER app
EXPOSE 3000
ENV GIN_MODE=release
ENTRYPOINT ["/app/cmd-server"]
