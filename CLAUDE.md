# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

### Development
- **Build**: `go build -o main ./cmd/cli/`
- **Run locally**: `go run ./cmd/cli/ server start`
- **Run with text formatter**: `go run ./cmd/cli/ server start text-formatter`
- **Test**: `go test ./...`
- **Test specific package**: `go test ./internal/jwt`
- **Build Docker image**: `docker build --build-arg VERSION=local -t gateway-proxy .`
- **Start dependencies**: `docker-compose up` (starts postgres, key-management-service, and echo server)

### Docker
- **Run container**: `docker run -p 8080:8080 gateway-proxy`
- The container runs `./main server start` by default

## Architecture

This is a Go HTTP reverse proxy service ("smokey") that acts as an authentication gateway:

### Core Components

1. **CLI Interface** (`cmd/cli/main.go`, `internal/commands/`)
   - Uses Cobra for command structure
   - Main command: `smokey server start`
   - Supports different log formatters (JSON default, text optional)

2. **HTTP Server** (`http/server.go`)
   - Listens on configurable port (default 8080)
   - Uses DataDog distributed tracing and profiling
   - Single endpoint `/` that proxies all requests

3. **Reverse Proxy** (`internal/handlers/proxy.go`)
   - Proxies requests to configured backend (default: apollo-router:4000)
   - Adds custom headers: `x-user-id`, `x-token-purpose`, `x-raw-token`, `x-remote-ip`, `x-user-agent`
   - Strips/modifies original headers like User-Agent and x-forwarded-for
   - Optionally overrides Origin header

4. **JWT Authentication** (`internal/jwt/`)
   - Parses Bearer tokens from Authorization header
   - Extracts subject (user ID) and purpose claims
   - Keys fetched from GraphQL key management service

5. **Key Management** (`internal/keys/`, `internal/poller/`)
   - Fetches JWT public keys from external GraphQL service
   - Background polling every 15 minutes (configurable)
   - Uses generic container for key storage

6. **Middlewares** (`http/middlewares/`)
   - **Logging**: Request/response logging
   - **Metrics**: DataDog metrics collection 
   - **Tracer**: DataDog distributed tracing

### Configuration
- Uses `jinzhu/configor` for config loading
- Config sources: environment variables + `config/config.json`
- Key settings: proxy URL, port, GraphQL endpoint, origin override, polling interval

### Dependencies
The service requires:
- PostgreSQL database (for key management service)
- Key Management Service (GraphQL endpoint for JWT keys)
- Target service to proxy to (e.g., apollo-router)

### Testing
- Standard Go testing with `testify`
- Test files follow `*_test.go` pattern
- Tests exist for JWT parsing, polling, and core utilities