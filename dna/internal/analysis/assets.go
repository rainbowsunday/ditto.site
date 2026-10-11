package analysis

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    "sort"
    "strings"
)

// Asset describes a media asset discovered from the capture.
type Asset struct {
    Type string `json:"type"`
    URL  string `json:"url"`
    Tag  string `json:"tag"`
}

// BuildAssets enumerates image, video and source candidates from the DOM tree.
func BuildAssets(root map[string]any) []Asset {
    assets := make([]Asset, 0)
    seen := map[string]struct{}{}
    var walk func(map[string]any)
    walk = func(node map[string]any) {
        if node == nil {
            return
        }
        attrs := asMap(node["attrs"])
        tag := strings.ToLower(asString(node["tag"]))
        if tag == "img" || tag == "source" || tag == "video" || tag == "audio" || tag == "link" {
            urls := make([]string, 0)
            for _, field := range []string{"src", "href", "poster", "srcset"} {
                if val := asString(attrs[field]); val != "" {
                    urls = append(urls, val)
                }
            }
            for _, u := range urls {
                if u == "" {
                    continue
                }
                key := strings.TrimSpace(u)
                if _, ok := seen[key]; ok {
                    continue
                }
                seen[key] = struct{}{}
                assets = append(assets, Asset{Type: inferAssetType(tag), URL: key, Tag: tag})
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
    sort.Slice(assets, func(i, j int) bool {
        if assets[i].Type == assets[j].Type {
            return assets[i].URL < assets[j].URL
        }
        return assets[i].Type < assets[j].Type
    })
    return assets
}

func inferAssetType(tag string) string {
    switch strings.ToLower(tag) {
    case "img":
        return "image"
    case "video":
        return "video"
    case "audio":
        return "audio"
    case "source":
        return "source"
    case "link":
        return "stylesheet"
    default:
        return "asset"
    }
}

// FontFamily describes an observed font family.
type FontFamily struct {
    Family string `json:"family"`
    Source string `json:"source,omitempty"`
}

// BuildFonts returns font families derived from DOM and captured metadata.
func BuildFonts(doc map[string]any, root map[string]any) []FontFamily {
    out := make([]FontFamily, 0)
    seen := map[string]struct{}{}

    add := func(v string, source string) {
        v = strings.TrimSpace(v)
        if v == "" {
            return
        }
        if _, ok := seen[v]; ok {
            return
        }
        seen[v] = struct{}{}
        out = append(out, FontFamily{Family: v, Source: source})
    }

    if doc != nil {
        for _, key := range []string{"bodyFont", "fontFamily", "font-family"} {
            if v, ok := doc[key].(string); ok {
                add(v, "document")
            }
        }
    }
    var walk func(map[string]any)
    walk = func(node map[string]any) {
        if node == nil {
            return
        }
        if computed, ok := node["computed"].(map[string]any); ok {
            if family, ok := computed["font-family"].(string); ok {
                add(family, "computed")
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
    sort.Slice(out, func(i, j int) bool { return out[i].Family < out[j].Family })
    return out
}

// SEOReport is a compact version of the site metadata emitted by the compiler.
type SEOReport struct {
    Title string `json:"title,omitempty"`
    Description string `json:"description,omitempty"`
    Lang string `json:"lang,omitempty"`
    Canonical string `json:"canonical,omitempty"`
    MetaViewport string `json:"metaViewport,omitempty"`
}

// BuildSEO extracts basic metadata from the capture document.
func BuildSEO(doc map[string]any) SEOReport {
    report := SEOReport{}
    if doc == nil {
        return report
    }
    if v, ok := doc["title"].(string); ok {
        report.Title = v
    }
    if v, ok := doc["lang"].(string); ok {
        report.Lang = v
    }
    if v, ok := doc["metaViewport"].(string); ok {
        report.MetaViewport = v
    }
    if head, ok := doc["head"].(map[string]any); ok {
        if v, ok := head["description"].(string); ok {
            report.Description = v
        }
        if v, ok := head["canonical"].(string); ok {
            report.Canonical = v
        }
    }
    return report
}

// WriteJSON marshals the supplied value to disk with stable formatting.
func WriteJSON(path string, val any) error {
    if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
        return fmt.Errorf("mkdir %s: %w", filepath.Dir(path), err)
    }
    b, err := json.MarshalIndent(val, "", "  ")
    if err != nil {
        return fmt.Errorf("marshal %s: %w", path, err)
    }
    return os.WriteFile(path, append(b, '\n'), 0o644)
}
