package pdftotext

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/rafael/lexis/internal/extract"
)

// Extractor extracts text from PDFs using the pdftotext CLI (poppler-utils).
type Extractor struct{}

func New() *Extractor { return &Extractor{} }

func (e *Extractor) Extract(_ context.Context, path string) ([]extract.Page, error) {
	out, err := exec.Command("pdftotext", "-layout", path, "-").Output()
	if err != nil {
		return nil, fmt.Errorf("pdftotext failed for %s: %w", path, err)
	}
	raw := strings.Split(string(out), "\f")
	pages := make([]extract.Page, 0, len(raw))
	for i, text := range raw {
		if strings.TrimSpace(text) != "" {
			pages = append(pages, extract.Page{Number: i + 1, Text: text})
		}
	}
	return pages, nil
}
