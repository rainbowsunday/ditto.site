package schema

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
)

func WriteGenerated(outDir string, sections []Section, tokens Tokens, recipes RecipeReport, interactions InteractionRecipeReport, assets []Asset, fonts []Font, seo SEO, signals SectionSignals, manifest map[string]interface{}) error {
    if err := os.MkdirAll(outDir, 0o755); err != nil {
        return fmt.Errorf("mkdir %s: %w", outDir, err)
    }

    files := map[string]interface{}{
        "sections.json":             sections,
        "tokens.json":               tokens,
        "recipes.json":              recipes,
        "interaction-recipes.json":  interactions,
        "assets.json":               assets,
        "fonts.json":                fonts,
        "seo.json":                  seo,
        "section-signals.json":      signals,
        "manifest.json":             manifest,
    }

    for name, data := range files {
        b, err := json.MarshalIndent(data, "", "  ")
        if err != nil {
            return fmt.Errorf("marshal %s: %w", name, err)
        }
        path := filepath.Join(outDir, name)
        if err := os.WriteFile(path, append(b, '\n'), 0o644); err != nil {
            return fmt.Errorf("write %s: %w", path, err)
        }
    }
    return nil
}
