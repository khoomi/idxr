FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build static binary with no CGO dependencies
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64
ENV GIN_MODE=release

RUN go build -ldflags="-s -w" -buildvcs=false -o khoomi ./cmd/khoomi

# Final stage - minimal runtime image
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the static binary from builder stage
COPY --from=builder /app/khoomi .

# Copy .env file if needed
COPY .env .

# Make binary executable
RUN chmod +x khoomi

EXPOSE 8080

CMD ["./khoomi"]
