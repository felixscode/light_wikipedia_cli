# CLI usage

The `wikipedia` binary reads and searches Wikipedia from the terminal.

```sh
# Random article summary (default)
wikipedia

# Search and show the top result's summary
wikipedia --search "Go programming language"
wikipedia -s "Go"

# Full article text
wikipedia -s "Go" --full
wikipedia -s "Go" -f

# Other languages
wikipedia -l de --search "Go"

# Explicitly force a random article
wikipedia --random

# Interactive TUI
wikipedia --tui

# Run as MCP server
wikipedia --mcp
```

## Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--search` | `-s` | Search query (omit for a random article) | — |
| `--full` | `-f` | Show full page content instead of the summary | `false` |
| `--lang` | `-l` | Wikipedia language code | `en` |
| `--random` | `-r` | Force a random article | `false` |
| `--tui` | `-t` | Start interactive TUI | `false` |
| `--mcp` | | Run as MCP server (stdio transport) | `false` |
| `--no-color` | | Disable ANSI colors | `false` |
| `--help` | `-h` | Show help | |

ANSI colors are also disabled when the `NO_COLOR` environment variable is set.
