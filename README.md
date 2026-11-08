# MCP HTML Share

A Go-based [MCP (Model Context Protocol)](https://modelcontextprotocol.io/) server that shares HTML content via Google Cloud Storage. Upload HTML and get back a public URL.

## Usage Example

An LLM can use this tool to create and share web content:

**User**: "Create a data visualization showing monthly sales trends and share the URL"

**Assistant**: "I'll create an interactive chart and upload it for you"
```json
{
  "tool": "share-html",
  "arguments": {
    "html_content": "<html>... chart content ...</html>",
    "short_name": "sales-trends-2024"
  }
}
```
**Result**: `https://storage.cloud.google.com/your-bucket/sales-trends-2024-a1b2c3d4.html`

## Quick Start

### Using Docker

```bash
docker run -p 8080:8080 -p 9090:9090 \
  -e GOOGLE_APPLICATION_CREDENTIALS=/creds/credentials.json \
  -v /path/to/credentials.json:/creds/credentials.json \
  ghcr.io/loveholidays/mcp-html-share:latest \
  --bucket=your-bucket-name
```

### Using Pre-built Binary

Download from [releases](https://github.com/loveholidays/mcp-html-share/releases) or install via Homebrew:

```bash
brew tap loveholidays/tap
brew install mcp-html-share
```

Run the server:
```bash
mcp-html-share --bucket=your-bucket-name --transport=http
```

## Configuration

### Required Setup

1. **Create a GCS bucket**:
   ```bash
   gcloud storage buckets create gs://your-bucket-name --location=US
   ```

2. **For public URLs** (default):
   ```bash
   gcloud storage buckets add-iam-policy-binding gs://your-bucket-name \
     --member=allUsers \
     --role=roles/storage.objectViewer
   ```

3. **Set up authentication** (choose one):
   - Service Account: `export GOOGLE_APPLICATION_CREDENTIALS=/path/to/credentials.json`
   - Application Default: `gcloud auth application-default login`

### Command Line Flags

- `--bucket` (required): GCS bucket name
- `--transport`: `http` (default in Docker) or `stdio` (for Claude Desktop)
- `--http-port`: HTTP port (default: "8080")
- `--health-port`: Health/metrics port (default: "9090")
- `--public-url`: Return public URLs (default: true) or signed URLs (false)
- `--sentry-dsn`: Optional Sentry DSN. Error reporting is disabled when omitted. Defaults to `SENTRY_DSN`.
- `--sentry-environment`: Optional Sentry environment name. Defaults to `SENTRY_ENVIRONMENT`.

### Optional error reporting

Sentry reporting is opt-in and disabled unless `--sentry-dsn` is provided. Events are reduced to an event ID, error level, and fixed `application error` message before sending. The server does not attach request URLs, request bodies, users, breadcrumbs, tags, local context, exception details, HTML content, short names, uploaded URLs, or local environment data to Sentry events.

## MCP Tools

### share-html

Upload HTML content and get a public URL.

**Request:**
```json
{
  "tool": "share-html",
  "arguments": {
    "html_content": "<html><body><h1>Hello World</h1></body></html>",
    "short_name": "hello-world"
  }
}
```

**Response:**
```json
{
  "url": "https://storage.cloud.google.com/your-bucket/hello-world-12345678.html"
}
```

## Using with Claude Desktop

Add to your Claude Desktop configuration:

```json
{
  "mcpServers": {
    "html-share": {
      "command": "npx",
      "args": [
        "mcp-remote",
        "https://your-mcp-html-share-server.example.com"
      ]
    }
  }
}
```

For local development with stdio:
```json
{
  "mcpServers": {
    "html-share": {
      "command": "mcp-html-share",
      "args": ["--bucket", "your-bucket-name", "--transport", "stdio"]
    }
  }
}
```

## Endpoints

- **MCP Server**: `http://localhost:8080` (when using `--transport=http`)
- **Health**: `GET http://localhost:9090/livez`
- **Metrics**: `GET http://localhost:9090/metrics` (Prometheus format)

## Development

```bash
# Clone and build
git clone https://github.com/loveholidays/mcp-html-share
cd mcp-html-share
make build

# Run tests
make test

# Run locally
make run-http BUCKET=your-bucket-name
```

### CodeScene coverage

Run `make coverage` with Go and Python 3 installed to generate `coverage.out`
and `coverage.info`. The report includes every production Go file, including
packages without tests. Report validation fails if a file is omitted or source
paths do not match this repository.

The CodeScene coverage workflow uses the public Go-to-LCOV converter and a
checksum-pinned CodeScene CLI directly. It checks project `66419` on pull
requests and uploads full reports from `main` for scheduled CodeScene analysis.
Python verifier tests also produce `tooling.info` so the coverage helpers have
measured coverage themselves.
The aggregate requires both the current CLI policy verdict and the native
`CodeScene Code Coverage (main)` check to pass. Existing code-health checks,
coverage thresholds and repository merge rules are unchanged.

This public repository needs a repository-scoped `CS_TOKEN` Actions secret
approved for CodeScene coverage checks and uploads. Private-only organization
secrets are unavailable here. Missing credentials fail the coverage workflow,
including secret-restricted pull requests; ordinary CI tests run independently.
An accepted main upload becomes visible only after a CodeScene analysis consumes
it, which must be verified after merge.

## License

Coverage onboarding documentation-only integration trial.

LGPL-3.0 - see [LICENSE](LICENSE) file.
