FROM alpine:latest

# Build argument for architecture (amd64 or arm64)
ARG ARCH=amd64

# Install ca-certificates for HTTPS connections
RUN apk --no-cache add ca-certificates tzdata

# Create app directory
WORKDIR /app

# Copy the pre-built binary for the specified architecture
COPY build/linux-${ARCH}/release/go-cert-provider /app/go-cert-provider

# Make the binary executable
RUN chmod +x /app/go-cert-provider

# Expose the default server port
EXPOSE 8080

# Set the entrypoint to the serve command
ENTRYPOINT ["/app/go-cert-provider", "certs", "serve"]

# Default flags (can be overridden)
CMD ["--listen-port", "8080", "--listen-addr", "0.0.0.0"]
