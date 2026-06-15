package chunk

import (
	"context"

	"github.com/rafael/lexis/internal/extract"
)

// Chunk is a piece of text derived from one or more document pages.
type Chunk struct {
	Text  string
	Pages []int // page numbers this chunk spans
}

// Chunker splits extracted pages into chunks suitable for embedding.
type Chunker interface {
	Chunk(ctx context.Context, pages []extract.Page) ([]Chunk, error)
}
