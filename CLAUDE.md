# CLAUDE.md

This file provides guidance to Claude Code when working with code in this repository.

## Build Commands

```bash
# Build
make build         # Build the binary (runs tests first)

# Test
make test          # Run unit tests with race detection and coverage

# Quality
make check         # Run linting (golangci-lint), fmt, and vet

# Run locally
make run BUCKET=your-bucket    # Run with stdio transport
make run-http BUCKET=your-bucket  # Run with HTTP transport

# Docker
make docker        # Build Docker image

# Clean
make clean         # Remove build artifacts
```

## Architecture

This is a Go-based MCP (Model Context Protocol) server that provides a single tool:
- **share-html**: Uploads HTML content to Google Cloud Storage and returns a public URL

### Directory Structure

- `cmd/mcp-html-share/` - Application entry point
- `internal/app/` - Application orchestration
- `internal/config/` - Configuration and flags
- `internal/gcs/` - Google Cloud Storage service
- `internal/health/` - Health check and metrics server
- `internal/server/` - MCP server implementation

### Transports

- **stdio**: For local Claude Desktop integration
- **http**: For remote/Docker deployment (exposes port 8080)

### Endpoints

- `:8080` - MCP server (HTTP transport)
- `:9090/livez` - Health check
- `:9090/metrics` - Prometheus metrics

---

## System Catalog

This repository is onboarded to the loveholidays system catalog.

**Status**: Onboarded
**Catalog location**: `.system-catalog/mcp-html-share.yaml`

To update or manage the system catalog, use the `lh-platform:system-catalog` skill from [lh-marketplace](https://github.com/loveholidays/lh-marketplace).
