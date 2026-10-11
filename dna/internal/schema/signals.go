package schema

// SectionSignals is a lightweight signal inventory for mining/external processing.
// It marks which sections contain search-relevant, footer, marquee, or full-bleed patterns.
type SectionSignals struct {
	Search []string `json:"search,omitempty"` // Section IDs with search-like content
	Footer []string `json:"footer,omitempty"` // Footer-role sections
	Marquee []string `json:"marquee,omitempty"` // Sections with carousel/ticker patterns
	Bleed  []string `json:"bleed,omitempty"`   // Full-width sections
}

// DetectSignals builds section signal inventory from sections.
func DetectSignals(sections []Section) SectionSignals {
	signals := SectionSignals{
		Search:  []string{},
		Footer:  []string{},
		Marquee: []string{},
		Bleed:   []string{},
	}

	for _, s := range sections {
		if s.Role == "footer" {
			signals.Footer = append(signals.Footer, s.ID)
		}
		// Add more signal detection here as needed
	}

	return signals
}
