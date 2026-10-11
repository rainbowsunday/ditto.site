# DNA — IR → .clone/generated

Phase 2 minimal implementation.

This module accepts a hosted IR snapshot from the SBR side and produces output under `.clone/generated/` for downstream consumption.

It is intentionally not a Playwright capture pipeline, and it does not claim TS parity.

## Input contract

The host provides a normalized IR JSON file, e.g.:

```bash
--ir /path/to/.clone/ir.json
```

The file is expected to contain the TS-style IR document structure with:
- `doc.sourceUrl`
- `doc.title`
- `doc.lang`
- `doc.canonicalViewport`
- `doc.perViewport`
- `root` node tree

## Output contract

The tool writes to:

```text
<out>/<site>/.clone/generated/
```

with, at minimum:
- `sections.json`
- `tokens.json`
- `assets.json`
- `seo.json`
- `manifest.json`
- `recipes.json`
- `interaction-recipes.json`
- `fonts.json`
- `section-signals.json`

## Build

```bash
cd dna
go build ./cmd/dna
go test ./...
```

## Run

```bash
./bin/dna --ir ./path/to/ir.json --out ./output
```

## Current status

Implemented:
- IR decode
- section inference
- token extraction
- asset extraction
- SEO extraction
- file writer

Placeholder-only:
- recipe detection
- interaction detection
- font extraction beyond a minimal body-font pass
- advanced mining signals (search / marquee / bleed)

This is intentionally a schema-first, buildable foundation for Phase 3 golden validation and downstream integration.
