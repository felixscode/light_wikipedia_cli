package main

import (
	"os"

	"github.com/felixschelling/light_wikipedia_cli/internal/mcp"
	"github.com/felixschelling/light_wikipedia_cli/internal/wikipedia"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
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
