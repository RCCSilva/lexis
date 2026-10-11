# MCP server

`lexis mcp` serves the library over MCP on stdio, so an assistant can search your documents and cite them.

## Register

Claude Code:

```bash
claude mcp add lexis -- /absolute/path/to/lexis mcp
```

Other clients: run `/absolute/path/to/lexis mcp` as a stdio server. Restart the client after registering.

## Tools

| Tool | Parameters | Returns |
|---|---|---|
| `search_books` | `query`, `source?`, `top_k?` (1-20) | Passages with source, location, chunk number and match type. |
| `get_chunk` | `source`, `chunk`, `neighbors?` (0-3) | One chunk and its neighbours, for results that look cut off. |
| `list_sources` | none | Valid values for `source`. |

## Citations

Each search response starts with an instruction to answer only from the returned passages, cite every claim with the source and page or section, and say so when the documents do not contain the answer.

Turn it off with `mcp.citation_enforcement: false` in the [configuration](configuration.md).
