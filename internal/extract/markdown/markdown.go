// Package markdown extracts text from Markdown files, one page per heading.
package markdown

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/RCCSilva/lexis/internal/extract"
	"github.com/RCCSilva/lexis/internal/extract/section"
)

var (
	atx    = regexp.MustCompile(`^ {0,3}(#{1,6})[ \t]+(.*?)(?:[ \t]+#+)?[ \t]*$`)
	fence  = regexp.MustCompile("^ {0,3}(`{3,}|~{3,})")
	image  = regexp.MustCompile(`!\[[^\]]*\]\([^)]*\)`)
	link   = regexp.MustCompile(`\[([^\]]*)\]\([^)]*\)`)
	htmlTg = regexp.MustCompile(`</?[A-Za-z][^>]*>`)
)

// Extractor splits Markdown on its ATX headings (# to ####).
type Extractor struct{}

func New() *Extractor { return &Extractor{} }

func (e *Extractor) Extract(_ context.Context, path string) ([]extract.Page, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	return section.Pages(parse(string(data))), nil
}

func parse(src string) []section.Section {
	lines := strings.Split(strings.ReplaceAll(src, "\r\n", "\n"), "\n")
	lines = skipFrontMatter(lines)

	var b section.Builder
	var open string // fence marker currently open, "" when outside code
	for _, line := range lines {
		if m := fence.FindStringSubmatch(line); m != nil {
			switch {
			case open == "":
				open = m[1]
			case strings.HasPrefix(m[1], open[:1]) && len(m[1]) >= len(open):
				open = ""
			}
			b.Text(line + "\n")
			continue
		}
		if open != "" {
			b.Text(line + "\n")
			continue
		}
		if m := atx.FindStringSubmatch(line); m != nil {
			b.Heading(len(m[1]), clean(m[2]))
			continue
		}
		b.Text(clean(line) + "\n")
	}
	return b.Sections()
}

// skipFrontMatter drops a leading "---" delimited metadata block.
func skipFrontMatter(lines []string) []string {
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return lines
	}
	for i := 1; i < len(lines); i++ {
		if t := strings.TrimSpace(lines[i]); t == "---" || t == "..." {
			return lines[i+1:]
		}
	}
	return lines
}

// clean removes markup that carries no searchable text.
func clean(s string) string {
	s = image.ReplaceAllString(s, "")
	s = link.ReplaceAllString(s, "$1")
	return htmlTg.ReplaceAllString(s, "")
}
