package schema

// RecipeCandidate matches compiler/src/infer/recipes.ts RecipeCandidate type.
type RecipeCandidate struct {
	ID                    string                      `json:"id"`
	Kind                  string                      `json:"kind"`     // logo-cloud, feature-grid, card-grid, product-grid, gallery-showcase, cta-band
	Confidence            float64                     `json:"confidence"`
	Risk                  string                      `json:"risk"`        // low, medium, high
	RootCid               string                      `json:"rootCid"`
	RootTag               string                      `json:"rootTag"`
	ItemParentCid         string                      `json:"itemParentCid,omitempty"`
	SectionID             string                      `json:"sectionId,omitempty"`
	SectionRole           string                      `json:"sectionRole,omitempty"`
	ComponentName         string                      `json:"componentName"`
	DataModel             string                      `json:"dataModel,omitempty"`
	ItemCount             int                         `json:"itemCount,omitempty"`
	RepeatedItems         []RecipeRepeatedItem        `json:"repeatedItems,omitempty"`
	ResponsiveRegimes     []RecipeResponsiveRegime    `json:"responsiveRegimes"`
	SourceHints           []string                    `json:"sourceHints"`
	Signals               []string                    `json:"signals"`
	EmissionStatus        string                      `json:"emissionStatus"` // report-only
	FallbackReason        string                      `json:"fallbackReason"`
}

type RecipeRepeatedItem struct {
	Cid         string `json:"cid"`
	Tag         string `json:"tag"`
	TextSample  string `json:"textSample,omitempty"`
	MediaCount  int    `json:"mediaCount"`
	HeadingCount int   `json:"headingCount"`
	BBox        BBox   `json:"bbox"`
}

type RecipeResponsiveRegime struct {
	Viewport     int     `json:"viewport"`
	Layout       string  `json:"layout"` // grid, flex, stack, block, absolute, mixed
	RootBox      BBox    `json:"rootBox"`
	VisibleItems int     `json:"visibleItems,omitempty"`
	Columns      int     `json:"columns,omitempty"`
	Rows         int     `json:"rows,omitempty"`
	GapX         float64 `json:"gapX,omitempty"`
	GapY         float64 `json:"gapY,omitempty"`
}

type RecipeReport struct {
	Version           int                       `json:"version"`
	SourceUrl         string                    `json:"sourceUrl"`
	CanonicalViewport int                       `json:"canonicalViewport"`
	Viewports         []int                     `json:"viewports"`
	SampledViewports  []int                     `json:"sampledViewports"`
	Summary           RecipeSummary             `json:"summary"`
	Candidates        []RecipeCandidate         `json:"candidates"`
}

type RecipeSummary struct {
	TotalCandidates      int      `json:"totalCandidates"`
	HighConfidence       int      `json:"highConfidence"`
	ByKind               map[string]int `json:"byKind"`
	TemplateReadyKinds   []string `json:"templateReadyKinds"`
	CatalogPatterns      []string `json:"catalogPatterns,omitempty"`
}

// NOTE: Recipes are detected by analyzing the IR for repeated item patterns.
// This is a placeholder; full implementation requires extensive pattern matching logic.
func DetectRecipes(root *IRNode, sections []Section, cw int) RecipeReport {
	// Placeholder: return empty report
	return RecipeReport{
		Version:           1,
		SourceUrl:         "",
		CanonicalViewport: cw,
		Viewports:         []int{cw},
		SampledViewports:  []int{cw},
		Summary: RecipeSummary{
			TotalCandidates:    0,
			HighConfidence:     0,
			ByKind:             make(map[string]int),
			TemplateReadyKinds: []string{"logo-cloud", "feature-grid", "card-grid", "product-grid", "cta-band"},
		},
		Candidates: []RecipeCandidate{},
	}
}
