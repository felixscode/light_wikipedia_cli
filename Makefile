BIN_DIR := bin
BIN_CLI := $(BIN_DIR)/wikipedia
BIN_MCP := $(BIN_DIR)/wikipedia-mcp

.PHONY: all build build-wikipedia build-mcp docs serve test test-e2e test-cover clean run

all: build

build: build-wikipedia build-mcp

docs:
	@mkdir -p $(BIN_DIR)
	go run ./cmd/docgen -content doc/content -out doc/public -site "light_wikipedia_cli docs" -skill SKILL.md -llms doc/content/llms.txt

serve:
	python3 -m http.server -d doc/public 8000

build-wikipedia:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_CLI) ./cmd/wikipedia

build-mcp:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_MCP) ./cmd/wikipedia-mcp

test:
	go test -race ./...

test-e2e:
	go test -tags e2e -v ./tests/...

test-cover:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

clean:
	rm -rf $(BIN_DIR) coverage.out coverage.html

run: build-wikipedia
	$(BIN_CLI) --random