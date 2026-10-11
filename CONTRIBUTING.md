# Contributing

## Build and test

```bash
go build -o lexis ./cmd/lexis/
go vet ./...
go test ./...
gofmt -l .          # must print nothing
```

The unit tests need neither Ollama nor `pdftotext`. To try ingestion without touching your library, see [configuration](docs/configuration.md#testing-without-touching-your-library).

## Add a file format

1. Create `internal/extract/<name>/<name>.go` implementing `extract.Extractor`.
2. Return `Page`s: set `Number` for paged formats, or `Label` for structured ones (use `section.Builder` if the format has headings).
3. Register the extension in `extractors()` in `cmd/lexis/main.go`.
4. Add a test and a row in [docs/formats.md](docs/formats.md).

## Add a provider

For example another embedder:

1. Create `internal/embed/<name>/<name>.go` implementing `embed.Embedder`.
2. Add its config fields to `config/config.go`.
3. Wire it in `buildLibrary()` in `cmd/lexis/main.go`.

## Conventions

- The project is general purpose. Keep code, comments, tests and docs free of any particular subject matter.
- Treat the MCP tool names (`search_documents`, `get_chunk`, `list_sources`) as a public interface: clients register them by name, so renaming is a breaking change.
- See [docs/architecture.md](docs/architecture.md) before changing the pipeline.
- By contributing you agree that your contribution is licensed under the [MIT License](LICENSE).
