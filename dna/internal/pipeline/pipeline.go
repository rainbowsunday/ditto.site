package pipeline

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    "strings"

    "github.com/rainbowsunday/ditto.site/dna/internal/analysis"
    "github.com/rainbowsunday/ditto.site/dna/internal/ingest"
)

// Generator orchestrates DNA extraction from a site capture.
type Generator struct {
    cloneDir string
    sourceDir string
    sourceURL string
}

// NewGenerator creates a generator for a source directory and output path.
func NewGenerator(cloneDir string, sourceURL string, sourceDir string) *Generator {
    return &Generator{cloneDir: cloneDir, sourceURL: sourceURL, sourceDir: sourceDir}
}

// Run generates the DNA output set at .clone.
func (g *Generator) Run() error {
    snap, err := ingest.LoadSnapshot(g.sourceDir)
    if err != nil {
        return err
    }
    if snap == nil {
        return fmt.Errorf("capture snapshot is empty")
    }

    root := snap.Root
    doc := snap.Doc
    sections := analysis.DetectSections(root)
    tokens := analysis.ExtractTokens(root)
    assets := analysis.BuildAssets(root)
    fonts := analysis.BuildFonts(doc, root)
    seo := analysis.BuildSEO(doc)

    manifest := map[string]any{
        "version": "go-dna-v1",
        "sourceUrl": g.sourceURL,
        "sourceDir": g.sourceDir,
        "clonedAt": nowISO(),
        "sections": len(sections),
        "tokens": map[string]any{
            "colors": len(tokens.Colors),
            "fonts": len(tokens.Fonts),
            "spacing": len(tokens.Spacing),
        },
        "assets": len(assets),
        "fonts": len(fonts),
        "seo": seo,
    }

    files := map[string]any{
        "sections.json": sections,
        "tokens.json": tokens,
        "assets.json": assets,
        "fonts.json": fonts,
        "seo.json": seo,
        "manifest.json": manifest,
    }

    for name, value := range files {
        b, err := json.MarshalIndent(value, "", "  ")
        if err != nil {
            return fmt.Errorf("marshal %s: %w", name, err)
        }
        if err := os.WriteFile(filepath.Join(g.cloneDir, name), append(b, '\n'), 0o644); err != nil {
            return fmt.Errorf("write %s: %w", name, err)
        }
    }

    // Write a lightweight summary for quick inspection.
    summary := map[string]any{
        "site": strings.TrimPrefix(filepath.Base(g.cloneDir), ".clone"),
        "sourceUrl": g.sourceURL,
        "totalSections": len(sections),
        "totalAssets": len(assets),
        "totalFonts": len(fonts),
        "title": seo.Title,
    }
    if b, err := json.MarshalIndent(summary, "", "  "); err == nil {
        _ = os.WriteFile(filepath.Join(g.cloneDir, "dna-summary.json"), append(b, '\n'), 0o644)
    }

    return nil
}

func nowISO() string {
    return "2026-10-11T00:00:00Z"
}
