package word

import (
	"context"
	"strings"

	"github.com/RCCSilva/lexis/internal/chunk"
	"github.com/RCCSilva/lexis/internal/extract"
)

// Chunker splits pages into overlapping word-based chunks.
type Chunker struct {
	Size     int // words per chunk
	Overlap  int // words shared between consecutive chunks
	MaxChars int // upper bound on characters per chunk; 0 means no bound
}

// New returns a chunker of up to size words, bounded by maxChars characters.
// The bound matters for dense text such as coordinate tables, where a window
// of words can hold far more tokens than prose and overflow the embedder.
func New(size, overlap, maxChars int) *Chunker {
	return &Chunker{Size: size, Overlap: overlap, MaxChars: maxChars}
}

type wordWithPage struct {
	word string
	page int
}

// Chunk windows over the text. Physical pages are windowed together, so a
// chunk may span a page break. A labelled page (a section) is windowed on its
// own: a chunk never mixes text from two sections.
func (c *Chunker) Chunk(_ context.Context, pages []extract.Page) ([]chunk.Chunk, error) {
	var chunks []chunk.Chunk
	var run []extract.Page
	flush := func() {
		chunks = append(chunks, c.window(run)...)
		run = run[:0]
	}
	for _, p := range pages {
		if p.Label != "" {
			flush()
			run = append(run, p)
			flush()
			continue
		}
		run = append(run, p)
	}
	flush()
	return chunks, nil
}

// window cuts a run of pages into overlapping chunks.
func (c *Chunker) window(pages []extract.Page) []chunk.Chunk {
	var words []wordWithPage
	label := ""
	for _, p := range pages {
		label = p.Label
		for _, w := range strings.Fields(p.Text) {
			words = append(words, wordWithPage{word: w, page: p.Number})
		}
	}
	if len(words) == 0 {
		return nil
	}

	var chunks []chunk.Chunk
	for start := 0; start < len(words); {
		end, chars := start, 0
		for end < len(words) && end-start < c.Size {
			chars += len(words[end].word) + 1
			if c.MaxChars > 0 && chars > c.MaxChars && end > start {
				break
			}
			end++
		}
		slice := words[start:end]

		texts := make([]string, len(slice))
		var spanPages []int
		seen := make(map[int]bool)
		for i, w := range slice {
			texts[i] = w.word
			if w.page > 0 && !seen[w.page] {
				spanPages = append(spanPages, w.page)
				seen[w.page] = true
			}
		}

		chunks = append(chunks, chunk.Chunk{Text: strings.Join(texts, " "), Pages: spanPages, Label: label})
		if end == len(words) {
			break
		}
		// Advance, keeping the overlap but always making progress.
		start = max(end-c.Overlap, start+1)
	}
	return chunks
}
