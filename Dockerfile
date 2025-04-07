FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum* ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /gogo-files

# Use a minimal alpine image for the final stage
FROM alpine:latest

# Install dependencies for runtime
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /

# Copy the binary from builder
COPY --from=builder /gogo-files /gogo-files
# Copy migrations
COPY --from=builder /app/migrations /migrations

# Set environment variables
ENV TZ=UTC

# Run the application
CMD ["/gogo-files"]