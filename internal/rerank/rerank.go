package rerank

import (
	"context"

	"github.com/RCCSilva/lexis/internal/store"
)

// Reranker reorders search results by relevance to the query.
// It is optional; pass nil to the Library to skip reranking.
type Reranker interface {
	Rerank(ctx context.Context, query string, results []store.Result) ([]store.Result, error)
}
