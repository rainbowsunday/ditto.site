# Go DNA Generator - Strict TS Compiler Compatibility

**Wire-compatible with `compiler/src/generate` outputs.**

This is a modular Go implementation of the ditto.site DNA extraction layer that produces **identical schema and algorithm output** to the existing TypeScript compiler. It is NOT a rewrite—it is an algorithm port of the canonical extraction pipeline.

## Contract

### Input
- Playwright `source/capture/dom-*.json` snapshots (per-viewport DOM tree + computed styles)
- Optional interaction metadata (`capture/interaction.json`)

### Output
- **Directory**: `<out>/<site>/.clone/generated/`
- **Files** (v1 schema, byte-compatible with TS):
  - `sections.json` — Sections with role/order/bboxByVp
  - `tokens.json` — Design tokens (colors/fonts/spacing/...)
  - `recipes.json` — Layout recipe candidates with confidence/risk
  - `interaction-recipes.json` — Interactive pattern inventory
  - `assets.json` — Media resource references ("reference-only" contract)
  - `fonts.json` — Font family declarations
  - `seo.json` — SEO metadata
  - `manifest.json` — Extraction summary
  - `section-signals.json` — Non-zero signals for mining (search/footer/marquee/bleed)

## Guarantees

✅ **Algorithm fidelity**: LCS-based DOM alignment, section detection, token extraction, recipe detection all migrated from TS source.  
✅ **Schema fidelity**: JSON output matches `compiler/src/infer/*.ts` and `compiler/src/generate/*.ts` type definitions exactly.  
✅ **Determinism**: Byte-identical output when given the same capture input (via canonical serialization).  
✅ **No app generation**: This layer does NOT produce React/Next.js code. That remains in the TS pipeline.

## Building

```bash
cd dna
go build -o ../bin/dna ./cmd/dna
```

## Usage

### From existing source capture

```bash
./bin/dna --source /path/to/source --url https://example.com --out ./output
```

### With auto-detection

```bash
./bin/dna --url https://example.com --out ./output
# (Looks for source in standard layout: runs/example.com/latest/source, etc.)
```

## Verification

After generation, the `.clone/generated/` artifacts are consumed by:

1. **AdaptDitto** (service layer): Loads IR + sections to understand page structure
2. **Mining** (Notion/Google): Extracts search/footer/marquee signals from `section-signals.json`
3. **Recipes** (code generation): Uses recipe candidates for component templating

No breaking changes to existing compiler downstream.

## Architecture

```
dna/
├── cmd/dna/
│   └── main.go                 # CLI entry
├── internal/
│   ├── ingest/
│   │   └── loader.go           # Capture loading (dom-*.json, interaction.json)
│   ├── schema/
│   │   ├── section.go          # Section type + detection algorithm (from TS)
│   │   ├── tokens.go           # Token extraction (from TS)
│   │   ├── recipes.go          # Recipe detection (from TS)
│   │   ├── interaction.go      # Interaction recipe detection
│   │   ├── assets.go           # Asset inventory
│   │   └── seo.go              # SEO extraction
│   └── pipeline/
│       └── generator.go        # Orchestration + JSON output
└── go.mod
```

Each schema module is a direct port of the corresponding TS file, preserving:
- Variable naming and constant thresholds
- Confidence scoring formulas
- Pattern matching heuristics
- Responsive regime detection
