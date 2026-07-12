package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/felixschelling/light_wikipedia_cli/internal/mcp"
	"github.com/felixschelling/light_wikipedia_cli/internal/tfmt"
	"github.com/felixschelling/light_wikipedia_cli/internal/tui"
	"github.com/felixschelling/light_wikipedia_cli/internal/wikipedia"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	search := flag.String("search", "", "search query (default: random article)")
	searchShort := flag.String("s", "", "search query (shorthand)")
	full := flag.Bool("full", false, "show full page content")
	fullShort := flag.Bool("f", false, "show full page content (shorthand)")
	lang := flag.String("lang", "en", "language code")
	langShort := flag.String("l", "en", "language code (shorthand)")
	random := flag.Bool("random", false, "force random article")
	randomShort := flag.Bool("r", false, "force random article (shorthand)")
	noColor := flag.Bool("no-color", false, "disable ANSI colors")
	tuiMode := flag.Bool("tui", false, "start interactive TUI")
	tuiShort := flag.Bool("t", false, "start interactive TUI (shorthand)")
	mcpMode := flag.Bool("mcp", false, "start MCP server (stdio transport)")
	help := flag.Bool("help", false, "show help")
	helpShort := flag.Bool("h", false, "show help (shorthand)")

	flag.Parse()

	if *help || *helpShort {
		printUsage()
		return
	}

	if os.Getenv("NO_COLOR") != "" {
		*noColor = true
	}

	if *mcpMode {
		runMCP()
		return
	}

	if *tuiMode || *tuiShort {
		client := buildClient("en")
		if err := tui.Run(client); err != nil {
			os.Exit(1)
		}
		return
	}

	query := *search
	if query == "" {
		query = *searchShort
	}

	langCode := *lang
	if *langShort != "en" {
		langCode = *langShort
	}

	client := buildClient(langCode)

	formatter := tfmt.New(*noColor)
	ctx := context.Background()

	shouldFull := *full || *fullShort
	shouldRandom := *random || *randomShort

	switch {
	case query != "" && shouldFull:
		results, err := client.Search(ctx, query, 1)
		if err != nil {
			formatter.Error(os.Stderr, fmt.Sprintf("search failed: %v", err))
			os.Exit(1)
		}
		if len(results) == 0 {
			formatter.Error(os.Stderr, "no results found")
			os.Exit(1)
		}
		page, err := client.GetPage(ctx, results[0].Title)
		if err != nil {
			formatter.Error(os.Stderr, fmt.Sprintf("failed to fetch page: %v", err))
			os.Exit(1)
		}
		formatter.Page(os.Stdout, page)

	case query != "":
		results, err := client.Search(ctx, query, 5)
		if err != nil {
			formatter.Error(os.Stderr, fmt.Sprintf("search failed: %v", err))
			os.Exit(1)
		}
		if len(results) == 0 {
			formatter.Error(os.Stderr, "no results found")
			os.Exit(1)
		}
		if len(results) > 1 {
			formatter.SearchResults(os.Stdout, results)
			fmt.Fprintln(os.Stdout)
			fmt.Fprintln(os.Stdout, "--- Showing first result ---")
			fmt.Fprintln(os.Stdout)
		}
		summary, err := client.GetSummary(ctx, results[0].Title)
		if err != nil {
			formatter.Error(os.Stderr, fmt.Sprintf("failed to fetch summary: %v", err))
			os.Exit(1)
		}
		formatter.Summary(os.Stdout, summary)

	case shouldRandom:
		summary, err := client.Random(ctx)
		if err != nil {
			formatter.Error(os.Stderr, fmt.Sprintf("failed to fetch random article: %v", err))
			os.Exit(1)
		}
		formatter.Summary(os.Stdout, summary)

	default:
		summary, err := client.Random(ctx)
		if err != nil {
			formatter.Error(os.Stderr, fmt.Sprintf("failed to fetch random article: %v", err))
			os.Exit(1)
		}
		formatter.Summary(os.Stdout, summary)
	}
}

func buildClient(langCode string) *wikipedia.Client {
	clientOpts := []wikipedia.ClientOption{
		wikipedia.WithLang(langCode),
		wikipedia.WithUserAgent("light_wikipedia_cli/1.0"),
	}
	if base := os.Getenv("WIKIPEDIA_API_BASE"); base != "" {
		clientOpts = append(clientOpts, wikipedia.WithBaseURL(base))
	}
	return wikipedia.New(clientOpts...)
}

func runMCP() {
	lang := "en"
	if l := os.Getenv("WIKIPEDIA_LANG"); l != "" {
		lang = l
	}

	clientOpts := []wikipedia.ClientOption{
		wikipedia.WithLang(lang),
		wikipedia.WithUserAgent("wikipedia-mcp/1.0"),
	}
	if base := os.Getenv("WIKIPEDIA_API_BASE"); base != "" {
		clientOpts = append(clientOpts, wikipedia.WithBaseURL(base))
	}

	client := wikipedia.New(clientOpts...)
	s := mcp.NewWikipediaServer(client)

	if err := server.ServeStdio(s); err != nil {
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`Wikipedia Terminal — fast Wikipedia reader

Usage:
  wikipedia [flags]

Flags:
  -s, --search string    Search query (default: random article)
  -f, --full             Show full page content
  -l, --lang string      Language code (default: "en")
  -r, --random           Force random article
  -t, --tui              Start interactive TUI
      --mcp              Start MCP server (stdio transport)
      --no-color         Disable ANSI colors
  -h, --help             Show help

Examples:
  wikipedia
  wikipedia --search "Go programming language"
  wikipedia -s "Go" -f
  wikipedia -l de --search "Go"
  wikipedia --random
  wikipedia --tui
  wikipedia --mcp`)
}
