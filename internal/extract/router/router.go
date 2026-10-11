// Package router picks an extractor from a file's extension.
package router

import (
	"context"
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/rafael/lexis/internal/extract"
)

// Router implements extract.Extractor by dispatching on file extension.
type Router struct {
	byExt map[string]extract.Extractor
}

func New() *Router {
	return &Router{byExt: make(map[string]extract.Extractor)}
}

// Handle registers e for the given extensions (with leading dot, any case).
func (r *Router) Handle(e extract.Extractor, exts ...string) *Router {
	for _, x := range exts {
		r.byExt[strings.ToLower(x)] = e
	}
	return r
}

// Supports reports whether a file's extension has a registered extractor.
func (r *Router) Supports(path string) bool {
	_, ok := r.byExt[strings.ToLower(filepath.Ext(path))]
	return ok
}

// Extensions lists the registered extensions, sorted.
func (r *Router) Extensions() []string {
	exts := make([]string, 0, len(r.byExt))
	for x := range r.byExt {
		exts = append(exts, x)
	}
	slices.Sort(exts)
	return exts
}

func (r *Router) Extract(ctx context.Context, path string) ([]extract.Page, error) {
	e, ok := r.byExt[strings.ToLower(filepath.Ext(path))]
	if !ok {
		return nil, fmt.Errorf("unsupported file type %q (supported: %s)",
			filepath.Ext(path), strings.Join(r.Extensions(), ", "))
	}
	return e.Extract(ctx, path)
}
