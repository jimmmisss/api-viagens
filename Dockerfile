# Stage 1: Build the application
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum to download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the application
# CGO_ENABLED=0 is important for a static binary
# -o /app/server builds the binary into the /app/server file
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/api

# Stage 2: Create the final, minimal image
FROM alpine:latest

WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/server .
# Copy migrations to be able to run them from the container if needed
COPY migrations ./migrations

# Expose the port the app runs on
EXPOSE 8080

# Command to run the executable
CMD ["./server"]