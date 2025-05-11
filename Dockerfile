# Stage 1: Build the Go app
FROM golang:1.24 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod tidy
COPY cmd/ ./cmd/
COPY . .
ENV CGO_ENABLED=0 GOOS=linux
RUN go build -a -installsuffix cgo -o main ./cmd/

# Stage 2: Create minimal image
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE 8080
CMD ["./main"]
