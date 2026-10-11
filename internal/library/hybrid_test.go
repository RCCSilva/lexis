package library

import (
	"slices"
	"testing"

	"github.com/rafael/lexis/internal/store"
)

func TestTokenize(t *testing.T) {
	cases := map[string][]string{
		"Second-level CACHE.":       {"second", "level", "cache"},
		"see 8-7-5, and 12-2-8(b).": {"see", "8-7-5", "and", "12-2-8"},
		"Team A's ball":             {"team", "a", "s", "ball"},
		"...... 4-1-2":              {"4-1-2"},
		"":                          nil,
	}
	for in, want := range cases {
		if got := tokenize(in); !slices.Equal(got, want) {
			t.Errorf("tokenize(%q) = %q, want %q", in, got, want)
		}
	}
}

func docs(texts ...string) []store.Document {
	out := make([]store.Document, len(texts))
	for i, tx := range texts {
		out[i] = store.Document{ID: store.ChunkID("s", i), Text: tx}
	}
	return out
}

func TestBM25FindsLiteralReference(t *testing.T) {
	d := docs(
		"the cache rules are described in many places",
		"ARTICLE 5. SECOND LEVEL CACHE see 4-2-1 for details",
		"unrelated text about something else",
	)
	got := bm25("4-2-1", d, 5)
	if len(got) != 1 || got[0].ID != d[1].ID {
		t.Fatalf("bm25 literal = %+v, want only %s", got, d[1].ID)
	}
	if got := bm25("nothing matches zzz", d, 5); len(got) != 0 {
		t.Fatalf("expected no results, got %+v", got)
	}
}

func TestFuseMergesAndMarksMatches(t *testing.T) {
	d := docs("a", "b", "c")
	res := func(i int, score float32) store.Result { return store.Result{Document: d[i], Score: score} }

	hits := fuse(
		[]store.Result{res(0, 0.9), res(1, 0.8)},
		[]store.Result{res(1, 7), res(2, 5)},
		3,
	)
	if len(hits) != 3 {
		t.Fatalf("got %d hits, want 3", len(hits))
	}
	// chunk 1 is in both rankings, so it wins the fusion.
	if hits[0].ID != d[1].ID || hits[0].Match != "semantic+keyword" {
		t.Errorf("top hit = %s (%s)", hits[0].ID, hits[0].Match)
	}
	if hits[0].Score != 0.8 {
		t.Errorf("score = %v, want the vector similarity 0.8", hits[0].Score)
	}
	byID := map[string]Hit{}
	for _, h := range hits {
		byID[h.ID] = h
	}
	if h := byID[d[2].ID]; h.Match != "keyword" || h.Score != 0 {
		t.Errorf("keyword-only hit = %+v", h)
	}
	if got := fuse(nil, nil, 3); len(got) != 0 {
		t.Errorf("fuse(nil, nil) = %+v", got)
	}
	if got := fuse([]store.Result{res(0, 1), res(1, 1)}, nil, 1); len(got) != 1 {
		t.Errorf("topK not applied: %d hits", len(got))
	}
}

func TestPagesLabelAndLocator(t *testing.T) {
	if got := pagesLabel([]int{34}); got != "34" {
		t.Errorf("single page = %q", got)
	}
	if got := pagesLabel([]int{34, 35}); got != "34-35" {
		t.Errorf("range = %q", got)
	}
	d := store.Document{Metadata: map[string]string{"source": "rb", "pages": "34-35", "chunk": "42"}}
	if got := Locator(d); got != "rb, pages 34-35, chunk 42" {
		t.Errorf("Locator = %q", got)
	}
	d.Metadata["pages"] = "7"
	if got := Locator(d); got != "rb, page 7, chunk 42" {
		t.Errorf("Locator single = %q", got)
	}
	d = store.Document{Metadata: map[string]string{"source": "guide", "pages": "", "section": "Setup › Linux", "chunk": "3"}}
	if got := Locator(d); got != "guide, section Setup › Linux, chunk 3" {
		t.Errorf("Locator section = %q", got)
	}
}
