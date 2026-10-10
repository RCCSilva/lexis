package store

import (
	"context"
	"fmt"
)

// ChunkID is the ID of the i-th chunk of a source.
func ChunkID(source string, i int) string {
	return fmt.Sprintf("%s-chunk-%d", source, i)
}

// Document is a chunk of text with its embedding and metadata stored in the vector DB.
type Document struct {
	ID       string
	Text     string
	Metadata map[string]string
	Vector   []float32
}

// Result is a document returned by a similarity search, with its score.
type Result struct {
	Document
	Score float32
}

// Store persists and retrieves documents by vector similarity.
type Store interface {
	Add(ctx context.Context, docs []Document) error
	// Search returns the topK nearest documents. A non-empty source restricts
	// the search to that source.
	Search(ctx context.Context, vector []float32, topK int, source string) ([]Result, error)
	// Get returns one document by ID.
	Get(ctx context.Context, id string) (Document, error)
	// Documents returns the documents of a source (all sources if empty)
	// without their vectors, for lexical matching.
	Documents(ctx context.Context, source string) ([]Document, error)
	Sources(ctx context.Context) ([]string, error)
	Delete(ctx context.Context, source string) error
}
