package word

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/RCCSilva/lexis/internal/extract"
)

func words(n int, prefix string) string {
	w := make([]string, n)
	for i := range w {
		w[i] = prefix
	}
	return strings.Join(w, " ")
}

func TestChunkSpansPageBreaks(t *testing.T) {
	c := New(6, 2, 0)
	chunks, err := c.Chunk(context.Background(), []extract.Page{
		{Number: 1, Text: words(4, "a")},
		{Number: 2, Text: words(4, "b")},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(chunks) != 2 {
		t.Fatalf("got %d chunks, want 2", len(chunks))
	}
	if !reflect.DeepEqual(chunks[0].Pages, []int{1, 2}) {
		t.Errorf("chunk 0 pages = %v, want [1 2]", chunks[0].Pages)
	}
}

func TestChunkNeverMixesSections(t *testing.T) {
	c := New(10, 2, 0)
	chunks, err := c.Chunk(context.Background(), []extract.Page{
		{Label: "A", Text: words(3, "a")},
		{Label: "B", Text: words(25, "b")},
		{Label: "C", Text: words(2, "c")},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, ch := range chunks {
		for _, w := range strings.Fields(ch.Text) {
			if w != strings.ToLower(ch.Label) {
				t.Fatalf("chunk labelled %q holds word %q: %+v", ch.Label, w, ch)
			}
		}
		if len(ch.Pages) != 0 {
			t.Errorf("section chunk has pages %v", ch.Pages)
		}
	}
	var labels []string
	for _, ch := range chunks {
		labels = append(labels, ch.Label)
	}
	// A fits in one chunk, B needs ceil((25-2)/8)=3 windows, C one.
	if got := strings.Join(labels, ""); got != "ABBBC" {
		t.Fatalf("labels = %s, want ABBBC", got)
	}
}

func TestChunkBoundsCharacters(t *testing.T) {
	c := New(100, 2, 20)
	chunks, err := c.Chunk(context.Background(), []extract.Page{
		{Text: words(30, "abcd")}, // 4 chars plus a space each
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(chunks) < 2 {
		t.Fatalf("got %d chunks, want the char bound to split the text", len(chunks))
	}
	for i, ch := range chunks {
		if len(ch.Text) > 20 {
			t.Errorf("chunk %d has %d chars, want <= 20", i, len(ch.Text))
		}
	}
	// Overlap keeps the last two words of one chunk at the start of the next.
	first := strings.Fields(chunks[0].Text)
	if len(first) != 4 || chunks[1].Text[:4] != "abcd" {
		t.Errorf("unexpected chunking: %q", chunks[0].Text)
	}
}

func TestChunkAlwaysProgresses(t *testing.T) {
	// Overlap as large as the window must not loop forever.
	c := New(4, 4, 0)
	chunks, err := c.Chunk(context.Background(), []extract.Page{{Text: words(10, "w")}})
	if err != nil || len(chunks) == 0 {
		t.Fatalf("chunks = %d, err = %v", len(chunks), err)
	}
}
