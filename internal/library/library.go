package library

import (
	"context"
	"fmt"
	"path/filepath"
	"slices"
	"strconv"
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

	// Re-ingesting replaces the source instead of duplicating its chunks.
	if existing, _ := l.store.Sources(ctx); slices.Contains(existing, source) {
		if err := l.store.Delete(ctx, source); err != nil {
			return fmt.Errorf("replace %s: %w", path, err)
		}
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
			ID:     store.ChunkID(source, i),
			Text:   c.Text,
			Vector: vecs[i],
			Metadata: map[string]string{
				"source": source,
				"chunk":  strconv.Itoa(i),
				"pages":  pagesLabel(c.Pages),
			},
		}
	}

	if err := l.store.Add(ctx, docs); err != nil {
		return fmt.Errorf("store %s: %w", path, err)
	}
	return nil
}

// SearchOptions tunes a search.
type SearchOptions struct {
	TopK   int    // number of results; defaults to 5
	Source string // restrict to one source; empty searches everything
}

// Hit is a search result and how it was found.
type Hit struct {
	store.Result
	Match string // "semantic", "keyword" or "semantic+keyword"
}

// Search combines vector similarity with BM25 keyword matching, so both
// paraphrased questions and literal references (e.g. "8-7-5") are found.
func (l *Library) Search(ctx context.Context, query string, opts SearchOptions) ([]Hit, error) {
	topK := opts.TopK
	if topK <= 0 {
		topK = 5
	}
	pool := max(topK*4, 20)

	vecs, err := l.embedder.Embed(ctx, []string{query})
	if err != nil {
		return nil, fmt.Errorf("embed query: %w", err)
	}
	semantic, err := l.store.Search(ctx, vecs[0], pool, opts.Source)
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}
	docs, err := l.store.Documents(ctx, opts.Source)
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}

	hits := fuse(semantic, bm25(query, docs, pool), topK)

	if l.reranker != nil && len(hits) > 0 {
		results := make([]store.Result, len(hits))
		byID := make(map[string]Hit, len(hits))
		for i, h := range hits {
			results[i] = h.Result
			byID[h.ID] = h
		}
		results, err = l.reranker.Rerank(ctx, query, results)
		if err != nil {
			return nil, fmt.Errorf("rerank: %w", err)
		}
		hits = hits[:0]
		for _, r := range results {
			h := byID[r.ID]
			h.Result = r
			hits = append(hits, h)
		}
	}

	return hits, nil
}

// Chunk returns the chunk at index of source, plus up to neighbors chunks on
// each side (in document order) when they exist.
func (l *Library) Chunk(ctx context.Context, source string, index, neighbors int) ([]store.Document, error) {
	if _, err := l.store.Get(ctx, store.ChunkID(source, index)); err != nil {
		return nil, fmt.Errorf("chunk %d of %q not found", index, source)
	}
	var docs []store.Document
	for i := max(index-neighbors, 0); i <= index+neighbors; i++ {
		if d, err := l.store.Get(ctx, store.ChunkID(source, i)); err == nil {
			docs = append(docs, d)
		}
	}
	return docs, nil
}

// Locator describes where a chunk comes from, e.g. "rulebook, pages 34-35, chunk 42".
func Locator(d store.Document) string {
	s := d.Metadata["source"]
	if p := d.Metadata["pages"]; p != "" {
		if strings.Contains(p, "-") {
			s += ", pages " + p
		} else {
			s += ", page " + p
		}
	}
	if c := d.Metadata["chunk"]; c != "" {
		s += ", chunk " + c
	}
	return s
}

func pagesLabel(pages []int) string {
	switch len(pages) {
	case 0:
		return ""
	case 1:
		return strconv.Itoa(pages[0])
	}
	return fmt.Sprintf("%d-%d", pages[0], pages[len(pages)-1])
}

// Sources returns all ingested source names.
func (l *Library) Sources(ctx context.Context) ([]string, error) {
	return l.store.Sources(ctx)
}

// Delete removes all chunks for the given source.
func (l *Library) Delete(ctx context.Context, source string) error {
	return l.store.Delete(ctx, source)
}
