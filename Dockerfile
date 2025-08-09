FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache curl git unzip wget zip

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o transaction-consumer ./cmd/main.go

# Final stage - Using distroless for maximum security
FROM gcr.io/distroless/static:nonroot

# Copy the binary from builder stage
COPY --from=builder /app/transaction-consumer /transaction-consumer

# The distroless nonroot image already runs as non-root user (65532)
USER nonroot:nonroot

# Run the application
CMD ["/transaction-consumer"]