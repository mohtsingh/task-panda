# Use the official Golang image as a base
FROM golang:1.20-alpine AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod tidy

# Copy the source code from the local 'cmd' directory into the container's '/app' directory
COPY cmd/ ./cmd/

# Run go get to fetch dependencies
RUN go get github.com/labstack/echo/v4
RUN go get github.com/lib/pq

# Build the Go app
RUN go build -o main ./cmd/

# Start a new stage to copy the built binary into a minimal image
FROM alpine:latest

# Install CA certificates to enable HTTPS requests
RUN apk --no-cache add ca-certificates

# Set the Current Working Directory inside the container
WORKDIR /root/

# Copy the Pre-built binary from the previous stage
COPY --from=builder /app/main .

# Expose the port the app runs on
EXPOSE 8080

# Command to run the executable
CMD ["./main"]
