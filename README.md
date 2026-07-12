<div align="center">

```
██╗     ██╗ ██████╗ ██╗  ██╗████████╗    ██╗    ██╗██╗██╗  ██╗██╗     ██████╗ ██╗     ██╗
██║     ██║██╔════╝ ██║  ██║╚══██╔══╝    ██║    ██║██║██║ ██╔╝██║    ██╔════╝ ██║     ██║
██║     ██║██║  ███╗███████║   ██║       ██║ █╗ ██║██║█████╔╝ ██║    ██║  ███╗██║     ██║
██║     ██║██║   ██║██╔══██║   ██║       ██║███╗██║██║██╔═██╗ ██║    ██║   ██║██║     ██║
███████╗██║╚██████╔╝██║  ██║   ██║       ╚███╔███╔╝██║██║  ██╗██║    ╚██████╔╝███████╗██║
╚══════╝╚═╝ ╚═════╝ ╚═╝  ╚═╝   ╚═╝        ╚══╝╚══╝ ╚═╝╚═╝  ╚═╝╚═╝    ╚═════╝ ╚══════╝╚═╝
```

# light_wikipedia_cli

[![Go Version](https://img.shields.io/badge/go-1.26.5-blue.svg)](https://go.dev/dl/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/felixscode/light_wikipedia_cli)](https://goreportcard.com/report/github.com/felixscode/light_wikipedia_cli)

**A fast, lightweight Wikipedia client for the terminal — and an MCP server so AI assistants can read Wikipedia too.**

[Usage](#usage) | [Quickstart](#quickstart) | [MCP Server](#mcp-server) | [Docs](https://felixscode.github.io/light_wikipedia_cli) | [Dev](#development)

</div>

Ships as two binaries sharing one internal Wikimedia REST API client:

- **`wikipedia`** — terminal CLI to search, read, and random-discover articles
- **`wikipedia-mcp`** — [MCP](https://modelcontextprotocol.io) server over stdio for AI clients

## Usage

```sh
wikipedia                              # random article summary
wikipedia -s "Go programming language" # search and read
wikipedia -s "Go" -f                   # full article
wikipedia -l de -s "Go"                # other languages
wikipedia --tui                        # interactive TUI
```

| Flag | Short | Description |
|------|-------|-------------|
| `--search` | `-s` | Search query (default: random) |
| `--full` | `-f` | Show full article |
| `--lang` | `-l` | Language code (default: `en`) |
| `--random` | `-r` | Force random article |
| `--tui` | `-t` | Interactive TUI |
| `--mcp` | | Run as MCP server |
| `--no-color` | | Disable ANSI colors |

## Quickstart

```sh
git clone https://github.com/felixscode/light_wikipedia_cli
cd light_wikipedia_cli
make build
```

Requires Go 1.26+. Builds `bin/wikipedia` and `bin/wikipedia-mcp`.

## MCP Server

Configure in your AI client (Claude Code, Copilot, Continue, etc.):

```json
{
  "mcpServers": {
    "wikipedia": { "command": "/path/to/bin/wikipedia-mcp" }
  }
}
```

**Tools:** `wikipedia_search` · `wikipedia_get_page` · `wikipedia_get_summary` · `wikipedia_random` — all accept optional `lang` (default `en`).

## Documentation

Full documentation is available at the [project site](https://felixscode.github.io/light_wikipedia_cli). Includes CLI reference, MCP setup, architecture notes, and a developer guide for AI agents.

## Development

```sh
make test       # all tests with race detector
make test-e2e   # e2e tests (local mock server)
make docs       # regenerate the documentation site
```

See [ARCH.md](./ARCH.md) for architecture and [AGENTS.md](./AGENTS.md) for the developer guide.

## License

MIT