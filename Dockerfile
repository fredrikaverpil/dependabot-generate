# Start from the official Go image.
FROM golang:1.24-alpine

# Copy the entire repository into the /app directory in the container.
WORKDIR /app
COPY . .

# Compile the Go application. The -o flag places the output binary at /app/dependabot-generate.
RUN go build -o /app/dependabot-generate ./cmd/dependabot-generate

# Copy the entrypoint script into the container.
COPY entrypoint.sh /entrypoint.sh

# Make the entrypoint script executable.
RUN chmod +x /entrypoint.sh

# Set the container's entrypoint to our script.
ENTRYPOINT ["/entrypoint.sh"]
