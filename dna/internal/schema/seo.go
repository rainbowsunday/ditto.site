package schema

// SEO is the minimal SEO metadata required by downstream consumers.
type SEO struct {
    Title        string `json:"title,omitempty"`
    Description  string `json:"description,omitempty"`
    Lang         string `json:"lang,omitempty"`
    Canonical    string `json:"canonical,omitempty"`
    MetaViewport string `json:"metaViewport,omitempty"`
}

func ExtractSEO(ir *IR) SEO {
    return SEO{
        Title: ir.Doc.Title,
        Lang:  ir.Doc.Lang,
    }
}
