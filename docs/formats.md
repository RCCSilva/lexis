# Formats

| Format | Extensions | Split into | Citation |
|---|---|---|---|
| PDF | `.pdf` | PDF pages | `manual, pages 34-35, chunk 42` |
| Markdown | `.md`, `.markdown` | sections under `#` to `####` | `guide, section Setup › Install, chunk 3` |
| HTML | `.html`, `.htm` | sections under `h1` to `h4` | `faq, section FAQ › Licensing, chunk 0` |
| Plain text | `.txt`, `.text` | one section | `notes, chunk 0` |

A chunk never crosses a section boundary. A document with no headings is one section.

## PDF

- Page numbers are the PDF's own (the 34th sheet), not the number printed on the page.
- Scanned PDFs yield no text: there is no OCR.
- Needs `pdftotext`.

## Markdown

- Front matter is skipped.
- `#` lines inside code fences are not headings.
- Images are dropped. Links keep their text.

## HTML

- Reads `<main>`, or `<article>` if there is no `main`, or else the whole document.
- Ignores `nav`, `footer`, `aside`, `script`, `style` and similar.
- Pages built by JavaScript are not supported. Save the rendered HTML, or use the Markdown source.

## Directories

`lexis ingest <dir>` reads the supported files directly inside the directory. It does not recurse.
