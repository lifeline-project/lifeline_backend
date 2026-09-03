# Stage 1: Build the Go binary
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy dependency manifests
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/lifeline_backend ./cmd/server/main.go

# Stage 2: Create a minimal production image
FROM alpine:3.18

WORKDIR /app

# Copy binary and configuration files
COPY --from=builder /app/lifeline_backend .
COPY configs/.env ./configs/.env

EXPOSE 8080

CMD ["./lifeline_backend"]
