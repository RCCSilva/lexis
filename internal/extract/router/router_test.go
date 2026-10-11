package router

import (
	"context"
	"strings"
	"testing"

	"github.com/RCCSilva/lexis/internal/extract"
)

type fake string

func (f fake) Extract(context.Context, string) ([]extract.Page, error) {
	return []extract.Page{{Text: string(f)}}, nil
}

func TestRouterDispatchesByExtensionIgnoringCase(t *testing.T) {
	r := New().Handle(fake("a"), ".aa").Handle(fake("b"), ".BB")

	for path, want := range map[string]string{"x.aa": "a", "dir/x.AA": "a", "x.bb": "b"} {
		pages, err := r.Extract(context.Background(), path)
		if err != nil || pages[0].Text != want {
			t.Errorf("%s: pages=%+v err=%v, want %q", path, pages, err, want)
		}
		if !r.Supports(path) {
			t.Errorf("%s should be supported", path)
		}
	}

	if r.Supports("x.cc") {
		t.Error("x.cc should not be supported")
	}
	_, err := r.Extract(context.Background(), "x.cc")
	if err == nil || !strings.Contains(err.Error(), ".aa, .bb") {
		t.Fatalf("err = %v, want list of supported extensions", err)
	}
}
