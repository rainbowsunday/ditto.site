package schema

import (
    "fmt"
    "math"
    "regexp"
    "sort"
    "strings"
)

// Tokens matches the TS design-token output shape.
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

var transparent = map[string]bool{
    "rgba(0, 0, 0, 0)": true,
    "transparent":      true,
    "rgba(0,0,0,0)":    true,
}

func ExtractTokens(ir *IR) Tokens {
    cw := ir.Doc.CanonicalViewport
    colors := map[string]int{}
    fonts := map[string]int{}
    fontSizes := map[string]int{}
    fontWeights := map[string]int{}
    lineHeights := map[string]int{}
    spacing := map[string]int{}
    radii := map[string]int{}
    shadows := map[string]int{}
    zIndices := map[string]int{}

    var visit func(*IRNode)
    visit = func(node *IRNode) {
        if computed, ok := node.ComputedByVp[cw]; ok && node.VisibleByVp[cw] {
            if v := computed["color"]; v != "" && !transparent[v] {
                colors[v]++
            }
            if v := computed["backgroundColor"]; v != "" && !transparent[v] {
                colors[v]++
            }
            if v := computed["fontFamily"]; v != "" {
                fonts[v]++
            }
            if v := computed["fontSize"]; v != "" {
                fontSizes[v]++
            }
            if v := computed["fontWeight"]; v != "" {
                fontWeights[v]++
            }
            if v := computed["lineHeight"]; v != "" && v != "normal" {
                lineHeights[v]++
            }
            for _, p := range []string{"paddingTop", "paddingBottom", "paddingLeft", "paddingRight", "marginTop", "marginBottom", "gap"} {
                if v := computed[p]; v != "" && v != "0px" && strings.HasSuffix(v, "px") {
                    spacing[v]++
                }
            }
            for _, p := range []string{"borderTopLeftRadius", "borderTopRightRadius"} {
                if v := computed[p]; v != "" && v != "0px" {
                    radii[v]++
                }
            }
            if v := computed["boxShadow"]; v != "" && v != "none" {
                shadows[v]++
            }
            if v := computed["zIndex"]; v != "" && v != "auto" {
                zIndices[v]++
            }
        }

        for _, child := range node.Children {
            if irChild := AsNode(child); irChild != nil {
                visit(irChild)
            }
        }
    }
    visit(ir.Root)

    pv := ir.Doc.PerViewport[cw]
    pageBg := pv.BodyBg
    if pageBg == "" || transparent[pageBg] {
        pageBg = "rgb(255, 255, 255)"
    }

    colorTokens := map[string]string{"--color-bg-page": pageBg}
    primaryText := pv.BodyColor
    if primaryText == "" || transparent[primaryText] {
        if top := topSorted(colors, 1); len(top) > 0 {
            primaryText = top[0].val
        }
    }
    if primaryText != "" {
        colorTokens["--color-text-primary"] = primaryText
    }

    usedColors := map[string]bool{pageBg: true, primaryText: true}
    ci := 1
    for _, pair := range topSorted(colors, 3) {
        if usedColors[pair.val] {
            continue
        }
        usedColors[pair.val] = true
        colorTokens[fmt.Sprintf("--color-%03d", ci)] = pair.val
        ci++
    }

    fontTokens := map[string]string{}
    if top := topSorted(fonts, 1); len(top) > 0 {
        fontTokens["--font-body"] = top[0].val
    }
    fi := 1
    primaryFont := ""
    if top := topSorted(fonts, 1); len(top) > 0 {
        primaryFont = top[0].val
    }
    for _, pair := range topSorted(fonts, 3) {
        if pair.val == primaryFont {
            continue
        }
        fontTokens[fmt.Sprintf("--font-%03d", fi)] = pair.val
        fi++
    }

    breakpoints := map[string]string{}
    for _, vp := range ir.Doc.Viewports {
        breakpoints[fmt.Sprintf("--bp-%d", vp)] = fmt.Sprintf("%dpx", vp)
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

type histEntry struct {
    val   string
    count int
}

func topSorted(hist map[string]int, min int) []histEntry {
    out := make([]histEntry, 0)
    for v, c := range hist {
        if c >= min {
            out = append(out, histEntry{val: v, count: c})
        }
    }
    sort.Slice(out, func(i, j int) bool {
        if out[i].count != out[j].count {
            return out[i].count > out[j].count
        }
        return out[i].val < out[j].val
    })
    return out
}

func numberedByValue(hist map[string]int, prefix string) map[string]string {
    pairs := topSorted(hist, 3)
    sort.Slice(pairs, func(i, j int) bool {
        ai := parsePx(pairs[i].val)
        aj := parsePx(pairs[j].val)
        if ai != aj {
            return ai < aj
        }
        return pairs[i].val < pairs[j].val
    })
    out := map[string]string{}
    for i, p := range pairs {
        out[fmt.Sprintf("%s-%03d", prefix, i+1)] = p.val
    }
    return out
}

func numberedByFreq(hist map[string]int, prefix string) map[string]string {
    out := map[string]string{}
    for i, p := range topSorted(hist, 3) {
        out[fmt.Sprintf("%s-%03d", prefix, i+1)] = p.val
    }
    return out
}

func parsePx(v string) float64 {
    re := regexp.MustCompile(`-?\d+(?:\.\d+)?`)
    if match := re.FindString(v); match != "" {
        var f float64
        _, _ = fmt.Sscanf(match, "%f", &f)
        return f
    }
    return math.NaN()
}
