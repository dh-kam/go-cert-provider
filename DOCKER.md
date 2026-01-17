# Docker Usage Guide

## Building Docker Image

```bash
make build-image
```

This command:
1. **Automatically detects your system architecture** (x86_64/amd64 or arm64)
2. Builds the Linux release binary for the detected architecture
   - x86_64 systems: builds `linux-amd64-release`
   - arm64 systems: builds `linux-arm64-release`
3. Creates an Alpine Linux-based Docker image
4. Copies the binary to the `/app` directory
5. Configures the server to start with the `certs serve` command

> **Note**: The Docker image is automatically created for your build system's architecture.

## Running Docker Image

### Basic Usage (default port 8080)

```bash
docker run -p 8080:8080 \
  -e JWT_SECRET_KEY="your-secret-key" \
  -e PORKBUN_API_KEY="your-api-key" \
  -e PORKBUN_SECRET_KEY="your-secret-key" \
  -e PORKBUN_DOMAINS="example.com,test.com" \
  go-cert-provider:latest
```

### Run with Custom Port

```bash
docker run -p 9000:9000 \
  -e JWT_SECRET_KEY="your-secret-key" \
  -e PORKBUN_API_KEY="your-api-key" \
  -e PORKBUN_SECRET_KEY="your-secret-key" \
  go-cert-provider:latest \
  --listen-port 9000
```

### Using Environment File

Create `.env` file:
```env
JWT_SECRET_KEY=your-secret-key-here
PORKBUN_API_KEY=your-porkbun-api-key
PORKBUN_SECRET_KEY=your-porkbun-secret-key
PORKBUN_DOMAINS=example.com,test.com
```

Run:
```bash
docker run -p 8080:8080 --env-file .env go-cert-provider:latest
```

### Docker Compose Example

`docker-compose.yml`:
```yaml
version: '3.8'

services:
  cert-provider:
    image: go-cert-provider:latest
    ports:
      - "8080:8080"
    environment:
      - JWT_SECRET_KEY=${JWT_SECRET_KEY}
      - PORKBUN_API_KEY=${PORKBUN_API_KEY}
      - PORKBUN_SECRET_KEY=${PORKBUN_SECRET_KEY}
      - PORKBUN_DOMAINS=example.com,test.com
    restart: unless-stopped
```

Run:
```bash
docker-compose up -d
```

## Environment Variables

| Variable | Description | Required |
|---------|------|------|
| `JWT_SECRET_KEY` | Secret key for JWT token signing | Yes |
| `PORKBUN_API_KEY` | Porkbun API key | Yes (when using Porkbun) |
| `PORKBUN_SECRET_KEY` | Porkbun Secret key | Yes (when using Porkbun) |
| `PORKBUN_DOMAINS` | Comma-separated list of domains to manage | No |

## Additional Command Options

The container runs the `certs serve` command by default. You can pass additional flags:

```bash
docker run -p 8080:8080 \
  --env-file .env \
  go-cert-provider:latest \
  --listen-port 8080 \
  --listen-addr 0.0.0.0
```

## Health Check

Check if the server is running properly:

```bash
curl http://localhost:8080/health
```

Example response:
```json
{
  "status": "ok",
  "version": "v0.9.0",
  "providers": ["porkbun"],
  "domains": ["example.com", "test.com"]
}
```

## Image Information

- **Base Image**: Alpine Linux (latest)
- **Image Size**: 
  - amd64: ~56MB
  - arm64: ~53MB
- **Supported Architectures**: 
  - x86_64 (amd64)
  - aarch64 (arm64)
- **Contents**: 
  - Statically linked binary
  - CA certificates
  - Timezone data
- **Default Port**: 8080
- **Working Directory**: `/app`

> **Multi-arch Build**: `make build-image` automatically detects your system architecture and builds an image for that architecture.
