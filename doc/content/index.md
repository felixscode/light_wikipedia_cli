# light_wikipedia_cli

A fast, lightweight Wikipedia client for the terminal — and an MCP server so AI
assistants can read Wikipedia too.

This site documents how to install, use, and develop the project.

## About

I started this repository in my early programming career. I wrote the CLI in pure Python using functional programming.
I mainly used it with my bashrc to render a random Wikipedia article on terminal startup. Took me roughly a week to build after work.

I stumbled across this project in 2026, and asked myself: can I rebuild this into a lightweight LLM-ready CLI in one prompt as cheap as possible?
I used DeepSeek V4 Flash via OpenRouter on Opencode. First shot I got a working CLI and MCP server. I added a TUI in 2 more prompts. And this wiki on gh-pages in 3 more prompts. I landed the entire rewrite in 1.5h at $1.10 API cost.

## Sections

- [Project](project.html) — what this project is and why it exists
- [Getting started](getting-started.html) — install and first run
- [CLI usage](cli.html) — flags and examples
- [MCP server](mcp.html) — wire it into your AI client

## For agents

A machine-readable map of this site (pages + the `setup-wiki` skill) is available
at [`agents.txt`](agents.txt). A project-level overview for LLMs is at
[`llms.txt`](llms.txt).
