# Use a glibc-based Go image for the builder stage
FROM golang:1.21 AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the go.mod and go.sum files to download dependencies
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code (including cmd/main.go)
COPY . .

# Build the Go application
RUN GOOS=linux GOARCH=arm64 go build -o /client-message-transformer ./cmd/main.go

# Use a Debian Bookworm-based image for the final stage
FROM debian:bookworm-slim

# Install ca-certificates and librdkafka for runtime
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    librdkafka1 \
    && rm -rf /var/lib/apt/lists/*

# Copy the binary from the builder stage
COPY --from=builder /client-message-transformer /client-message-transformer

# Copy .env file if your application uses it
# COPY .env ./

# Expose any necessary ports (uncomment and adjust if needed)
# EXPOSE 8080

# Set the entry point to run the binary
ENTRYPOINT ["/client-message-transformer"]