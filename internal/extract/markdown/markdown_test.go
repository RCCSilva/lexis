package markdown

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const doc = `---
title: ignored
---
# Guide

Opening words.

## Setup

![diagram](a.png) See [the docs](http://example.com) for more.

` + "```sh" + `
# not a heading
make install
` + "```" + `

## Usage ##

Run <b>it</b>.
`

func TestExtractSplitsOnHeadings(t *testing.T) {
	path := filepath.Join(t.TempDir(), "guide.md")
	if err := os.WriteFile(path, []byte(doc), 0o644); err != nil {
		t.Fatal(err)
	}
	pages, err := New().Extract(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	if len(pages) != 3 {
		t.Fatalf("got %d pages, want 3: %+v", len(pages), pages)
	}
	labels := []string{"Guide", "Guide › Setup", "Guide › Usage"}
	for i, want := range labels {
		if pages[i].Label != want {
			t.Errorf("page %d label = %q, want %q", i, pages[i].Label, want)
		}
	}
	setup := pages[1].Text
	for _, want := range []string{"See the docs for more.", "# not a heading", "make install"} {
		if !strings.Contains(setup, want) {
			t.Errorf("setup text missing %q: %q", want, setup)
		}
	}
	if strings.Contains(setup, "diagram") || strings.Contains(setup, "http://") {
		t.Errorf("markup leaked into text: %q", setup)
	}
	if strings.Contains(pages[0].Text, "title") || pages[2].Text != "Run it." {
		t.Errorf("front matter or tags leaked: %q / %q", pages[0].Text, pages[2].Text)
	}
}

func TestExtractWithoutHeadingsIsOneUnlabelledPage(t *testing.T) {
	path := filepath.Join(t.TempDir(), "plain.md")
	os.WriteFile(path, []byte("just some words\n"), 0o644)
	pages, err := New().Extract(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	if len(pages) != 1 || pages[0].Label != "" || pages[0].Text != "just some words" {
		t.Fatalf("pages = %+v", pages)
	}
}
