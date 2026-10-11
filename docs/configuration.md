# Configuration

Optional. Create `~/.lexis/config.yaml`; omitted keys keep their defaults.

```yaml
store:
  provider: chromem            # on-disk vector database
  path: ~/.lexis/chromem-db
embedder:
  provider: ollama
  model: nomic-embed-text
  url: ""                      # empty = Ollama's default, http://localhost:11434
chunker:
  strategy: word
  size: 800                    # words per chunk
  overlap: 125                 # words shared by neighbouring chunks
mcp:
  top_k: 5                     # default passages per search
  citation_enforcement: true
```

## Changing settings

Stored data is not updated when settings change. After changing `embedder.model`, `chunker.size` or `chunker.overlap`, re-ingest your documents.

Vectors from different models are not comparable. When switching models, also set a new `store.path`.

## Testing without touching your library

Config and the database both live under `$HOME/.lexis`. Run with another `HOME` to use a separate library:

```bash
HOME=/tmp/lexis-test ./lexis ingest ./samples
```
