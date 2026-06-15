package extract

import "context"

// Page is a single page of text extracted from a document.
type Page struct {
	Number int
	Text   string
}

// Extractor extracts pages of text from a document file.
type Extractor interface {
	Extract(ctx context.Context, path string) ([]Page, error)
}
