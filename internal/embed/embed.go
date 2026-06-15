package embed

import "context"

// Embedder converts text into a vector representation.
type Embedder interface {
	Embed(ctx context.Context, texts []string) ([][]float32, error)
}
