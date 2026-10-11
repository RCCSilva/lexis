package mcp

import (
	"context"
	"fmt"
	"strings"

	"github.com/RCCSilva/lexis/internal/library"
	mcpgo "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const citationInstruction = `Use ONLY the passages below to answer. For every claim, include an inline citation [N] and name the source and the pages or section shown in the passage header.
Do not use outside knowledge. If the passages don't contain the answer, say so.
If a passage looks cut off, call get_chunk with its source and chunk number (neighbors=1) to read the surrounding text.

`

const (
	maxTopK      = 20
	maxNeighbors = 3
)

// Serve starts the MCP stdio server exposing search_documents, get_chunk and
// list_sources via lib.
func Serve(lib *library.Library, topK int, citationEnforcement bool) error {
	s := server.NewMCPServer("lexis", "1.1.0",
		server.WithToolCapabilities(false),
	)

	s.AddTool(
		mcpgo.NewTool("search_documents",
			mcpgo.WithDescription("Search the user's ingested documents (PDF, Markdown, HTML, text) for relevant passages. Combines semantic similarity with keyword matching, so literal references such as rule numbers also work. Use this whenever the user asks about a topic that may be covered in their documents."),
			mcpgo.WithString("query",
				mcpgo.Required(),
				mcpgo.Description("Search query to find relevant passages"),
			),
			mcpgo.WithString("source",
				mcpgo.Description("Restrict the search to one source (see list_sources). Omit to search everything."),
			),
			mcpgo.WithNumber("top_k",
				mcpgo.Description(fmt.Sprintf("Number of passages to return (1-%d, default %d)", maxTopK, topK)),
			),
		),
		func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			query, err := req.RequireString("query")
			if err != nil {
				return mcpgo.NewToolResultError("query parameter is required"), nil
			}
			opts := library.SearchOptions{
				TopK:   min(max(req.GetInt("top_k", topK), 1), maxTopK),
				Source: req.GetString("source", ""),
			}

			hits, err := lib.Search(ctx, query, opts)
			if err != nil {
				return mcpgo.NewToolResultError(fmt.Sprintf("search failed: %v", err)), nil
			}
			if len(hits) == 0 {
				return mcpgo.NewToolResultText("No relevant passages found."), nil
			}

			var sb strings.Builder
			if citationEnforcement {
				sb.WriteString(citationInstruction)
			}
			for i, h := range hits {
				fmt.Fprintf(&sb, "[%d] Source: %s (%s)\nText: %s\n",
					i+1, library.Locator(h.Document), matchLabel(h), h.Text)
				if i < len(hits)-1 {
					sb.WriteString("\n---\n\n")
				}
			}
			return mcpgo.NewToolResultText(sb.String()), nil
		},
	)

	s.AddTool(
		mcpgo.NewTool("get_chunk",
			mcpgo.WithDescription("Read a specific passage by source and chunk number, optionally with the chunks before and after it. Use it when a search result looks cut off. Consecutive chunks overlap slightly."),
			mcpgo.WithString("source",
				mcpgo.Required(),
				mcpgo.Description("Source name, as shown in a search result"),
			),
			mcpgo.WithNumber("chunk",
				mcpgo.Required(),
				mcpgo.Description("Chunk number, as shown in a search result"),
			),
			mcpgo.WithNumber("neighbors",
				mcpgo.Description(fmt.Sprintf("Chunks to include on each side (0-%d, default 0)", maxNeighbors)),
			),
		),
		func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			source, err := req.RequireString("source")
			if err != nil {
				return mcpgo.NewToolResultError("source parameter is required"), nil
			}
			index, err := req.RequireInt("chunk")
			if err != nil {
				return mcpgo.NewToolResultError("chunk parameter is required"), nil
			}
			neighbors := min(max(req.GetInt("neighbors", 0), 0), maxNeighbors)

			docs, err := lib.Chunk(ctx, source, index, neighbors)
			if err != nil {
				return mcpgo.NewToolResultError(err.Error()), nil
			}

			var sb strings.Builder
			for i, d := range docs {
				fmt.Fprintf(&sb, "Source: %s\nText: %s\n", library.Locator(d), d.Text)
				if i < len(docs)-1 {
					sb.WriteString("\n---\n\n")
				}
			}
			return mcpgo.NewToolResultText(sb.String()), nil
		},
	)

	s.AddTool(
		mcpgo.NewTool("list_sources",
			mcpgo.WithDescription("List the names of all ingested documents, usable as the source filter of search_documents."),
		),
		func(ctx context.Context, _ mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			sources, err := lib.Sources(ctx)
			if err != nil {
				return mcpgo.NewToolResultError(fmt.Sprintf("listing sources failed: %v", err)), nil
			}
			if len(sources) == 0 {
				return mcpgo.NewToolResultText("No sources ingested yet."), nil
			}
			return mcpgo.NewToolResultText(strings.Join(sources, "\n")), nil
		},
	)

	return server.ServeStdio(s)
}

// matchLabel says how a hit was found, with the similarity when it has one.
func matchLabel(h library.Hit) string {
	if h.Match == "keyword" {
		return "keyword match"
	}
	return fmt.Sprintf("similarity %.2f, %s", h.Score, h.Match)
}
