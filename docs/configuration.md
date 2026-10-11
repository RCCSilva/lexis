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
  max_chars: 3200              # hard limit per chunk; 0 = no limit
mcp:
  top_k: 5                     # default passages per search
  citation_enforcement: true
```

## Chunk size

A chunk ends at `size` words or `max_chars` characters, whichever comes first. The character limit keeps dense text, such as tables of coordinates, within the embedding model's input limit (2048 tokens for `nomic-embed-text`). Prose usually ends on the character limit, at roughly 500 words.

## Changing settings

Stored data is not updated when settings change. After changing `embedder.model`, `chunker.size` or `chunker.overlap`, re-ingest your documents.

Vectors from different models are not comparable. When switching models, also set a new `store.path`.

## Testing without touching your library

Config and the database both live under `$HOME/.lexis`. Run with another `HOME` to use a separate library:

```bash
HOME=/tmp/lexis-test ./lexis ingest ./samples
```
