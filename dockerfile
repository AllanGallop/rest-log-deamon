# Use the Go image to build the application
FROM golang:1.23-alpine AS builder

# Create a working directory in the container
WORKDIR /app

# Copy go.mod and go.sum into the container
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the entire source code into the container
COPY . .

# Build the application
RUN go build -o /go-server

# Use a minimal base image to run the compiled binary
FROM alpine:3.15

# Set the working directory
WORKDIR /

# Copy the compiled binary from the builder stage
COPY --from=builder /go-server /go-server

# Expose port 8080 for the application
EXPOSE 8080

# Command to run the application
CMD ["/go-server"]
