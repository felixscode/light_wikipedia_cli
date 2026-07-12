<div align="center">

```
██╗     ██╗ ██████╗ ██╗  ██╗████████╗    ██╗    ██╗██╗██╗  ██╗██╗      ██████╗ ██╗     ██╗
██║     ██║██╔════╝ ██║  ██║╚══██╔══╝    ██║    ██║██║██║ ██╔╝██║    ██╔════╝ ██║     ██║
██║     ██║██║  ███╗███████║   ██║       ██║ █╗ ██║██║█████╔╝ ██║    ██║      ██║     ██║
██║     ██║██║   ██║██╔══██║   ██║       ██║███╗██║██║██╔═██╗ ██║    ██║      ██║     ██║
███████╗██║╚██████╔╝██║  ██║   ██║       ╚███╔███╔╝██║██║  ██╗██║    ╚██████╗ ███████╗██║
╚══════╝╚═╝ ╚═════╝ ╚═╝  ╚═╝   ╚═╝        ╚══╝╚══╝ ╚═╝╚═╝  ╚═╝╚═╝     ╚═════╝ ╚══════╝╚═╝
```

[![Go Version](https://img.shields.io/badge/go-1.26.5-blue.svg)](https://go.dev/dl/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![golangci-lint](https://github.com/felixscode/light_wikipedia_cli/actions/workflows/golangci-lint.yml/badge.svg)](https://github.com/felixscode/light_wikipedia_cli/actions/workflows/golangci-lint.yml)
[![Docs](https://img.shields.io/badge/docs-github_pages-blue.svg)](https://felixscode.github.io/light_wikipedia_cli)

**A fast, lightweight Wikipedia client for the terminal — and an MCP server so AI assistants can read Wikipedia too.**

[Usage](#usage) | [Quickstart](#quickstart) | [MCP Server](#mcp-server) | [Docs](https://felixscode.github.io/light_wikipedia_cli) | [Dev](#development)

</div>

Ships as two binaries sharing one internal Wikimedia REST API client:

- **`wikipedia`** — terminal CLI to search, read, and random-discover articles
- **`wikipedia-mcp`** — [MCP](https://modelcontextprotocol.io) server over stdio for AI clients

## About

I started this repository in my early programming career. I wrote the CLI in pure Python using functional programming.
I mainly used it with my bashrc to render a random Wikipedia article on terminal startup. Took me roughly a week to build after work.

I stumbled across this project in 2026, and asked myself: can I rebuild this into a lightweight LLM-ready CLI in one prompt as cheap as possible?
I used DeepSeek V4 Flash via OpenRouter on Opencode. First shot I got a working CLI and MCP server. I added a TUI in 2 more prompts. And this wiki on gh-pages in 3 more prompts. I landed the entire rewrite in 1.5h at $1.10 API cost.

## Quickstart

Download the latest release and make binaries executable:

```sh
chmod +x wikipedia
wikipedia -r   # gets you a random article summary
```

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


Requires Go 1.26+ to build. Builds `bin/wikipedia` and `bin/wikipedia-mcp`.

## MCP Server

Configure in your AI client (Claude Code, Copilot, Continue, etc.):

```json
{
  "mcpServers": {
    "wikipedia": { "command": "/path/to/bin/wikipedia-mcp" }
  }
}
```

## Documentation

Documentation is available at the [project site](https://felixscode.github.io/light_wikipedia_cli). Includes CLI reference, MCP setup, architecture notes, and a developer guide for AI agents.


## License

MIT
