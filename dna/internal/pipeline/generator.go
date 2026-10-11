package pipeline

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rainbowsunday/ditto.site/dna/internal/ingest"
	"github.com/rainbowsunday/ditto.site/dna/internal/schema"
)

// Generator orchestrates DNA extraction to .clone/generated/.
type Generator struct {
	CloneDir  string
	SourceDir string
	SourceURL string
}

func NewGenerator(cloneDir, sourceDir, sourceURL string) *Generator {
	return &Generator{
		CloneDir:  cloneDir,
		SourceDir: sourceDir,
		SourceURL: sourceURL,
	}
}

// Run executes the full DNA generation pipeline and writes to .clone/generated/.
func (g *Generator) Run() error {
	// Load canonical snapshot
	capDir := filepath.Join(g.SourceDir, "capture")
	capture, canonical, err := ingest.LoadCanonical(capDir)
	if err != nil {
		return fmt.Errorf("load canonical snapshot: %w", err)
	}

	// Extract IR metadata
	var root *schema.IRNode
	if rootRaw, ok := capture.Root["root"]; ok {
		// Note: in practice, the entire root structure is already parsed as IRNode from JSON
		if m, ok := rootRaw.(map[string]interface{}); ok {
			root = unmarshalIRNode(m)
		}
	}

	if root == nil {
		return fmt.Errorf("root node not found in capture")
	}

	// Build per-viewport metadata
	perViewport := make(map[int]schema.ViewportInfo)
	perViewport[canonical] = schema.ViewportInfo{
		ScrollHeight: capture.Doc.ScrollHeight,
		ScrollWidth:  capture.Doc.ScrollWidth,
		HtmlBg:       capture.Doc.HtmlBg,
		BodyBg:       capture.Doc.BodyBg,
		BodyColor:    capture.Doc.BodyColor,
		BodyFont:     capture.Doc.BodyFont,
	}

	viewports := []int{canonical} // Minimal set for now

	// Extract DNA elements
	sections := schema.DetectSections(root, canonical, perViewport)
	tokens := schema.ExtractTokens(root, canonical, perViewport, viewports)
	assets := schema.ExtractAssets(root)
	recipes := schema.DetectRecipes(root, sections, canonical)
	seo := schema.ExtractSEO(capture.Root)
	signals := schema.DetectSignals(sections)

	// Build output directory
	generatedDir := filepath.Join(g.CloneDir, "generated")
	if err := os.MkdirAll(generatedDir, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", generatedDir, err)
	}

	// Create manifest
	manifest := map[string]interface{}{
		"schemaVersion":   1,
		"compilerVersion": "0.1.0-go",
		"sourceUrl":       g.SourceURL,
		"canonicalViewport": canonical,
		"clonedAt":        time.Now().UTC().Format(time.RFC3339),
		"sections":        len(sections),
		"recipes":         len(recipes.Candidates),
		"assets":          len(assets),
	}

	// Write all JSON outputs
	outputs := map[string]interface{}{
		"sections.json":             sections,
		"tokens.json":               tokens,
		"recipes.json":              recipes,
		"interaction-recipes.json":  map[string]interface{}{"version": 1, "patterns": []interface{}{}}, // Placeholder
		"assets.json":               assets,
		"fonts.json":                []interface{}{}, // Placeholder
		"seo.json":                  seo,
		"manifest.json":             manifest,
		"section-signals.json":      signals,
	}

	for filename, data := range outputs {
		path := filepath.Join(generatedDir, filename)
		b, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal %s: %w", filename, err)
		}

		if err := os.WriteFile(path, append(b, '\n'), 0o644); err != nil {
			return fmt.Errorf("write %s: %w", path, err)
		}
	}

	return nil
}

// Helper to unmarshal raw JSON map to IRNode (simplified).
func unmarshalIRNode(m map[string]interface{}) *schema.IRNode {
	b, _ := json.Marshal(m)
	var node schema.IRNode
	json.Unmarshal(b, &node)
	return &node
}
