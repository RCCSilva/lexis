package store

import "context"

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
	Search(ctx context.Context, vector []float32, topK int) ([]Result, error)
	Sources(ctx context.Context) ([]string, error)
	Delete(ctx context.Context, source string) error
}
