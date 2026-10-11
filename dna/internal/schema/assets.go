package schema

import (
	"strings"
)

type Asset struct {
	Type   string `json:"type"`   // image, video, audio, source, stylesheet
	URL    string `json:"url"`    // Full or data: URL
	Tag    string `json:"tag"`    // img, video, source, link, etc.
	Status string `json:"status,omitempty"` // reference-only (not copied)
}

// ExtractAssets finds all media URLs in the IR.
func ExtractAssets(root *IRNode) []Asset {
	var assets []Asset
	seen := make(map[string]bool)

	var visit func(*IRNode)
	visit = func(node *IRNode) {
		for _, attr := range []string{"src", "href", "poster", "data-src"} {
			if url := node.Attrs[attr]; url != "" {
				if !seen[url] {
					seen[url] = true
					assets = append(assets, Asset{
						Type:   inferAssetType(node.Tag),
						URL:    url,
						Tag:    node.Tag,
						Status: "reference-only", // No copying; just metadata
					})
				}
			}
		}

		for _, child := range node.Children {
			if irChild, ok := child.(*IRNode); ok {
				visit(irChild)
			}
		}
	}

	visit(root)
	return assets
}

func inferAssetType(tag string) string {
	tag = strings.ToLower(tag)
	switch tag {
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
