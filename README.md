# Lexis

Local search over your own documents, built for answers **with citations**. Ingest PDFs, Markdown, HTML and text; Lexis returns the passages that match a question and where each one came from (document and page or section).

It runs on your machine, using a local embedding model through [Ollama](https://ollama.com). Use it from the command line or as an [MCP](https://modelcontextprotocol.io) server so an assistant can search and cite your documents.

## Quick start

1. [Install the requirements](docs/install.md): Go, Ollama with `nomic-embed-text`, and `pdftotext` for PDFs.
2. Build and use:

```bash
go build -o lexis ./cmd/lexis/
./lexis ingest ~/Documents/manuals          # a file or a directory
./lexis search "how do I reset the device"
```

```
[1] user-manual, pages 34-35, chunk 42 (semantic+keyword)
To reset the device, hold the power button for ten seconds ...
```

## Commands

| Command | What it does |
|---|---|
| `lexis ingest <file-or-directory>` | Extract, embed and store documents. Re-ingesting replaces. |
| `lexis search [--source S] [-n N] "<query>"` | Hybrid search (semantic + keyword). |
| `lexis chunk <source> <n> [neighbors]` | Print a chunk and its neighbours. |
| `lexis list` | List ingested sources. |
| `lexis remove <source>` | Remove a source. |
| `lexis mcp` | Start the MCP server on stdio. |

A source is named after its file, without the extension.

## Formats

`.pdf`, `.md`/`.markdown`, `.html`/`.htm`, `.txt`/`.text`. See [docs/formats.md](docs/formats.md) for how each is split and cited.

## Documentation

- [Install](docs/install.md): requirements and setup
- [Formats](docs/formats.md): what is supported and how citations look
- [MCP](docs/mcp.md): using Lexis from an assistant
- [Configuration](docs/configuration.md): `~/.lexis/config.yaml`
- [Architecture](docs/architecture.md): how it is built
- [Contributing](CONTRIBUTING.md): build, test, add a format or provider

## License

[MIT](LICENSE)
