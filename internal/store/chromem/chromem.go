package chromem

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/RCCSilva/lexis/internal/store"
	chromemgo "github.com/philippgille/chromem-go"
)

const collectionName = "lexis"

// Store is a vector store backed by chromem-go (embedded, persistent).
type Store struct {
	db        *chromemgo.DB
	col       *chromemgo.Collection
	indexPath string                      // path to sources.json
	sources   map[string]int              // source -> chunk count
	texts     map[string][]store.Document // source -> docs without vectors, filled lazily
}

// New opens (or creates) a persistent chromem-go DB at the given path.
func New(path, model, embedURL string) (*Store, error) {
	if err := os.MkdirAll(path, 0o755); err != nil {
		return nil, fmt.Errorf("creating store dir: %w", err)
	}
	db, err := chromemgo.NewPersistentDB(path, false)
	if err != nil {
		return nil, fmt.Errorf("opening chromem DB: %w", err)
	}
	embedFn := chromemgo.NewEmbeddingFuncOllama(model, embedURL)
	col, err := db.GetOrCreateCollection(collectionName, nil, embedFn)
	if err != nil {
		return nil, fmt.Errorf("getting collection: %w", err)
	}

	indexPath := filepath.Join(path, "sources.json")
	sources, err := loadSources(indexPath)
	if err != nil {
		return nil, err
	}

	return &Store{db: db, col: col, indexPath: indexPath, sources: sources, texts: make(map[string][]store.Document)}, nil
}

func (s *Store) Add(ctx context.Context, docs []store.Document) error {
	clear(s.texts)
	cdocs := make([]chromemgo.Document, len(docs))
	for i, d := range docs {
		cdocs[i] = chromemgo.Document{
			ID:        d.ID,
			Content:   d.Text,
			Metadata:  d.Metadata,
			Embedding: d.Vector,
		}
		if src, ok := d.Metadata["source"]; ok {
			s.sources[src] += 1
		}
	}
	if err := s.col.AddDocuments(ctx, cdocs, 4); err != nil {
		return err
	}
	return saveSources(s.indexPath, s.sources)
}

func (s *Store) Search(ctx context.Context, vector []float32, topK int, source string) ([]store.Result, error) {
	count := s.col.Count()
	var where map[string]string
	if source != "" {
		n, ok := s.sources[source]
		if !ok {
			return nil, fmt.Errorf("source %q not found", source)
		}
		where = map[string]string{"source": source}
		count = n
	}
	if count == 0 {
		return nil, nil
	}
	if topK > count {
		topK = count
	}
	raw, err := s.col.QueryEmbedding(ctx, vector, topK, where, nil)
	if err != nil {
		return nil, err
	}
	results := make([]store.Result, len(raw))
	for i, r := range raw {
		results[i] = store.Result{
			Document: store.Document{
				ID:       r.ID,
				Text:     r.Content,
				Metadata: r.Metadata,
			},
			Score: r.Similarity,
		}
	}
	return results, nil
}

func (s *Store) Get(ctx context.Context, id string) (store.Document, error) {
	d, err := s.col.GetByID(ctx, id)
	if err != nil {
		return store.Document{}, err
	}
	return store.Document{ID: d.ID, Text: d.Content, Metadata: d.Metadata}, nil
}

func (s *Store) Documents(ctx context.Context, source string) ([]store.Document, error) {
	if source != "" {
		if _, ok := s.sources[source]; !ok {
			return nil, fmt.Errorf("source %q not found", source)
		}
		return s.sourceDocs(ctx, source), nil
	}
	var all []store.Document
	for src := range s.sources {
		all = append(all, s.sourceDocs(ctx, src)...)
	}
	return all, nil
}

// sourceDocs loads the chunks of one source, relying on the
// "<source>-chunk-<i>" ID scheme and the per-source chunk count.
func (s *Store) sourceDocs(ctx context.Context, source string) []store.Document {
	if docs, ok := s.texts[source]; ok {
		return docs
	}
	docs := make([]store.Document, 0, s.sources[source])
	for i := 0; i < s.sources[source]; i++ {
		d, err := s.Get(ctx, store.ChunkID(source, i))
		if err != nil {
			continue
		}
		docs = append(docs, d)
	}
	s.texts[source] = docs
	return docs
}

func (s *Store) Sources(_ context.Context) ([]string, error) {
	srcs := make([]string, 0, len(s.sources))
	for src := range s.sources {
		srcs = append(srcs, src)
	}
	sort.Strings(srcs)
	return srcs, nil
}

func (s *Store) Delete(ctx context.Context, source string) error {
	if _, ok := s.sources[source]; !ok {
		return fmt.Errorf("source %q not found", source)
	}
	if err := s.col.Delete(ctx, map[string]string{"source": source}, nil); err != nil {
		return err
	}
	delete(s.texts, source)
	delete(s.sources, source)
	return saveSources(s.indexPath, s.sources)
}

func loadSources(path string) (map[string]int, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return make(map[string]int), nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading sources index: %w", err)
	}
	var m map[string]int
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parsing sources index: %w", err)
	}
	return m, nil
}

func saveSources(path string, sources map[string]int) error {
	data, err := json.MarshalIndent(sources, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
