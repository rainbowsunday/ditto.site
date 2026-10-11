# Go DNA generator

This directory contains a compact Go implementation of the site DNA layer used by the ditto site compiler.

It is intentionally focused on the data-generation phase that sits under the current Playwright capture pipeline.

## What it does

Given a source capture directory or a URL, the tool produces a `.clone/` directory containing:

- `manifest.json`
- `sections.json`
- `tokens.json`
- `assets.json`
- `fonts.json`
- `seo.json`
- `dna-summary.json`

The goal is to encapsulate the DNA extraction logic cleanly, while leaving the browser capture flow in the existing TypeScript implementation.

## Example

```bash
go run ./dna/cmd/dna --url https://example.com --out ./output
```

This creates output like:

```text
output/
  example.com/
    .clone/
      manifest.json
      sections.json
      tokens.json
      assets.json
      fonts.json
      seo.json
```

## Why this structure

- `internal/ingest`: source capture loading
- `internal/analysis`: domain inference (sections, tokens, SEO, assets, fonts)
- `internal/pipeline`: orchestration and output writing
- `cmd/dna`: CLI entry point

This keeps the workflow extensible while avoiding a single monolithic implementation.
