package schema

import (
	"math"
	"regexp"
	"sort"
	"strings"
)

// Tokens matches compiler/src/infer/tokens.ts Tokens type.
type Tokens struct {
	Colors      map[string]string `json:"colors"`
	Fonts       map[string]string `json:"fonts"`
	FontSizes   map[string]string `json:"fontSizes"`
	FontWeights map[string]string `json:"fontWeights"`
	LineHeights map[string]string `json:"lineHeights"`
	Spacing     map[string]string `json:"spacing"`
	Radii       map[string]string `json:"radii"`
	Shadows     map[string]string `json:"shadows"`
	ZIndices    map[string]string `json:"zIndices"`
	Breakpoints map[string]string `json:"breakpoints"`
}

var TRANSPARENT = map[string]bool{
	"rgba(0, 0, 0, 0)": true,
	"transparent":      true,
	"rgba(0,0,0,0)":    true,
}

// ExtractTokens extracts design tokens from IR (ported from TS).
func ExtractTokens(root *IRNode, cw int, perViewport map[int]ViewportInfo, viewports []int) Tokens {
	colors := make(map[string]int)
	fonts := make(map[string]int)
	fontSizes := make(map[string]int)
	fontWeights := make(map[string]int)
	lineHeights := make(map[string]int)
	spacing := make(map[string]int)
	radii := make(map[string]int)
	shadows := make(map[string]int)
	zIndices := make(map[string]int)

	// Traverse and collect
	visit := func(node *IRNode) {
		var visitNode func(*IRNode)
		visitNode = func(n *IRNode) {
			cs := n.ComputedByVp[cw]
			if cs != nil && n.VisibleByVp[cw] {
				if cs["color"] != "" && !TRANSPARENT[cs["color"]] {
					colors[cs["color"]]++
				}
				if cs["backgroundColor"] != "" && !TRANSPARENT[cs["backgroundColor"]] {
					colors[cs["backgroundColor"]]++
				}
				if cs["fontFamily"] != "" {
					fonts[cs["fontFamily"]]++
				}
				if cs["fontSize"] != "" {
					fontSizes[cs["fontSize"]]++
				}
				if cs["fontWeight"] != "" {
					fontWeights[cs["fontWeight"]]++
				}
				if cs["lineHeight"] != "" && cs["lineHeight"] != "normal" {
					lineHeights[cs["lineHeight"]]++
				}

				// Spacing properties
				for _, prop := range []string{"paddingTop", "paddingBottom", "paddingLeft", "paddingRight", "marginTop", "marginBottom", "gap"} {
					if v := cs[prop]; v != "" && v != "0px" && strings.HasSuffix(v, "px") {
						spacing[v]++
					}
				}

				// Border radius
				for _, prop := range []string{"borderTopLeftRadius", "borderTopRightRadius"} {
					if v := cs[prop]; v != "" && v != "0px" {
						radii[v]++
					}
				}

				if v := cs["boxShadow"]; v != "" && v != "none" {
					shadows[v]++
				}
				if v := cs["zIndex"]; v != "" && v != "auto" {
					zIndices[v]++
				}
			}

			for _, child := range n.Children {
				if irChild, ok := child.(*IRNode); ok {
					visitNode(irChild)
				}
			}
		}
		visitNode(node)
	}
	visit(root)

	pv := perViewport[cw]

	// Build token maps
	colorTokens := make(map[string]string)
	pageBg := pv.BodyBg
	if pageBg == "" || TRANSPARENT[pageBg] {
		pageBg = "rgb(255, 255, 255)"
	}
	colorTokens["--color-bg-page"] = pageBg

	// Primary text
	primaryText := pv.BodyColor
	if primaryText == "" || TRANSPARENT[primaryText] {
		// Use most frequent color
		if sorted := topSorted(colors, 1); len(sorted) > 0 {
			primaryText = sorted[0][0]
		}
	}
	if primaryText != "" {
		colorTokens["--color-text-primary"] = primaryText
	}

	usedColors := make(map[string]bool)
	usedColors[pageBg] = true
	usedColors[primaryText] = true

	ci := 1
	for _, pair := range topSorted(colors, 3) {
		if usedColors[pair[0]] {
			continue
		}
		usedColors[pair[0]] = true
		colorTokens[pad("--color-", ci)] = pair[0]
		ci++
	}

	fontTokens := make(map[string]string)
	if sorted := topSorted(fonts, 1); len(sorted) > 0 {
		fontTokens["--font-body"] = sorted[0][0]
	}
	fi := 1
	primaryFont := ""
	if sorted := topSorted(fonts, 1); len(sorted) > 0 {
		primaryFont = sorted[0][0]
	}
	for _, pair := range topSorted(fonts, 3) {
		if pair[0] == primaryFont {
			continue
		}
		fontTokens[pad("--font-", fi)] = pair[0]
		fi++
	}

	// Breakpoints
	breakpoints := make(map[string]string)
	for _, vp := range viewports {
		breakpoints[pad("--bp-", vp)] = pad("", vp) + "px"
	}

	return Tokens{
		Colors:      colorTokens,
		Fonts:       fontTokens,
		FontSizes:   numberedByValue(fontSizes, "--font-size"),
		FontWeights: numberedByValue(fontWeights, "--font-weight"),
		LineHeights: numberedByFreq(lineHeights, "--line-height"),
		Spacing:     numberedByValue(spacing, "--space"),
		Radii:       numberedByValue(radii, "--radius"),
		Shadows:     numberedByFreq(shadows, "--shadow"),
		ZIndices:    numberedByValue(zIndices, "--z"),
		Breakpoints: breakpoints,
	}
}

func topSorted(hist map[string]int, minCount int) [][2]string {
	var pairs [][2]string
	for v, c := range hist {
		if c >= minCount {
			pairs = append(pairs, [2]string{v, ""})
		}
	}
	sort.Slice(pairs, func(i, j int) bool {
		// Sort by frequency desc, then lexicographically
		if hist[pairs[i][0]] != hist[pairs[j][0]] {
			return hist[pairs[i][0]] > hist[pairs[j][0]]
		}
		return pairs[i][0] < pairs[j][0]
	})
	return pairs
}

func numberedByValue(hist map[string]int, prefix string) map[string]string {
	out := make(map[string]string)
	pairs := topSorted(hist, 3)

	// Sort by numeric value
	sort.Slice(pairs, func(i, j int) bool {
		vi := parsePx(pairs[i][0])
		vj := parsePx(pairs[j][0])
		if vi != vj {
			return vi < vj
		}
		return pairs[i][0] < pairs[j][0]
	})

	for i, pair := range pairs {
		key := prefix + "-" + pad("", i+1)
		out[key] = pair[0]
	}
	return out
}

func numberedByFreq(hist map[string]int, prefix string) map[string]string {
	out := make(map[string]string)
	for i, pair := range topSorted(hist, 3) {
		key := prefix + "-" + pad("", i+1)
		out[key] = pair[0]
	}
	return out
}

func parsePx(v string) float64 {
	re := regexp.MustCompile(`-?\d+(?:\.\d+)?`)
	match := re.FindString(v)
	if match == "" {
		return 0
	}
	var f float64
	fmt.Sscanf(match, "%f", &f)
	return f
}

func pad(prefix string, n int) string {
	if prefix == "" {
		return fmt.Sprintf("%03d", n)
	}
	return fmt.Sprintf("%s%03d", prefix, n)
}
