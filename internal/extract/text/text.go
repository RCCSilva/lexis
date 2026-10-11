// Package text extracts plain text files.
package text

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/RCCSilva/lexis/internal/extract"
)

// Extractor reads a file as a single unpaged, unlabelled page.
type Extractor struct{}

func New() *Extractor { return &Extractor{} }

func (e *Extractor) Extract(_ context.Context, path string) ([]extract.Page, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	t := strings.TrimSpace(strings.ToValidUTF8(string(data), ""))
	if t == "" {
		return nil, nil
	}
	return []extract.Page{{Text: t}}, nil
}
