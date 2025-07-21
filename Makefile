APP_NAME=mcp-html-share
GOFILES=$(shell find . -type f -name '*.go' -not -path "./vendor/*")

.PHONY: all build lint test docker clean tools check run run-http help

all: build

##@ Tools
tools: ## Installs required binaries locally
	GOTOOLCHAIN=go1.24.4 go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest

##@ Building
build: test ## Builds $(APP_NAME) go binary for local arch
	@echo "== build"
	CGO_ENABLED=0 go build -o bin/$(APP_NAME) ./cmd/$(APP_NAME)

##@ Testing
test: ## Run unit tests
	@echo "== test"
	go test -v -race -cover ./...

##@ Quality
check: tools ## Runs lint, fmt and vet checks against the codebase
	@echo "== check"
	golangci-lint run --timeout 5m
	go fmt ./...
	go vet ./...

tidy: ## Runs go mod tidy
	@echo "== tidy"
	go mod tidy

##@ Running
run: build ## Run the application with stdio transport (requires --bucket flag)
	@echo "== run (stdio)"
	@echo "Usage: make run BUCKET=your-bucket-name"
	@if [ -z "$(BUCKET)" ]; then echo "Error: BUCKET variable is required"; exit 1; fi
	./bin/$(APP_NAME) --bucket=$(BUCKET)

run-http: build ## Run the application with HTTP transport (requires --bucket flag)
	@echo "== run-http"
	@echo "Usage: make run-http BUCKET=your-bucket-name"
	@if [ -z "$(BUCKET)" ]; then echo "Error: BUCKET variable is required"; exit 1; fi
	./bin/$(APP_NAME) --bucket=$(BUCKET) --transport=http

##@ Docker
docker: ## Build Docker image
	@echo "== docker"
	docker build -t $(APP_NAME):latest .

##@ Cleanup
clean: ## Deletes binaries from the bin folder
	@echo "== clean"
	rm -rfv ./bin

##@ Help
help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)