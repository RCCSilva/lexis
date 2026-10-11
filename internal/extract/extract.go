package extract

import "context"

// Page is a unit of text extracted from a document: a physical page for paged
// formats, or a section for structured ones.
type Page struct {
	// Number is the page number, or 0 when the format has no pages.
	Number int
	// Label names where the text sits in the document, e.g. "Setup › Install"
	// for a section. Empty for physical pages.
	Label string
	Text  string
}

// Extractor extracts pages of text from a document file.
type Extractor interface {
	Extract(ctx context.Context, path string) ([]Page, error)
}
