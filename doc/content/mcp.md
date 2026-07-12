# MCP server

The `wikipedia-mcp` binary is an [MCP](https://modelcontextprotocol.io) server
that communicates over stdio, so an AI client spawns it as a subprocess. No HTTP
server, port, or auth is required.

You can also start the MCP server via the CLI binary: `wikipedia --mcp`.

## Configuration

Add it to your client's MCP server list (Claude Code, GitHub Copilot, Continue,
etc.):

```json
{
  "mcpServers": {
    "wikipedia": {
      "command": "/path/to/bin/wikipedia-mcp"
    }
  }
}
```

## Tools

- `wikipedia_search` — search articles (args: `query`, optional `limit`, `lang`)
- `wikipedia_get_page` — full article content (args: `title`, optional `lang`)
- `wikipedia_get_summary` — article summary (args: `title`, optional `lang`)
- `wikipedia_random` — random article (optional `lang`)

All tools accept an optional `lang` parameter (default `en`).

## Environment variables

| Variable | Effect |
|----------|--------|
| `WIKIPEDIA_API_BASE` | Override the Wikimedia REST API base URL |
| `WIKIPEDIA_LANG` | Default language for the server |
