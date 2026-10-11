package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/RCCSilva/lexis/config"
	"github.com/RCCSilva/lexis/internal/chunk/word"
	"github.com/RCCSilva/lexis/internal/embed/ollama"
	"github.com/RCCSilva/lexis/internal/extract/html"
	"github.com/RCCSilva/lexis/internal/extract/markdown"
	"github.com/RCCSilva/lexis/internal/extract/pdftotext"
	"github.com/RCCSilva/lexis/internal/extract/router"
	"github.com/RCCSilva/lexis/internal/extract/text"
	"github.com/RCCSilva/lexis/internal/library"
	"github.com/RCCSilva/lexis/internal/mcp"
	chromemstore "github.com/RCCSilva/lexis/internal/store/chromem"
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
		query, opts, err := parseSearchArgs(os.Args[2:])
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			fmt.Fprintln(os.Stderr, "usage: lexis search [--source <name>] [-n <count>] \"<query>\"")
			os.Exit(1)
		}
		if err := cmdSearch(ctx, query, opts); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	case "chunk":
		if len(os.Args) < 4 {
			fmt.Fprintln(os.Stderr, "usage: lexis chunk <source> <chunk-number> [neighbors]")
			os.Exit(1)
		}
		if err := cmdChunk(ctx, os.Args[2:]); err != nil {
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
	fmt.Println("  lexis ingest <file-or-directory>       extract, embed and store documents")
	fmt.Println("                                         (" + strings.Join(extractors().Extensions(), " ") + ")")
	fmt.Println("  lexis search [--source S] [-n N] \"<q>\"  search ingested documents")
	fmt.Println("  lexis chunk <source> <n> [neighbors]   print one chunk and its neighbors")
	fmt.Println("  lexis list                             list all ingested sources")
	fmt.Println("  lexis remove <source>                  remove a source from the store")
	fmt.Println("  lexis mcp                              start MCP server (stdio)")
}

// extractors maps file extensions to the extractor that reads them.
func extractors() *router.Router {
	return router.New().
		Handle(pdftotext.New(), ".pdf").
		Handle(markdown.New(), ".md", ".markdown").
		Handle(html.New(), ".html", ".htm").
		Handle(text.New(), ".txt", ".text")
}

func buildLibrary(cfg *config.Config) (*library.Library, error) {
	ext := extractors()
	chk := word.New(cfg.Chunker.Size, cfg.Chunker.Overlap, cfg.Chunker.MaxChars)
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

	files, err := collectFiles(path, extractors())
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return fmt.Errorf("no supported files found at %s", path)
	}
	fmt.Printf("Found %d file(s)\n", len(files))

	for _, p := range files {
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

// parseSearchArgs splits the search arguments into the query and its options.
func parseSearchArgs(args []string) (string, library.SearchOptions, error) {
	var opts library.SearchOptions
	var words []string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--source", "-s", "-n", "--top-k":
			if i+1 >= len(args) {
				return "", opts, fmt.Errorf("%s needs a value", args[i])
			}
			i++
			if args[i-1] == "--source" || args[i-1] == "-s" {
				opts.Source = args[i]
				continue
			}
			n, err := strconv.Atoi(args[i])
			if err != nil || n < 1 {
				return "", opts, fmt.Errorf("%s needs a positive number, got %q", args[i-1], args[i])
			}
			opts.TopK = n
		default:
			words = append(words, args[i])
		}
	}
	if len(words) == 0 {
		return "", opts, fmt.Errorf("missing query")
	}
	return strings.Join(words, " "), opts, nil
}

func cmdSearch(ctx context.Context, query string, opts library.SearchOptions) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	lib, err := buildLibrary(cfg)
	if err != nil {
		return err
	}
	if opts.TopK == 0 {
		opts.TopK = cfg.MCP.TopK
	}

	hits, err := lib.Search(ctx, query, opts)
	if err != nil {
		return err
	}
	if len(hits) == 0 {
		fmt.Println("No relevant passages found.")
		return nil
	}
	for i, h := range hits {
		fmt.Printf("[%d] %s (%s)\n%s\n\n", i+1, library.Locator(h.Document), h.Match, h.Text)
	}
	return nil
}

func cmdChunk(ctx context.Context, args []string) error {
	index, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("chunk number must be an integer, got %q", args[1])
	}
	neighbors := 0
	if len(args) > 2 {
		if neighbors, err = strconv.Atoi(args[2]); err != nil {
			return fmt.Errorf("neighbors must be an integer, got %q", args[2])
		}
	}
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	lib, err := buildLibrary(cfg)
	if err != nil {
		return err
	}
	docs, err := lib.Chunk(ctx, args[0], index, neighbors)
	if err != nil {
		return err
	}
	for _, d := range docs {
		fmt.Printf("%s\n%s\n\n", library.Locator(d), d.Text)
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

// collectFiles returns path itself if it is a supported file, or the
// supported files directly inside it if it is a directory.
func collectFiles(path string, r *router.Router) ([]string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		if r.Supports(path) {
			return []string{path}, nil
		}
		return nil, fmt.Errorf("%s: unsupported file type (supported: %s)",
			path, strings.Join(r.Extensions(), ", "))
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, e := range entries {
		if !e.IsDir() && r.Supports(e.Name()) {
			files = append(files, filepath.Join(path, e.Name()))
		}
	}
	return files, nil
}
