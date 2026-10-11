package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config holds all runtime configuration for lexis.
type Config struct {
	Store    StoreConfig    `yaml:"store"`
	Embedder EmbedderConfig `yaml:"embedder"`
	Chunker  ChunkerConfig  `yaml:"chunker"`
	Reranker RerankerConfig `yaml:"reranker"`
	MCP      MCPConfig      `yaml:"mcp"`
}

type StoreConfig struct {
	Provider string `yaml:"provider"` // "chromem"
	Path     string `yaml:"path"`
}

type EmbedderConfig struct {
	Provider string `yaml:"provider"` // "ollama"
	Model    string `yaml:"model"`
	URL      string `yaml:"url"`
}

type ChunkerConfig struct {
	Strategy string `yaml:"strategy"` // "word"
	Size     int    `yaml:"size"`
	Overlap  int    `yaml:"overlap"`
	MaxChars int    `yaml:"max_chars"` // 0 = no limit
}

type RerankerConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Provider string `yaml:"provider"`
}

type MCPConfig struct {
	TopK                int  `yaml:"top_k"`
	CitationEnforcement bool `yaml:"citation_enforcement"`
}

// DefaultConfigDir returns ~/.lexis.
func DefaultConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".lexis"), nil
}

// Load reads ~/.lexis/config.yaml, returning safe defaults if the file doesn't exist.
func Load() (*Config, error) {
	dir, err := DefaultConfigDir()
	if err != nil {
		return nil, fmt.Errorf("resolving config dir: %w", err)
	}
	return LoadFrom(filepath.Join(dir, "config.yaml"))
}

// LoadFrom reads config from an explicit path.
func LoadFrom(path string) (*Config, error) {
	cfg := defaults()

	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return cfg, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	// Resolve ~ in store path
	if len(cfg.Store.Path) >= 2 && cfg.Store.Path[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		cfg.Store.Path = filepath.Join(home, cfg.Store.Path[2:])
	}

	return cfg, nil
}

func defaults() *Config {
	home, _ := os.UserHomeDir()
	return &Config{
		Store: StoreConfig{
			Provider: "chromem",
			Path:     filepath.Join(home, ".lexis", "chromem-db"),
		},
		Embedder: EmbedderConfig{
			Provider: "ollama",
			Model:    "nomic-embed-text",
			URL:      "",
		},
		Chunker: ChunkerConfig{
			Strategy: "word",
			Size:     800,
			Overlap:  125,
			MaxChars: 3200,
		},
		Reranker: RerankerConfig{
			Enabled: false,
		},
		MCP: MCPConfig{
			TopK:                5,
			CitationEnforcement: true,
		},
	}
}
