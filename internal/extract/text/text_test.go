package text

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestExtractIsOneUnpagedPage(t *testing.T) {
	path := filepath.Join(t.TempDir(), "notes.txt")
	os.WriteFile(path, []byte("\n  hello\nworld \n"), 0o644)
	pages, err := New().Extract(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	if len(pages) != 1 || pages[0].Number != 0 || pages[0].Label != "" || pages[0].Text != "hello\nworld" {
		t.Fatalf("pages = %+v", pages)
	}
}

func TestExtractEmptyFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "empty.txt")
	os.WriteFile(path, []byte("  \n"), 0o644)
	pages, err := New().Extract(context.Background(), path)
	if err != nil || len(pages) != 0 {
		t.Fatalf("pages = %+v, err = %v", pages, err)
	}
}
