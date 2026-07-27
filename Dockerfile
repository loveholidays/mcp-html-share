# Build stage
FROM golang:1.25.0 as builder

WORKDIR /src

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build only (no tests/checks - they run in CI)
RUN CGO_ENABLED=0 go build -o bin/mcp-html-share ./cmd/mcp-html-share

# Final stage
FROM scratch

COPY --from=builder /src/bin/mcp-html-share /app/mcp-html-share
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

WORKDIR /app
EXPOSE 8080 9090

ENTRYPOINT ["./mcp-html-share", "--transport", "http"]
