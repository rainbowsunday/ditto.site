package schema

type SEO struct {
	Title        string `json:"title,omitempty"`
	Description  string `json:"description,omitempty"`
	Lang         string `json:"lang,omitempty"`
	Canonical    string `json:"canonical,omitempty"`
	MetaViewport string `json:"metaViewport,omitempty"`
}

// ExtractSEO extracts SEO metadata from capture.
func ExtractSEO(capture map[string]interface{}) SEO {
	doc, ok := capture["doc"].(map[string]interface{})
	if !ok {
		return SEO{}
	}

	return SEO{
		Title:       toString(doc["title"]),
		Description: toString(doc["description"]),
		Lang:        toString(doc["lang"]),
		MetaViewport: toString(doc["metaViewport"]),
	}
}

func toString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
