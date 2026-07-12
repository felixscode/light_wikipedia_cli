# Project

`light_wikipedia_cli` is a fast, lightweight Wikipedia client written in Go. It ships
as two binaries that share a single internal API client:

- **`wikipedia`** — a terminal CLI to search, read, and discover random articles
- **`wikipedia-mcp`** — an [MCP](https://modelcontextprotocol.io) server so AI
  assistants can read Wikipedia over stdio

## Design

- **Stdlib `net/http`** for all Wikimedia REST API calls — no third-party HTTP client
- **`internal/` packages** keep the public surface limited to the two `cmd/` binaries
- **ANSI terminal formatting** with `NO_COLOR` support for the CLI
- **Zero config** — talks to the public Wikimedia REST API, no API key required

## Repository layout

```
light_wikipedia_cli/
├── cmd/
│   ├── wikipedia/        # CLI binary
│   ├── wikipedia-mcp/    # MCP server binary
│   └── docgen/           # docs static site generator
├── internal/
│   ├── wikipedia/        # shared REST API client
│   ├── tfmt/             # terminal formatting
│   ├── tui/              # interactive TUI
│   └── mcp/              # MCP tool definitions
├── doc/
│   ├── content/          # docs Markdown sources (committed)
│   └── public/           # generated HTML output (gitignored)
├── tests/                # end-to-end tests
├── Makefile
├── go.mod
├── AGENTS.md             # developer guide for AI agents
├── ARCH.md               # architecture notes
├── SKILL.md              # opencode setup skill
├── llms.txt              # LLM project overview
└── README.md
```

## Machine-readable docs

AI agents can read [`agents.txt`](/agents.txt) for a structured map of this
documentation site, or [`llms.txt`](/llms.txt) for a project-level overview.
