package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rafael/lexis/config"
	"github.com/rafael/lexis/internal/chunk/word"
	"github.com/rafael/lexis/internal/embed/ollama"
	"github.com/rafael/lexis/internal/extract/pdftotext"
	"github.com/rafael/lexis/internal/library"
	"github.com/rafael/lexis/internal/mcp"
	chromemstore "github.com/rafael/lexis/internal/store/chromem"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	ctx := context.Background()

	switch os.Args[1] {
	case "ingest":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: lexis ingest <pdf-file-or-directory>")
			os.Exit(1)
		}
		if err := cmdIngest(ctx, os.Args[2]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	case "search":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: lexis search \"<query>\"")
			os.Exit(1)
		}
		query := strings.Join(os.Args[2:], " ")
		if err := cmdSearch(ctx, query); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	case "list":
		if err := cmdList(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	case "remove":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: lexis remove <source>")
			os.Exit(1)
		}
		if err := cmdRemove(ctx, os.Args[2]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	case "mcp":
		if err := cmdMCP(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "mcp server error: %v\n", err)
			os.Exit(1)
		}
	default:
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Println("usage:")
	fmt.Println("  lexis ingest <pdf-file-or-directory>   extract, embed and store PDFs")
	fmt.Println("  lexis search \"<query>\"                 search ingested documents")
	fmt.Println("  lexis list                             list all ingested sources")
	fmt.Println("  lexis remove <source>                  remove a source from the store")
	fmt.Println("  lexis mcp                              start MCP server (stdio)")
}

func buildLibrary(cfg *config.Config) (*library.Library, error) {
	ext := pdftotext.New()
	chk := word.New(cfg.Chunker.Size, cfg.Chunker.Overlap)
	emb := ollama.New(cfg.Embedder.Model, cfg.Embedder.URL)
	st, err := chromemstore.New(cfg.Store.Path, cfg.Embedder.Model, cfg.Embedder.URL)
	if err != nil {
		return nil, fmt.Errorf("opening store: %w", err)
	}
	return library.New(ext, chk, emb, st, nil), nil
}

func cmdIngest(ctx context.Context, path string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	lib, err := buildLibrary(cfg)
	if err != nil {
		return err
	}

	pdfs, err := collectPDFs(path)
	if err != nil {
		return err
	}
	if len(pdfs) == 0 {
		return fmt.Errorf("no PDF files found at %s", path)
	}
	fmt.Printf("Found %d PDF(s)\n", len(pdfs))

	for _, p := range pdfs {
		fmt.Printf("Ingesting %s ...\n", filepath.Base(p))
		if err := lib.Ingest(ctx, p); err != nil {
			fmt.Fprintf(os.Stderr, "  skip: %v\n", err)
			continue
		}
		fmt.Printf("  done\n")
	}
	fmt.Println("Ingestion complete.")
	return nil
}

func cmdSearch(ctx context.Context, query string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	lib, err := buildLibrary(cfg)
	if err != nil {
		return err
	}

	results, err := lib.Search(ctx, query, cfg.MCP.TopK)
	if err != nil {
		return err
	}
	if len(results) == 0 {
		fmt.Println("No relevant passages found.")
		return nil
	}
	for i, r := range results {
		fmt.Printf("[%d] Source: %s (similarity: %.2f)\n%s\n\n", i+1, r.Metadata["source"], r.Score, r.Text)
	}
	return nil
}

func cmdList(ctx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	lib, err := buildLibrary(cfg)
	if err != nil {
		return err
	}
	sources, err := lib.Sources(ctx)
	if err != nil {
		return err
	}
	if len(sources) == 0 {
		fmt.Println("No sources ingested yet.")
		return nil
	}
	for _, s := range sources {
		fmt.Println(s)
	}
	return nil
}

func cmdRemove(ctx context.Context, source string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	lib, err := buildLibrary(cfg)
	if err != nil {
		return err
	}
	if err := lib.Delete(ctx, source); err != nil {
		return err
	}
	fmt.Printf("Removed %q.\n", source)
	return nil
}

func cmdMCP(ctx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	lib, err := buildLibrary(cfg)
	if err != nil {
		return err
	}
	return mcp.Serve(lib, cfg.MCP.TopK, cfg.MCP.CitationEnforcement)
}

func collectPDFs(path string) ([]string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		if strings.ToLower(filepath.Ext(path)) == ".pdf" {
			return []string{path}, nil
		}
		return nil, fmt.Errorf("%s is not a PDF file", path)
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	var pdfs []string
	for _, e := range entries {
		if !e.IsDir() && strings.ToLower(filepath.Ext(e.Name())) == ".pdf" {
			pdfs = append(pdfs, filepath.Join(path, e.Name()))
		}
	}
	return pdfs, nil
}
