package schema

// SectionSignals is the mining-signal inventory.
type SectionSignals struct {
    Footer  []string `json:"footer,omitempty"`
    Search  []string `json:"search,omitempty"`
    Marquee []string `json:"marquee,omitempty"`
    Bleed   []string `json:"bleed,omitempty"`
}

func DetectSignals(sections []Section) SectionSignals {
    out := SectionSignals{
        Footer:  []string{},
        Search:  []string{},
        Marquee: []string{},
        Bleed:   []string{},
    }
    for _, s := range sections {
        if s.Role == "footer" {
            out.Footer = append(out.Footer, s.ID)
        }
    }
    return out
}
