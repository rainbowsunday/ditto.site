package schema

// RecipeCandidate is a placeholder recipe representation.
type RecipeCandidate struct {
    ID             string   `json:"id"`
    Kind           string   `json:"kind"`
    Confidence     float64  `json:"confidence"`
    Risk           string   `json:"risk"`
    RootCid        string   `json:"rootCid"`
    ComponentName  string   `json:"componentName"`
    ResponsiveRegimes []ResponsiveRegime `json:"responsiveRegimes"`
    Signals        []string `json:"signals"`
    EmissionStatus string   `json:"emissionStatus"`
    FallbackReason string   `json:"fallbackReason"`
}

type ResponsiveRegime struct {
    Viewport int    `json:"viewport"`
    Layout   string `json:"layout"`
}

type RecipeSummary struct {
    TotalCandidates int            `json:"totalCandidates"`
    HighConfidence  int            `json:"highConfidence"`
    ByKind          map[string]int `json:"byKind"`
}

type RecipeReport struct {
    Version           int               `json:"version"`
    SourceURL         string            `json:"sourceUrl"`
    CanonicalViewport int               `json:"canonicalViewport"`
    Viewports         []int             `json:"viewports"`
    Summary           RecipeSummary     `json:"summary"`
    Candidates        []RecipeCandidate `json:"candidates"`
}

func DetectRecipes(ir *IR, sections []Section) RecipeReport {
    return RecipeReport{
        Version:           1,
        SourceURL:         ir.Doc.SourceURL,
        CanonicalViewport: ir.Doc.CanonicalViewport,
        Viewports:         ir.Doc.Viewports,
        Summary: RecipeSummary{
            TotalCandidates: 0,
            HighConfidence:  0,
            ByKind:          map[string]int{},
        },
        Candidates: []RecipeCandidate{},
    }
}

type InteractionPattern struct {
    ID     string `json:"id"`
    Kind   string `json:"kind"`
    RootID string `json:"rootId"`
}

type InteractionRecipeReport struct {
    Version   int                  `json:"version"`
    SourceURL string               `json:"sourceUrl"`
    Patterns  []InteractionPattern `json:"patterns"`
}

func DetectInteractionRecipes(ir *IR) InteractionRecipeReport {
    return InteractionRecipeReport{
        Version:   1,
        SourceURL: ir.Doc.SourceURL,
        Patterns:  []InteractionPattern{},
    }
}
