# Getting started

## Requirements

- Go 1.26 or newer
- Dependencies: `bubbletea`, `lipgloss`, `bubbles`, `goldmark`, `mcp-go`, `golang.org/x/net/html`

## Build

```sh
git clone https://github.com/felixschelling/light_wikipedia_cli
cd light_wikipedia_cli
make build
```

This produces two binaries under `bin/`:

- `bin/wikipedia` — terminal CLI
- `bin/wikipedia-mcp` — MCP server

Add `bin/` to your `PATH`, or invoke the binaries directly.

## Quick start

```sh
# Random article summary
wikipedia

# Search and read
wikipedia --search "Go programming language"

# Interactive TUI
wikipedia --tui
```

## Building this documentation site

The docs you are reading are generated from Markdown in `doc/content` by the
`docgen` tool:

```sh
make docs
```

Output is written to `doc/public` and can be served by any static file server:

```sh
python3 -m http.server -d doc/public 8000
```
