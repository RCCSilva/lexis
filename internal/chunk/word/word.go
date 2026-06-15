package word

import (
	"context"
	"strings"

	"github.com/rafael/lexis/internal/chunk"
	"github.com/rafael/lexis/internal/extract"
)

// Chunker splits pages into overlapping word-based chunks.
type Chunker struct {
	Size    int // words per chunk
	Overlap int // words shared between consecutive chunks
}

func New(size, overlap int) *Chunker {
	return &Chunker{Size: size, Overlap: overlap}
}

type wordWithPage struct {
	word string
	page int
}

func (c *Chunker) Chunk(_ context.Context, pages []extract.Page) ([]chunk.Chunk, error) {
	var words []wordWithPage
	for _, p := range pages {
		for _, w := range strings.Fields(p.Text) {
			words = append(words, wordWithPage{word: w, page: p.Number})
		}
	}
	if len(words) == 0 {
		return nil, nil
	}

	step := c.Size - c.Overlap
	var chunks []chunk.Chunk
	for start := 0; start < len(words); start += step {
		end := start + c.Size
		if end > len(words) {
			end = len(words)
		}
		slice := words[start:end]

		texts := make([]string, len(slice))
		pageSet := make(map[int]struct{})
		for i, w := range slice {
			texts[i] = w.word
			pageSet[w.page] = struct{}{}
		}

		text := strings.Join(texts, " ")
		if strings.TrimSpace(text) == "" {
			if end == len(words) {
				break
			}
			continue
		}

		spanPages := make([]int, 0, len(pageSet))
		// preserve order
		seen := make(map[int]bool)
		for _, w := range slice {
			if !seen[w.page] {
				spanPages = append(spanPages, w.page)
				seen[w.page] = true
			}
		}

		chunks = append(chunks, chunk.Chunk{Text: text, Pages: spanPages})
		if end == len(words) {
			break
		}
	}
	return chunks, nil
}
