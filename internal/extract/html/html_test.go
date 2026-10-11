package html

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func pagesOf(t *testing.T, src string) []string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "doc.html")
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	pages, err := New().Extract(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, p := range pages {
		out = append(out, p.Label+"|"+strings.Join(strings.Fields(p.Text), " "))
	}
	return out
}

func TestExtractUsesMainAndHeadings(t *testing.T) {
	got := pagesOf(t, `<html><head><title>x</title><style>p{}</style></head><body>
<nav><a href="/">Home</a></nav>
<main>
  <h1>Guide</h1><p>Opening &amp; words.</p>
  <section><h2>Setup</h2><p>Install <code>it</code>.</p><script>evil()</script>
    <h3>Linux</h3><ul><li>one</li><li>two</li></ul></section>
  <h2>Usage</h2><table><tr><td>a</td><td>b</td></tr></table>
</main>
<footer>copyright</footer></body></html>`)
	want := []string{
		"Guide|Opening & words.",
		"Guide › Setup|Install it.",
		"Guide › Setup › Linux|one two",
		"Guide › Usage|a b",
	}
	if len(got) != len(want) {
		t.Fatalf("got %q, want %q", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("page %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestExtractWithoutStructureIsOneUnlabelledPage(t *testing.T) {
	got := pagesOf(t, `<div>some <b>plain</b> words</div>`)
	if len(got) != 1 || got[0] != "|some plain words" {
		t.Fatalf("got %q", got)
	}
}
