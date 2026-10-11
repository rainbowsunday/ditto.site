package schema

// Font is a simple extracted font declaration.
type Font struct {
    Family string `json:"family"`
    Source string `json:"source,omitempty"`
}

func ExtractFonts(ir *IR) []Font {
    out := make([]Font, 0)
    cw := ir.Doc.CanonicalViewport
    seen := map[string]bool{}

    if font := ir.Doc.PerViewport[cw].BodyFont; font != "" && !seen[font] {
        seen[font] = true
        out = append(out, Font{Family: font, Source: "document"})
    }

    // Placeholder; real font extraction can be added later via computed styles.
    return out
}
