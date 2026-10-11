package schema

import (
    "strings"
)

// Asset is a media reference from the IR.
type Asset struct {
    Type string `json:"type"`
    URL  string `json:"url"`
    Tag  string `json:"tag"`
}

func ExtractAssets(root *IRNode) []Asset {
    seen := map[string]bool{}
    out := make([]Asset, 0)

    var visit func(*IRNode)
    visit = func(node *IRNode) {
        for _, field := range []string{"src", "href", "poster", "data-src"} {
            if url := node.Attrs[field]; url != "" && !seen[url] {
                seen[url] = true
                out = append(out, Asset{Type: inferAssetType(node.Tag), URL: url, Tag: node.Tag})
            }
        }
        for _, child := range node.Children {
            if irChild := AsNode(child); irChild != nil {
                visit(irChild)
            }
        }
    }

    if root != nil {
        visit(root)
    }
    return out
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
