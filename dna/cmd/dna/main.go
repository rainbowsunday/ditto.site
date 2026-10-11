package main

import (
    "flag"
    "fmt"
    "log"
    "os"
    "path/filepath"
    "strings"
    "time"

    "github.com/rainbowsunday/ditto.site/dna/internal/schema"
)

func main() {
    irPath := flag.String("ir", "", "Path to the hosted IR snapshot (.clone/ir.json)")
    outDir := flag.String("out", "./output", "Output root directory")
    flag.Parse()

    if *irPath == "" {
        flag.Usage()
        log.Fatal("error: --ir is required")
    }

    b, err := os.ReadFile(*irPath)
    if err != nil {
        log.Fatalf("error: read IR file: %v", err)
    }

    ir, err := schema.LoadIR(b)
    if err != nil {
        log.Fatalf("error: parse IR: %v", err)
    }

    host := strings.TrimPrefix(ir.Doc.SourceURL, "https://")
    host = strings.TrimPrefix(host, "http://")
    host = strings.Split(host, "/")[0]
    host = strings.TrimPrefix(host, "www.")
    siteName := sanitize(host)

    generated := filepath.Join(*outDir, siteName, ".clone", "generated")

    sections := schema.DetectSections(ir)
    tokens := schema.ExtractTokens(ir)
    assets := schema.ExtractAssets(ir.Root)
    fonts := schema.ExtractFonts(ir)
    seo := schema.ExtractSEO(ir)
    recipes := schema.DetectRecipes(ir, sections)
    interactions := schema.DetectInteractionRecipes(ir)
    signals := schema.DetectSignals(sections)

    manifest := map[string]interface{}{
        "schemaVersion":     1,
        "compilerVersion":   "0.1.0-go",
        "sourceUrl":         ir.Doc.SourceURL,
        "canonicalViewport": ir.Doc.CanonicalViewport,
        "clonedAt":          time.Now().UTC().Format(time.RFC3339),
        "sections":          len(sections),
        "recipes":           len(recipes.Candidates),
        "assets":            len(assets),
    }

    if err := schema.WriteGenerated(generated, sections, tokens, recipes, interactions, assets, fonts, seo, signals, manifest); err != nil {
        log.Fatalf("error: write generated output: %v", err)
    }

    fmt.Printf("✓ DNA generated at %s\n", generated)
}

func sanitize(s string) string {
    s = strings.ToLower(s)
    s = strings.Map(func(r rune) rune {
        if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '.' {
            return r
        }
        return '-'
    }, s)
    s = strings.Trim(s, ".-")
    if s == "" {
        return "site"
    }
    return s
}
