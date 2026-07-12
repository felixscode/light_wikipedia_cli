package mcp

import (
	"context"
	"fmt"

	"github.com/felixschelling/light_wikipedia_cli/internal/wikipedia"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func NewWikipediaServer(client *wikipedia.Client) *server.MCPServer {
	s := server.NewMCPServer(
		"wikipedia-mcp",
		"1.0.0",
		server.WithLogging(),
	)

	searchTool := mcp.NewTool("wikipedia_search",
		mcp.WithDescription("Search Wikipedia articles"),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("Search query"),
		),
		mcp.WithNumber("limit",
			mcp.Description("Maximum number of results (1-50)"),
		),
		mcp.WithString("lang",
			mcp.Description("Language code (default: en)"),
		),
	)

	pageTool := mcp.NewTool("wikipedia_get_page",
		mcp.WithDescription("Get full content of a Wikipedia article"),
		mcp.WithString("title",
			mcp.Required(),
			mcp.Description("Article title"),
		),
		mcp.WithString("lang",
			mcp.Description("Language code (default: en)"),
		),
	)

	summaryTool := mcp.NewTool("wikipedia_get_summary",
		mcp.WithDescription("Get a summary of a Wikipedia article"),
		mcp.WithString("title",
			mcp.Required(),
			mcp.Description("Article title"),
		),
		mcp.WithString("lang",
			mcp.Description("Language code (default: en)"),
		),
	)

	randomTool := mcp.NewTool("wikipedia_random",
		mcp.WithDescription("Get a random Wikipedia article summary"),
		mcp.WithString("lang",
			mcp.Description("Language code (default: en)"),
		),
	)

	s.AddTool(searchTool, handleSearch(client))
	s.AddTool(pageTool, handleGetPage(client))
	s.AddTool(summaryTool, handleGetSummary(client))
	s.AddTool(randomTool, handleRandom(client))

	return s
}

func handleSearch(client *wikipedia.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()

		query, _ := args["query"].(string)
		if query == "" {
			return mcp.NewToolResultError("query is required"), nil
		}

		limit := 10
		if l, ok := args["limit"].(float64); ok {
			limit = int(l)
		}

		lang, _ := args["lang"].(string)
		if lang != "" {
			client.SetLang(lang)
		}

		results, err := client.Search(ctx, query, limit)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("search failed: %v", err)), nil
		}

		var text string
		for _, r := range results {
			text += fmt.Sprintf("- **%s**", r.Title)
			if r.Snippet != "" {
				text += fmt.Sprintf(": %s", r.Snippet)
			}
			text += "\n"
		}
		if text == "" {
			text = "No results found."
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{Text: text},
			},
		}, nil
	}
}

func handleGetPage(client *wikipedia.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()

		title, _ := args["title"].(string)
		if title == "" {
			return mcp.NewToolResultError("title is required"), nil
		}

		lang, _ := args["lang"].(string)
		if lang != "" {
			client.SetLang(lang)
		}

		page, err := client.GetPage(ctx, title)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to fetch page: %v", err)), nil
		}

		content := wikipedia.StripHTML(page.Content)
		if content == "" {
			content = "(Full content not available as plain text)"
		}

		text := fmt.Sprintf("# %s\n\n%s\n\n---\n%s", page.Title, content, page.URL)

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{Text: text},
			},
		}, nil
	}
}

func handleGetSummary(client *wikipedia.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()

		title, _ := args["title"].(string)
		if title == "" {
			return mcp.NewToolResultError("title is required"), nil
		}

		lang, _ := args["lang"].(string)
		if lang != "" {
			client.SetLang(lang)
		}

		summary, err := client.GetSummary(ctx, title)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to fetch summary: %v", err)), nil
		}

		text := fmt.Sprintf("# %s\n\n%s\n\n---\n%s", summary.Title, summary.Extract, summary.URL)

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{Text: text},
			},
		}, nil
	}
}

func handleRandom(client *wikipedia.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()

		lang, _ := args["lang"].(string)
		if lang != "" {
			client.SetLang(lang)
		}

		summary, err := client.Random(ctx)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to fetch random article: %v", err)), nil
		}

		text := fmt.Sprintf("# %s\n\n%s\n\n---\n%s", summary.Title, summary.Extract, summary.URL)

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{Text: text},
			},
		}, nil
	}
}
