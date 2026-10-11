// Package section turns a document's heading structure into extract.Pages.
package section

import (
	"strings"

	"github.com/RCCSilva/lexis/internal/extract"
)

// Separator joins the headings of a section's path in its label.
const Separator = " › "

// MaxLevel is the deepest heading level that opens a new section. Deeper
// headings stay inside the section they belong to as plain text.
const MaxLevel = 4

// Section is the text under a heading, with the headings that lead to it.
type Section struct {
	Path []string
	Text string
}

// Pages converts sections to pages labelled with their heading path. Empty
// sections are dropped. A document with no headings yields one unlabelled page.
func Pages(sections []Section) []extract.Page {
	var pages []extract.Page
	for _, s := range sections {
		text := strings.TrimSpace(s.Text)
		if text == "" {
			continue
		}
		pages = append(pages, extract.Page{
			Label: strings.Join(s.Path, Separator),
			Text:  text,
		})
	}
	return pages
}

// Builder accumulates sections while a parser walks a document.
type Builder struct {
	sections []Section
	stack    []heading
	text     strings.Builder
}

type heading struct {
	level int
	title string
}

// Heading closes the current section and opens one for a heading of the given
// level. Levels deeper than MaxLevel are treated as text.
func (b *Builder) Heading(level int, title string) {
	title = strings.Join(strings.Fields(title), " ")
	if level > MaxLevel || title == "" {
		b.Text(title)
		return
	}
	b.flush()
	for len(b.stack) > 0 && b.stack[len(b.stack)-1].level >= level {
		b.stack = b.stack[:len(b.stack)-1]
	}
	b.stack = append(b.stack, heading{level, title})
}

// Text appends body text to the current section.
func (b *Builder) Text(s string) {
	b.text.WriteString(s)
}

// Sections returns everything accumulated so far.
func (b *Builder) Sections() []Section {
	b.flush()
	return b.sections
}

func (b *Builder) flush() {
	if strings.TrimSpace(b.text.String()) != "" {
		path := make([]string, len(b.stack))
		for i, h := range b.stack {
			path[i] = h.title
		}
		b.sections = append(b.sections, Section{Path: path, Text: b.text.String()})
	}
	b.text.Reset()
}
