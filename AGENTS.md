# AGENTS.md — Dev Guide for AI Agents

## Project: `light_wikipedia_cli`

Fast Wikipedia CLI + MCP server written in Go. Two binaries: `wikipedia` (terminal) and `wikipedia-mcp` (MCP server over stdio).

---

## Code Layout

```
light_wikipedia_cli/
├── cmd/
│   ├── wikipedia/             # CLI — flag parsing + main flow
│   ├── wikipedia-mcp/         # MCP server — stdio entry point
│   └── docgen/                # docs static site generator
├── internal/
│   ├── wikipedia/             # shared API client (all HTTP to Wikimedia API)
│   ├── tfmt/                  # terminal formatting (ANSI colors + word-wrap)
│   ├── tui/                   # interactive TUI (bubbletea)
│   └── mcp/                   # MCP tool definitions + server factory
├── doc/
│   ├── content/               # docs Markdown sources (committed)
│   └── public/                # generated HTML output (gitignored)
├── tests/                     # end-to-end tests
├── Makefile
├── go.mod
├── AGENTS.md                  # this file
├── ARCH.md                    # architecture notes
├── SKILL.md                   # opencode setup skill
├── llms.txt                   # LLM project overview
└── README.md
```

---

## Development Commands

| Command | What it does |
|---------|-------------|
| `make build` | Build both binaries to `./bin/` |
| `make build-wikipedia` | Build only the CLI binary |
| `make build-mcp` | Build only the MCP server binary |
| `make test` | Run all tests with race detector |
| `make test-e2e` | Run end-to-end tests (CLI + MCP against a local mock server) |
| `make test-cover` | Run tests with coverage |
| `make clean` | Remove build artifacts |
| `make run` | Build and run `wikipedia` CLI with `--random` |
| `go vet ./...` | Static analysis |
| `go fmt ./...` | Format all Go files |
| `go mod tidy` | Clean up dependencies |

---

## Testing Conventions

- Unit tests use `net/http/httptest` to mock the Wikimedia API — never hit real endpoints
- Table-driven tests preferred
- File: `internal/wikipedia/client_test.go`
- End-to-end tests live in `tests/` behind the `e2e` build tag (`go test -tags e2e ./tests/...`).
  They build both binaries, start a local mock Wikipedia REST server, and exercise every CLI flag path and MCP tool call against it. No network access required.

## Environment Variables

| Variable | Effect |
|----------|--------|
| `WIKIPEDIA_API_BASE` | Override the Wikimedia REST API base URL (used by e2e tests and for self-hosting/proxies) |
| `WIKIPEDIA_LANG` | Default language for the MCP server |
| `NO_COLOR` | Disable ANSI colors in CLI output |

## Coding Conventions

- Stdlib `net/http` for API calls (no third-party HTTP clients)
- All exports documented with Go doc comments
- Errors wrapped with `fmt.Errorf("context: %w", err)`
- `context.Context` as first arg in all public API methods
- Respect `NO_COLOR` env var in any output logic

## MCP Server Details

Uses `github.com/mark3labs/mcp-go`. Tools are defined in `internal/mcp/server.go`:

- `wikipedia_search` — search articles
- `wikipedia_get_page` — full article content
- `wikipedia_get_summary` — article summary
- `wikipedia_random` — random article

All tools accept optional `lang` parameter (default `en`). MCP transport is stdio.

## Wikimedia API Details

The client builds per-language host URLs:

```
https://{lang}.wikipedia.org/api/rest_v1/{endpoint}
https://{lang}.wikipedia.org/w/api.php?{params}
```

Endpoints used:
- `GET /api/rest_v1/page/summary/{title}` — article summary
- `GET /api/rest_v1/page/html/{title}` — full article HTML
- `GET /api/rest_v1/page/random/summary` — random article
- `GET /w/api.php?action=query&list=search&srsearch={query}&format=json&srlimit={limit}` — search (Action API)

Override with `WIKIPEDIA_API_BASE` env var to set a custom base host.