package ollama

import (
	"context"
	"fmt"

	chromem "github.com/philippgille/chromem-go"
)

// Embedder embeds text using a model served by Ollama.
type Embedder struct {
	fn chromem.EmbeddingFunc
}

func New(model, baseURL string) *Embedder {
	return &Embedder{fn: chromem.NewEmbeddingFuncOllama(model, baseURL)}
}

// Embed returns one vector per input text.
func (e *Embedder) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	vecs := make([][]float32, len(texts))
	for i, t := range texts {
		v, err := e.fn(ctx, t)
		if err != nil {
			return nil, fmt.Errorf("embedding text %d: %w", i, err)
		}
		vecs[i] = v
	}
	return vecs, nil
}
