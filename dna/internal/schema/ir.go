package schema

// IRNode matches compiler/src/normalize/ir.ts IRNode type.
type IRNode struct {
	ID           string              `json:"id"`
	Tag          string              `json:"tag"`
	Attrs        map[string]string   `json:"attrs,omitempty"`
	VisibleByVp  map[int]bool        `json:"visibleByVp"`
	BBoxByVp     map[int]BBox        `json:"bboxByVp,omitempty"`
	ComputedByVp map[int]map[string]string `json:"computedByVp,omitempty"`
	Children     []interface{}       `json:"children,omitempty"`
}

// IRTextNode is a text node in the IR.
type IRTextNode struct {
	Text string `json:"text"`
}

// ViewportInfo is per-viewport document metadata.
type ViewportInfo struct {
	ScrollHeight float64
	ScrollWidth  float64
	HtmlBg       string
	BodyBg       string
	BodyColor    string
	BodyFont     string
}

// PageCapture is the minimal canonical snapshot.
type PageCapture struct {
	Root map[string]interface{} `json:"root"`
	Doc  struct {
		URL            string  `json:"url"`
		Title          string  `json:"title"`
		Lang           string  `json:"lang"`
		ScrollHeight   float64 `json:"scrollHeight"`
		ScrollWidth    float64 `json:"scrollWidth"`
		HtmlBg         string  `json:"htmlBg"`
		BodyBg         string  `json:"bodyBg"`
		BodyColor      string  `json:"bodyColor"`
		BodyFont       string  `json:"bodyFont"`
	} `json:"doc"`
}

func isText(child interface{}) bool {
	_, ok := child.(*IRTextNode)
	return ok
}
