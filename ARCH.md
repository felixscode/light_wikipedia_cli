# Architecture — light_wikipedia_cli

## Overview

A fast, lightweight Wikipedia client shipping as two binaries:

1. **`wikipedia`** — terminal CLI (search, read, random articles)
2. **`wikipedia-mcp`** — [Model Context Protocol](https://modelcontextprotocol.io) server over stdio

Both share the same internal API client and differ only in their entry point and output formatting.

---

## Directory Layout

```
light_wikipedia_cli/
├── cmd/
│   ├── wikipedia/                 # CLI binary entry point
│   │   └── main.go                # flag parsing -> fetch -> format -> print
│   └── wikipedia-mcp/             # MCP server binary entry point
│       └── main.go                # server setup, tool registration, stdio listen
├── internal/
│   ├── wikipedia/                 # Wikipedia REST API client (shared core)
│   │   ├── client.go              # Client struct, Search/GetPage/GetSummary/Random
│   │   ├── client_test.go         # unit tests with recorded fixtures
│   │   ├── types.go               # Page, SearchResult, Summary structs
│   │   └── errors.go              # typed errors (NotFound, APIError, etc.)
│   ├── tfmt/                      # terminal formatting
│   │   └── fmt.go                 # colorized headings, word-wrap, no-color mode
│   └── mcp/                       # MCP protocol glue
│       └── server.go              # tool definitions, handler functions, server factory
├── tests/                         # e2e tests (//go:build e2e) — local mock server
│   ├── e2e_test.go                # CLI flag-path coverage
│   └── mcp_e2e_test.go            # MCP handshake + tool calls
├── go.mod
├── go.sum
├── Makefile
├── AGENTS.md                      # dev guide for AI agents
├── ARCH.md                        # this file
└── README.md
```

---

## Component Architecture

```
┌────────────────────────────────────────────────────┐
│                   Entry Points                      │
│                                                     │
│  cmd/wikipedia/main.go    cmd/wikipedia-mcp/main.go │
│        (flag parse)           (stdio transport)     │
└──────────┬──────────────────────────┬───────────────┘
           │                          │
           ▼                          ▼
┌──────────────────────┐   ┌──────────────────────┐
│  internal/tfmt/fmt.go  │   │ internal/mcp/server.go│
│  ANSI terminal fmt   │   │ MCP tool definitions │
│  (CLI only)          │   │ (Tool/Resource/Prompt)│
└──────────┬───────────┘   └──────────┬────────────┘
           │                          │
           └──────────┬───────────────┘
                      ▼
┌─────────────────────────────────────────────┐
│         internal/wikipedia/client.go         │
│  HTTP client -> api.wikimedia.org (REST)     │
│                                              │
│  Search(ctx, q, limit)  → []SearchResult     │
│  GetPage(ctx, title)    → Page               │
│  GetSummary(ctx, title) → Summary            │
│  Random(ctx)            → Page               │
└──────────────────┬──────────────────────────┘
                   │
                   ▼
    https://api.wikimedia.org/core/v1/wikipedia/{lang}/...
```

### Data flow

1. Entry point receives user intent (CLI flags or MCP tool call)
2. Calls the appropriate `internal/wikipedia` method
3. Response goes through formatter:
   - **CLI path**: `internal/fmt` renders ANSI-colored output to stdout
   - **MCP path**: raw struct serialized to JSON-RPC response (AI client handles display)
4. Result returned to user

---

## Task-Level Dev Ops Plan

### Phase 1: Scaffold ✅
- [x] `go mod init`
- [x] directory structure
- [x] add dependencies: `mcp-go`, `x/term`
- [x] `.gitignore` (inherited from old project)
- **Output**: compilable empty module

### Phase 2: API Client (`internal/wikipedia/`)
- [ ] `types.go` — `Page`, `SearchResult`, `Summary`, `SearchResponse`, `PageResponse`
- [ ] `errors.go` — `ErrNotFound`, `ErrAPI`, `ErrInvalidTitle`
- [ ] `client.go` — `Client` struct with:
  - Constructor `New(lang string, userAgent string)`
  - `Search(ctx, query, limit)` — `GET /core/v1/wikipedia/{lang}/search/page`
  - `GetPage(ctx, title)` — `GET /core/v1/wikipedia/{lang}/page/{title}/bare`
  - `GetSummary(ctx, title)` — `GET /core/v1/wikipedia/{lang}/page/{title}/summary`
  - `Random(ctx)` — `GET /core/v1/wikipedia/{lang}/page/random/summary`
- [ ] `client_test.go` — table-driven tests with HTTP test server

### Phase 3: Terminal Formatter (`internal/tfmt/`)
- [ ] `fmt.go` — public API:
  - `Page(w io.Writer, p wikipedia.Page, full bool)` — prints title, optional full content
  - `SearchResults(w io.Writer, results []wikipedia.SearchResult)` — numbered list
  - `Summary(w io.Writer, s wikipedia.Summary)` — title + extract
- [ ] Word-wrap to terminal width via `x/term`
- [ ] ANSI color constants (cyan headings, dim separators)
- [ ] Respect `NO_COLOR` env var / `--no-color` flag

### Phase 4: CLI Binary (`cmd/wikipedia/`)
- [ ] Flags:
  - `--search, -s` — search query (default: empty → random article)
  - `--full, -f` — show full content (default: summary only)
  - `--lang, -l` — language code (default: `en`)
  - `--random, -r` — force random article
  - `--no-color` — disable ANSI
- [ ] Default behavior: random article summary

### Phase 5: MCP Server (`internal/mcp/`)
- [ ] `server.go` — factory function `NewWikipediaServer(client *wikipedia.Client) *mcp.Server`
- [ ] Tools registered:
  - `wikipedia_search` — args: `query` (string, req), `limit` (int, opt), `lang` (str, opt)
  - `wikipedia_get_page` — args: `title` (string, req), `lang` (str, opt)
  - `wikipedia_get_summary` — args: `title` (string, req), `lang` (str, opt)
  - `wikipedia_random` — args: `lang` (str, opt)

### Phase 6: MCP Binary (`cmd/wikipedia-mcp/`)
- [ ] Create client, create server, register tools, `ServeStdio()`

### Phase 7: Polish & Docs
- [ ] Makefile targets: `build`, `build-all`, `test`, `clean`
- [ ] Cross-compile for linux/darwin amd64/arm64
- [ ] README.md with install instructions and examples
- [ ] AGENTS.md for AI-assisted development

---

## Dependencies

| Dependency | Why | Alternatives Considered |
|-----------|-----|------------------------|
| stdlib `net/http` + `encoding/json` | API client | — |
| `golang.org/x/term` | Terminal width detection | `term.GetSize` in stdlib |
| `github.com/mark3labs/mcp-go` | MCP protocol (JSON-RPC 2.0 over stdio) | Raw JSON-RPC impl (more code) |

---

## Design Decisions

### Why `api.wikimedia.org` instead of `en.wikipedia.org/w/api.php`?
The REST API returns cleaner JSON, uses standard HTTP semantics, and doesn't require `action=parse` hackery. Response schemas are versioned and documented.

### Why `internal/` packages?
Prevents external imports. The public surface is just the two `cmd/` binaries.

### Why ANSI-only formatting?
Zero dependency terminal styling. Wikipedia articles are plain text — we don't need markdown rendering or HTML parsing. Colorized headings + word-wrap is the sweet spot for readability without bloat.

### Why stdio transport for MCP?
Local MCP servers communicate over stdin/stdout — no HTTP server, no port conflicts, no auth. The AI client (Claude Code, Copilot, Continue) spawns the binary as a subprocess.

---

## Testing Strategy

### Unit tests (`internal/wikipedia/client_test.go`)
- Table-driven, `net/http/httptest` mock server per test
- Cover: search, summary, page, random, error paths, limit clamping, HTML stripping
- No network access

### End-to-end tests (`tests/`, `//go:build e2e`)
- Build both binaries once in `TestMain`
- Start a **local mock** Wikipedia REST server (not the real API — avoids rate limits and flakiness)
- Point binaries at the mock via `WIKIPEDIA_API_BASE` env var
- CLI e2e: every flag path (`--search`, `-s`, `-f`, `-l`, `--random`, `--no-color`, `--help`, not-found)
- MCP e2e: `initialize` handshake, `tools/list`, `tools/call` for search + summary
- Run with: `make test-e2e` or `go test -tags e2e ./tests/...`

### Configurability
- `WIKIPEDIA_API_BASE` — base URL override (also usable for self-hosting a Wikimedia mirror)
- `WIKIPEDIA_LANG` — default language for the MCP server
- `NO_COLOR` — disable ANSI in CLI output