package analysis

import (
    "fmt"
    "regexp"
    "sort"
    "strings"
)

// Section is the minimal structural fingerprint used by the DNA output.
type Section struct {
    ID    string `json:"id"`
    Tag   string `json:"tag"`
    Kind  string `json:"kind"`
    Text  string `json:"text,omitempty"`
    Depth int    `json:"depth"`
}

// DetectSections builds a lightweight site map from the normalized DOM tree.
func DetectSections(root map[string]any) []Section {
    sections := make([]Section, 0)
    var walk func(node map[string]any, depth int)
    walk = func(node map[string]any, depth int) {
        if node == nil {
            return
        }
        tag := asString(node["tag"])
        attrs := asMap(node["attrs"])
        if tag != "" {
            kind := classifyTag(tag, attrs)
            if kind != "" {
                text := summarizeText(node)
                sections = append(sections, Section{
                    ID:    asString(node["id"]),
                    Tag:   tag,
                    Kind:  kind,
                    Text:  text,
                    Depth: depth,
                })
            }
        }
        children, _ := node["children"].([]any)
        for _, child := range children {
            if m, ok := child.(map[string]any); ok {
                walk(m, depth+1)
            }
        }
    }
    walk(root, 0)
    // Stable ordering is important for snapshot tests.
    sort.Slice(sections, func(i, j int) bool {
        if sections[i].Depth == sections[j].Depth {
            return sections[i].Tag < sections[j].Tag
        }
        return sections[i].Depth < sections[j].Depth
    })
    return sections
}

func classifyTag(tag string, attrs map[string]any) string {
    tag = strings.ToLower(tag)
    switch tag {
    case "header":
        return "header"
    case "nav":
        return "nav"
    case "main":
        return "main"
    case "section":
        return "section"
    case "article":
        return "article"
    case "aside":
        return "aside"
    case "footer":
        return "footer"
    }
    classStr := strings.ToLower(asString(attrs["class"]))
    if strings.Contains(classStr, "hero") {
        return "hero"
    }
    if strings.Contains(classStr, "section") {
        return "section"
    }
    if strings.Contains(classStr, "banner") {
        return "banner"
    }
    return ""
}

func summarizeText(node map[string]any) string {
    if text, ok := node["text"].(string); ok {
        return strings.TrimSpace(text)
    }
    children, _ := node["children"].([]any)
    parts := make([]string, 0)
    for _, child := range children {
        if m, ok := child.(map[string]any); ok {
            if t := summarizeText(m); t != "" {
                parts = append(parts, t)
            }
        }
    }
    return strings.Join(parts, " ")
}

var colorRe = regexp.MustCompile(`(?i)(#[0-9a-f]{3,8}|rgba?\([^)]*\)|hsla?\([^)]*\)|var\(--[a-z0-9_-]+\))`)

// TokenSet is a compact DNA token summary.
type TokenSet struct {
    Colors []string `json:"colors"`
    Fonts  []string `json:"fonts"`
    Spacing []int   `json:"spacing"`
}

// ExtractTokens extracts colors, font families and spacing from the captured DOM.
func ExtractTokens(root map[string]any) TokenSet {
    colors := make(map[string]struct{})
    fonts := make(map[string]struct{})
    spacing := make(map[int]struct{})

    var walk func(map[string]any)
    walk = func(node map[string]any) {
        if node == nil {
            return
        }
        // Accept styles in attrs or computed maps.
        if attrs, ok := node["attrs"].(map[string]any); ok {
            for _, key := range []string{"class", "style"} {
                if raw, ok := attrs[key].(string); ok {
                    for _, m := range colorRe.FindAllString(raw, -1) {
                        colors[m] = struct{}{}
                    }
                    if strings.Contains(raw, "font") || strings.Contains(raw, "font-family") {
                        // coarse extraction
                        for _, token := range strings.Fields(raw) {
                            if strings.Contains(token, "family") || strings.Contains(token, "font") {
                                continue
                            }
                        }
                    }
                }
            }
        }
        if computed, ok := node["computed"].(map[string]any); ok {
            if v, ok := computed["color"].(string); ok {
                for _, m := range colorRe.FindAllString(v, -1) {
                    colors[m] = struct{}{}
                }
            }
            if v, ok := computed["font-family"].(string); ok {
                fonts[strings.TrimSpace(v)] = struct{}{}
            }
            if v, ok := computed["font-size"].(string); ok {
                if px := parsePx(v); px > 0 {
                    spacing[px] = struct{}{}
                }
            }
        }
        children, _ := node["children"].([]any)
        for _, child := range children {
            if m, ok := child.(map[string]any); ok {
                walk(m)
            }
        }
    }
    walk(root)

    out := TokenSet{Colors: sortedKeys(colors), Fonts: sortedKeys(fonts), Spacing: sortedInts(spacing)}
    return out
}

func parsePx(raw string) int {
    raw = strings.TrimSpace(raw)
    if raw == "" {
        return 0
    }
    raw = strings.TrimSuffix(raw, "px")
    var v int
    if _, err := fmt.Sscanf(raw, "%d", &v); err != nil {
        return 0
    }
    return v
}

func sortedKeys(set map[string]struct{}) []string {
    out := make([]string, 0, len(set))
    for key := range set {
        out = append(out, key)
    }
    sort.Strings(out)
    return out
}

func sortedInts(set map[int]struct{}) []int {
    out := make([]int, 0, len(set))
    for key := range set {
        out = append(out, key)
    }
    sort.Ints(out)
    return out
}

func asString(v any) string {
    if v == nil {
        return ""
    }
    if s, ok := v.(string); ok {
        return s
    }
    return fmt.Sprintf("%v", v)
}

func asMap(v any) map[string]any {
    if m, ok := v.(map[string]any); ok {
        return m
    }
    if m, ok := v.(map[string]string); ok {
        out := make(map[string]any, len(m))
        for k, val := range m {
            out[k] = val
        }
        return out
    }
    return nil
}
