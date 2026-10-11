package schema

import (
    "encoding/json"
    "fmt"
)

// IR is the normalized input snapshot provided by the host / SBR layer.
type IR struct {
    Doc  DocInfo `json:"doc"`
    Root *IRNode `json:"root"`
}

// DocInfo is the document-level metadata of the IR snapshot.
type DocInfo struct {
    SourceURL         string                 `json:"sourceUrl"`
    Title             string                 `json:"title"`
    Lang              string                 `json:"lang"`
    CanonicalViewport int                    `json:"canonicalViewport"`
    Viewports         []int                  `json:"viewports"`
    NodeCount         int                    `json:"nodeCount"`
    PerViewport       map[int]ViewportInfo   `json:"perViewport"`
}

type ViewportInfo struct {
    ScrollHeight float64 `json:"scrollHeight"`
    ScrollWidth  float64 `json:"scrollWidth"`
    HtmlBg       string  `json:"htmlBg,omitempty"`
    BodyBg       string  `json:"bodyBg,omitempty"`
    BodyColor    string  `json:"bodyColor,omitempty"`
    BodyFont     string  `json:"bodyFont,omitempty"`
}

type IRNode struct {
    ID           string                  `json:"id"`
    Tag          string                  `json:"tag"`
    Attrs        map[string]string       `json:"attrs,omitempty"`
    VisibleByVp  map[int]bool            `json:"visibleByVp"`
    BBoxByVp     map[int]BBox            `json:"bboxByVp,omitempty"`
    ComputedByVp map[int]map[string]string `json:"computedByVp,omitempty"`
    Children     []IRChild               `json:"children,omitempty"`
}

type BBox struct {
    X      float64 `json:"x"`
    Y      float64 `json:"y"`
    Width  float64 `json:"width"`
    Height float64 `json:"height"`
}

type IRChild interface{}

type IRTextNode struct {
    Text string `json:"text"`
}

func IsTextChild(child IRChild) (*IRTextNode, bool) {
    t, ok := child.(*IRTextNode)
    return t, ok
}

func AsNode(child IRChild) *IRNode {
    n, ok := child.(*IRNode)
    if !ok {
        return nil
    }
    return n
}

func LoadIR(data []byte) (*IR, error) {
    var raw struct {
        Doc  DocInfo        `json:"doc"`
        Root json.RawMessage `json:"root"`
    }
    if err := json.Unmarshal(data, &raw); err != nil {
        return nil, fmt.Errorf("unmarshal IR root: %w", err)
    }

    root, err := unmarshalIRNode(raw.Root)
    if err != nil {
        return nil, err
    }

    return &IR{Doc: raw.Doc, Root: root}, nil
}

func unmarshalIRNode(data json.RawMessage) (*IRNode, error) {
    var node IRNode
    if err := json.Unmarshal(data, &node); err != nil {
        return nil, fmt.Errorf("unmarshal IRNode: %w", err)
    }

    // Mixed child types are handled by raw JSON decode into an interim object.
    var raw map[string]json.RawMessage
    if err := json.Unmarshal(data, &raw); err != nil {
        return nil, fmt.Errorf("decode IRNode raw: %w", err)
    }
    childRaw, ok := raw["children"]
    if !ok {
        return &node, nil
    }

    var items []json.RawMessage
    if err := json.Unmarshal(childRaw, &items); err != nil {
        return nil, fmt.Errorf("decode children: %w", err)
    }

    for _, item := range items {
        var text IRTextNode
        if err := json.Unmarshal(item, &text); err == nil && text.Text != "" {
            node.Children = append(node.Children, &text)
            continue
        }
        child, err := unmarshalIRNode(item)
        if err == nil {
            node.Children = append(node.Children, child)
        }
    }

    return &node, nil
}
