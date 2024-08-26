# Use the latest Go base image
FROM golang:latest AS builder

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the application
RUN go build -o terminal-notes ./cmd

# Use a smaller base image for the final stage
FROM alpine:latest

# Set the working directory
WORKDIR /root/

# Install necessary libraries for running the Go binary
RUN apk --no-cache add ca-certificates

# Copy the binary from the builder stage
COPY --from=builder /app/terminal-notes .

# Verify the file is present
RUN ls -l /root/terminal-notes

# Expose the port the application listens on (adjust if necessary)
EXPOSE 23236

# Command to run the binary
CMD ["./terminal-notes"]