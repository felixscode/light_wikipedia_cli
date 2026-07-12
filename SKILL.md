---
name: light-wikipedia-cli-setup
description: Use when setting up, building, documenting, or running the light_wikipedia_cli project — install deps, build the CLI/MCP binaries, generate the static docs site, and invoke the CLI or MCP server.
---

# light_wikipedia_cli — setup & usage

A fast Wikipedia CLI + MCP server written in Go. Two binaries: `wikipedia`
(terminal) and `wikipedia-mcp` (MCP server over stdio). Both share an internal
REST API client for the Wikimedia REST API.

## Prerequisites

- Go 1.26 or newer

## Setup

```sh
git clone https://github.com/felixscode/light_wikipedia_cli
cd light_wikipedia_cli
make build          # builds bin/wikipedia and bin/wikipedia-mcp
```

The `bin/` directory is gitignored. Add it to your `PATH` or run binaries directly.

## Usage — CLI

```sh
wikipedia                                  # random article summary
wikipedia --search "Go programming language"
wikipedia -s "Go" -f                       # full article
wikipedia -l de --search "Go"              # language
wikipedia --random
```

Flags: `-s/--search`, `-f/--full`, `-l/--lang` (default `en`),
`-r/--random`, `--no-color`, `-h/--help`. ANSI colors are disabled when the
`NO_COLOR` env var is set.

## Usage — MCP server

Configure the MCP server in your AI client (Claude Code, Copilot, Continue, …):

```json
{
  "mcpServers": {
    "wikipedia": { "command": "/path/to/bin/wikipedia-mcp" }
  }
}
```

Tools: `wikipedia_search`, `wikipedia_get_summary`, `wikipedia_get_page`,
`wikipedia_random`. All accept optional `lang` (default `en`). Env vars:
`WIKIPEDIA_API_BASE` (override API base), `WIKIPEDIA_LANG` (default language).

## Documentation site

Docs live as Markdown in `doc/content/` and are built into a static site by
`cmd/docgen`:

```sh
make docs     # generates doc/public/*.html + doc/public/agents.txt
make serve    # serves doc/public/ at http://localhost:8000
```

`doc/public/` is gitignored; only `doc/content/` is committed. A machine-readable
map of the site is at `doc/public/agents.txt`.

## Conventions

- Stdlib `net/http` for API calls (no third-party HTTP client)
- Public surface limited to the two `cmd/` binaries; shared code in `internal/`
- `context.Context` as the first arg of all public API methods
- Errors wrapped with `fmt.Errorf("context: %w", err)`
- Respect `NO_COLOR` in any output logic
- Tests use `net/http/httptest` (no real network); e2e tests behind the `e2e` build tag
