# CLAUDE.md

This file provides guidance to Claude Code when working with code in this repository.

## Build

Go is installed at `/home/rafael/go-dist/go/bin/go` (not on the system PATH by default).

```bash
export PATH="/home/rafael/go-dist/go/bin:$PATH"

go build -o lexis ./cmd/lexis/   # compile
go vet ./...                      # lint
go test ./...                     # tests
```

The compiled binary lands at `./lexis`.

## Runtime dependencies

- **Ollama** must be running locally with `nomic-embed-text` pulled (`ollama pull nomic-embed-text`).
- **pdftotext** (poppler-utils) must be on PATH for ingestion.

## Config

User config lives at `~/.lexis/config.yaml`. If absent, safe defaults are used (Ollama embedder, chromem store at `~/.lexis/chromem-db`).

## Architecture

Package-oriented design following https://www.ardanlabs.com/blog/2017/02/package-oriented-design.html

```
cmd/lexis/          CLI entry point — wires providers, dispatches subcommands
config/             Config types + YAML loader (~/.lexis/config.yaml)
internal/
  extract/          Extractor interface + Page type
    pdftotext/      pdftotext implementation
  chunk/            Chunker interface + Chunk type
    word/           Word-based sliding window chunker
  embed/            Embedder interface
    ollama/         Ollama implementation
  store/            Store interface + Document/Result types
    chromem/        chromem-go persistent vector DB implementation
  rerank/           Reranker interface (no implementation yet)
  library/          Orchestrator: wires Extractor+Chunker+Embedder+Store+Reranker
  mcp/              MCP stdio server with citation enforcement
```

**Adding a new provider** (e.g. VoyageAI embedder):
1. Create `internal/embed/voyageai/voyageai.go` implementing `embed.Embedder`
2. Add config fields to `config/config.go`
3. Wire it in `cmd/lexis/main.go` `buildLibrary()`

## CLI

```
lexis ingest <pdf-file-or-directory>   # extract, embed and store PDFs
lexis search "<query>"                 # search from terminal
lexis list                             # list ingested sources
lexis remove <source>                  # remove a source
lexis mcp                              # start MCP server (stdio)
```

## MCP

Registered in `~/.claude/settings.json` under `mcpServers.lexis`. Responses include a citation enforcement instruction so Claude always cites sources inline.
