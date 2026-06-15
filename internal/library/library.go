package library

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/rafael/lexis/internal/chunk"
	"github.com/rafael/lexis/internal/embed"
	"github.com/rafael/lexis/internal/extract"
	"github.com/rafael/lexis/internal/rerank"
	"github.com/rafael/lexis/internal/store"
)

// Library orchestrates ingestion and search across pluggable providers.
type Library struct {
	extractor extract.Extractor
	chunker   chunk.Chunker
	embedder  embed.Embedder
	store     store.Store
	reranker  rerank.Reranker // nil = no reranking
}

func New(e extract.Extractor, c chunk.Chunker, emb embed.Embedder, s store.Store, r rerank.Reranker) *Library {
	return &Library{extractor: e, chunker: c, embedder: emb, store: s, reranker: r}
}

// Ingest extracts, chunks, embeds and stores all content from path.
// source is the human-readable name stored as metadata (defaults to filename without extension).
func (l *Library) Ingest(ctx context.Context, path string) error {
	source := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))

	pages, err := l.extractor.Extract(ctx, path)
	if err != nil {
		return fmt.Errorf("extract %s: %w", path, err)
	}
	if len(pages) == 0 {
		return fmt.Errorf("no text extracted from %s", path)
	}

	chunks, err := l.chunker.Chunk(ctx, pages)
	if err != nil {
		return fmt.Errorf("chunk %s: %w", path, err)
	}
	if len(chunks) == 0 {
		return fmt.Errorf("no chunks produced from %s", path)
	}

	texts := make([]string, len(chunks))
	for i, c := range chunks {
		texts[i] = c.Text
	}

	vecs, err := l.embedder.Embed(ctx, texts)
	if err != nil {
		return fmt.Errorf("embed %s: %w", path, err)
	}

	docs := make([]store.Document, len(chunks))
	for i, c := range chunks {
		docs[i] = store.Document{
			ID:     fmt.Sprintf("%s-chunk-%d", source, i),
			Text:   c.Text,
			Vector: vecs[i],
			Metadata: map[string]string{
				"source": source,
			},
		}
	}

	if err := l.store.Add(ctx, docs); err != nil {
		return fmt.Errorf("store %s: %w", path, err)
	}
	return nil
}

// Search returns the topK most relevant results for query.
func (l *Library) Search(ctx context.Context, query string, topK int) ([]store.Result, error) {
	vecs, err := l.embedder.Embed(ctx, []string{query})
	if err != nil {
		return nil, fmt.Errorf("embed query: %w", err)
	}

	results, err := l.store.Search(ctx, vecs[0], topK)
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}

	if l.reranker != nil && len(results) > 0 {
		results, err = l.reranker.Rerank(ctx, query, results)
		if err != nil {
			return nil, fmt.Errorf("rerank: %w", err)
		}
	}

	return results, nil
}

// Sources returns all ingested source names.
func (l *Library) Sources(ctx context.Context) ([]string, error) {
	return l.store.Sources(ctx)
}

// Delete removes all chunks for the given source.
func (l *Library) Delete(ctx context.Context, source string) error {
	return l.store.Delete(ctx, source)
}
