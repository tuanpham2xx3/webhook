# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o webhook-deploy .

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests and git for deployment
RUN apk --no-cache add ca-certificates git

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/webhook-deploy .

# Expose port
EXPOSE 8300

# Run the binary
CMD ["./webhook-deploy"] 