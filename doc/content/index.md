# light_wikipedia_cli

A fast, lightweight Wikipedia client for the terminal — and an MCP server so AI
assistants can read Wikipedia too.

This site documents how to install, use, and develop the project.

## Highlights

- **Two binaries**: a terminal CLI (`wikipedia`) and an MCP server (`wikipedia-mcp`)
- **Zero-config**: talks to the public Wikimedia REST API, no API key
- **ANSI output** that respects `NO_COLOR`
- **Self-contained**: stdlib `net/http` for the API client

## Sections

- [Project](project.html) — what this project is and how it's structured
- [Getting started](getting-started.html) — install and first run
- [CLI usage](cli.html) — flags and examples
- [MCP server](mcp.html) — wire it into your AI client

## For agents

A machine-readable map of this site (pages + the `setup-wiki` skill) is available
at [`agents.txt`](agents.txt). A project-level overview for LLMs is at
[`llms.txt`](llms.txt).
