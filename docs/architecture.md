# Architecture

Package-oriented design (https://www.ardanlabs.com/blog/2017/02/package-oriented-design.html). Each package owns one interface. `library` is the only place that wires them together, and `cmd/lexis` chooses the concrete providers.

## Flow

```
file -> extract -> chunk -> embed -> store
                                      |
question -> embed -> vector search ---+
         -> keyword search (BM25) ----+-> fuse (RRF) -> passages + citations
```

## Layout

```
cmd/lexis/          CLI entry point: builds providers, dispatches subcommands
config/             Config types and YAML loader
internal/
  extract/          Extractor interface and Page type
    pdftotext/      PDF via the pdftotext CLI; one Page per PDF page
    markdown/       one Page per heading (# to ####)
    html/           one Page per heading (h1 to h4); reads <main>/<article>
    text/           a single unpaged Page
    section/        shared heading-path to Page conversion
    router/         picks an extractor by file extension
  chunk/            Chunker interface and Chunk type
    word/           overlapping word windows; never crosses a section
  embed/            Embedder interface
    ollama/         Ollama implementation
  store/            Store interface and Document/Result types
    chromem/        chromem-go persistent vector database
  rerank/           Reranker interface (no implementation yet)
  library/          Orchestrator: ingest, hybrid search, chunk lookup
  mcp/              MCP stdio server with citation enforcement
```

## Data model

- `extract.Page{Number, Label, Text}`: `Number` is a PDF page (0 when the format has none). `Label` is a heading path such as `Guide › Setup` (empty for physical pages).
- `chunk.Chunk{Text, Pages, Label}`: carries the location of its text.
- Metadata stored per chunk: `source` (file name without extension), `chunk` (index), and either `pages` or `section`.
- Chunk IDs are `<source>-chunk-<i>` (`store.ChunkID`). `library.Locator` renders a citation from the metadata.
- Re-ingesting a source deletes its old chunks first.

## Search

`library.Search` embeds the query and runs a vector search and a BM25 search over a pool of `max(top_k*4, 20)`. The two rankings are merged with Reciprocal Rank Fusion (k=60). `Hit.Match` records which ranker found each result.

The tokenizer keeps references such as `4-2-1` whole, so literal lookups work. See `internal/library/hybrid.go`.

## Limits

- Ollama rejects an embedding batch over 2048 tokens. The PDF extractor collapses runs of `.`, `_` and `-` (index dot leaders) because each character becomes its own token.
- The reranker has an interface but no implementation.
