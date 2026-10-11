// Package html extracts text from HTML files, one page per heading.
package html

import (
	"context"
	"fmt"
	"os"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/rafael/lexis/internal/extract"
	"github.com/rafael/lexis/internal/extract/section"
)

// Extractor splits HTML on its h1 to h4 headings. Navigation, page chrome,
// scripts and styles are ignored; when the document has <main> or <article>
// only that subtree is read.
type Extractor struct{}

func New() *Extractor { return &Extractor{} }

// skipped elements carry no document text.
var skipped = map[atom.Atom]bool{
	atom.Script: true, atom.Style: true, atom.Noscript: true, atom.Template: true,
	atom.Nav: true, atom.Footer: true, atom.Aside: true, atom.Svg: true,
	atom.Head: true, atom.Form: true, atom.Button: true, atom.Iframe: true,
}

// blocks start and end on their own line.
var blocks = map[atom.Atom]bool{
	atom.P: true, atom.Div: true, atom.Section: true, atom.Article: true,
	atom.Main: true, atom.Header: true, atom.Li: true, atom.Ul: true, atom.Ol: true,
	atom.Tr: true, atom.Table: true, atom.Pre: true, atom.Blockquote: true,
	atom.Br: true, atom.Hr: true, atom.Figure: true, atom.Figcaption: true,
	atom.Dl: true, atom.Dt: true, atom.Dd: true, atom.Details: true, atom.Summary: true,
}

func (e *Extractor) Extract(_ context.Context, path string) ([]extract.Page, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	defer f.Close()
	doc, err := html.Parse(f)
	if err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	var b section.Builder
	walk(contentRoot(doc), &b)
	return section.Pages(b.Sections()), nil
}

// contentRoot prefers <main>, then <article>, then the whole document.
func contentRoot(doc *html.Node) *html.Node {
	for _, a := range []atom.Atom{atom.Main, atom.Article} {
		if n := find(doc, a); n != nil {
			return n
		}
	}
	return doc
}

func find(n *html.Node, a atom.Atom) *html.Node {
	if n.Type == html.ElementNode && n.DataAtom == a {
		return n
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if r := find(c, a); r != nil {
			return r
		}
	}
	return nil
}

func walk(n *html.Node, b *section.Builder) {
	switch n.Type {
	case html.TextNode:
		b.Text(n.Data)
		return
	case html.ElementNode:
		if skipped[n.DataAtom] {
			return
		}
		if level := headingLevel(n.DataAtom); level > 0 {
			b.Heading(level, text(n))
			b.Text("\n")
			return
		}
		switch n.DataAtom {
		case atom.Td, atom.Th:
			b.Text(" ")
		}
		if blocks[n.DataAtom] {
			b.Text("\n")
			defer b.Text("\n")
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		walk(c, b)
	}
}

func headingLevel(a atom.Atom) int {
	switch a {
	case atom.H1:
		return 1
	case atom.H2:
		return 2
	case atom.H3:
		return 3
	case atom.H4:
		return 4
	case atom.H5, atom.H6:
		return 5 // deeper than section.MaxLevel: kept as text
	}
	return 0
}

// text returns all text under n.
func text(n *html.Node) string {
	var sb strings.Builder
	var rec func(*html.Node)
	rec = func(n *html.Node) {
		if n.Type == html.TextNode {
			sb.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			rec(c)
		}
	}
	rec(n)
	return sb.String()
}
