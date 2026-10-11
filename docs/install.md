# Install

You need three things: Go to build Lexis, Ollama to embed text, and `pdftotext` to read PDFs.

## 1. Go

Go 1.27 or newer: https://go.dev/dl

```bash
go version
```

## 2. Ollama

Ollama runs the embedding model locally. Lexis talks to it at `http://localhost:11434`.

**Linux** (including WSL):

```bash
curl -fsSL https://ollama.com/install.sh | sh
```

**macOS / Windows:** download the installer from https://ollama.com/download and run it. The app starts the server for you.

**Docker** (any OS):

```bash
docker run -d --name ollama -p 11434:11434 -v ollama:/root/.ollama ollama/ollama
```

### Start the server

On Linux the installer registers a service that starts automatically. If you have no service manager (common on WSL), start it yourself and leave it running:

```bash
ollama serve
```

Check that it answers:

```bash
curl http://localhost:11434
# Ollama is running
```

### Pull the embedding model

```bash
ollama pull nomic-embed-text      # with Docker: docker exec ollama ollama pull nomic-embed-text
ollama list                       # nomic-embed-text should be listed
```

The first pull downloads about 270 MB. To use another model, see [configuration](configuration.md).

## 3. pdftotext

Only needed to ingest PDFs. It is part of poppler-utils.

```bash
sudo apt install poppler-utils     # Debian / Ubuntu
brew install poppler               # macOS
```

```bash
pdftotext -v
```

## Build Lexis

```bash
go build -o lexis ./cmd/lexis/
./lexis list
```

If `list` runs without an error, Lexis can reach its store. Ollama is first needed when you ingest or search.

## Troubleshooting

| Symptom | Cause and fix |
|---|---|
| Connection refused on ingest or search | Ollama is not running. Run `ollama serve`. |
| Error saying the model is not found | Run `ollama pull nomic-embed-text`. |
| Error from `pdftotext` or "not found" for it | Install poppler-utils, and check the file is a valid PDF. |
| "no text extracted" for a PDF | The PDF is scanned images. There is no OCR step. |
