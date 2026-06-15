package mcp

import (
	"context"
	"fmt"
	"strings"

	mcpgo "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rafael/lexis/internal/library"
)

const citationInstruction = `Use ONLY the passages below to answer. For every claim, include an inline citation [N].
Do not use outside knowledge. If the passages don't contain the answer, say so.

`

// Serve starts the MCP stdio server exposing search_books via lib.
func Serve(lib *library.Library, topK int, citationEnforcement bool) error {
	s := server.NewMCPServer("lexis", "1.0.0",
		server.WithToolCapabilities(false),
	)

	tool := mcpgo.NewTool("search_books",
		mcpgo.WithDescription("Search the personal book and document collection for relevant passages. Use this whenever the user asks about a topic that may be covered in their documents."),
		mcpgo.WithString("query",
			mcpgo.Required(),
			mcpgo.Description("Search query to find relevant passages"),
		),
	)

	s.AddTool(tool, func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		query, err := req.RequireString("query")
		if err != nil {
			return mcpgo.NewToolResultError("query parameter is required"), nil
		}

		results, err := lib.Search(ctx, query, topK)
		if err != nil {
			return mcpgo.NewToolResultError(fmt.Sprintf("search failed: %v", err)), nil
		}
		if len(results) == 0 {
			return mcpgo.NewToolResultText("No relevant passages found."), nil
		}

		var sb strings.Builder
		if citationEnforcement {
			sb.WriteString(citationInstruction)
		}
		for i, r := range results {
			fmt.Fprintf(&sb, "[%d] Source: %s (similarity: %.2f)\nText: %s\n",
				i+1, r.Metadata["source"], r.Score, r.Text)
			if i < len(results)-1 {
				sb.WriteString("\n---\n\n")
			}
		}

		return mcpgo.NewToolResultText(sb.String()), nil
	})

	return server.ServeStdio(s)
}
