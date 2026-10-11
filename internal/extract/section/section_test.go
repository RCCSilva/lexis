package section

import (
	"reflect"
	"testing"
)

func TestBuilderTracksHeadingPath(t *testing.T) {
	var b Builder
	b.Text("intro\n")
	b.Heading(1, "Guide")
	b.Heading(2, "Setup")
	b.Text("install it\n")
	b.Heading(3, "Linux")
	b.Text("use the package manager\n")
	b.Heading(2, "Usage")
	b.Text("run it\n")

	got := b.Sections()
	want := []Section{
		{Path: []string{}, Text: "intro\n"},
		{Path: []string{"Guide", "Setup"}, Text: "install it\n"},
		{Path: []string{"Guide", "Setup", "Linux"}, Text: "use the package manager\n"},
		{Path: []string{"Guide", "Usage"}, Text: "run it\n"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("sections = %#v\nwant %#v", got, want)
	}
}

func TestBuilderKeepsDeepHeadingsAsText(t *testing.T) {
	var b Builder
	b.Heading(1, "Top")
	b.Heading(MaxLevel+1, "Detail")
	b.Text("\nbody\n")
	got := b.Sections()
	if len(got) != 1 || got[0].Text != "Detail\nbody\n" {
		t.Fatalf("sections = %#v", got)
	}
}

func TestPagesLabelsAndDropsEmpty(t *testing.T) {
	pages := Pages([]Section{
		{Path: []string{"A", "B"}, Text: "  text  "},
		{Path: []string{"C"}, Text: "   "},
		{Text: "no heading"},
	})
	if len(pages) != 2 {
		t.Fatalf("got %d pages, want 2", len(pages))
	}
	if pages[0].Label != "A › B" || pages[0].Text != "text" || pages[0].Number != 0 {
		t.Fatalf("page 0 = %+v", pages[0])
	}
	if pages[1].Label != "" {
		t.Fatalf("page 1 label = %q, want empty", pages[1].Label)
	}
}
