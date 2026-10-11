package schema

import (
    "fmt"
    "math"
    "sort"
)

// Section is the inferred top-level band metadata.
type Section struct {
    ID       string       `json:"id"`
    NodeID   string       `json:"nodeId"`
    Role     string       `json:"role"`
    Order    int          `json:"order"`
    BBoxByVp map[int]BBox `json:"bboxByVp"`
}

const (
    MIN_BAND_H         = 64.0
    MIN_SEMANTIC_H     = 24.0
    MIN_BAND_W_FRAC    = 0.55
    WRAPPER_FRAC       = 0.5
    MIN_CHILD_COVERAGE = 0.8
    MAX_STACK_OVERLAP  = 0.3
    MAX_BANDS_PER_SPLIT = 20
    MAX_DEPTH          = 12
)

var semanticTags = map[string]bool{
    "header":  true,
    "nav":     true,
    "main":    true,
    "section": true,
    "article": true,
    "footer":  true,
    "aside":   true,
}

func DetectSections(ir *IR) []Section {
    cw := ir.Doc.CanonicalViewport
    pageH := ir.Doc.PerViewport[cw].ScrollHeight
    nodes := detectSectionNodes(ir.Root, cw, pageH)
    roles := guessRoles(nodes, cw, pageH)

    out := make([]Section, 0, len(nodes))
    for i, node := range nodes {
        out = append(out, Section{
            ID:       fmt.Sprintf("section-%03d", i+1),
            NodeID:   node.ID,
            Role:     roles[i],
            Order:    i,
            BBoxByVp: bboxesFor(node),
        })
    }
    return out
}

type candidate struct {
    node   *IRNode
    y      float64
    width  float64
    height float64
}

type context struct {
    cw    int
    pageH float64
}

func detectSectionNodes(root *IRNode, cw int, pageH float64) []*IRNode {
    bands := bandsOf(root, 0, &context{cw: cw, pageH: pageH})
    sort.Slice(bands, func(i, j int) bool { return bands[i].y < bands[j].y })
    final := deduplicateBands(bands)
    if len(final) == 0 {
        return []*IRNode{root}
    }

    nodes := make([]*IRNode, 0, len(final))
    for _, c := range final {
        nodes = append(nodes, c.node)
    }
    return nodes
}

func bandsOf(container *IRNode, depth int, ctx *context) []*candidate {
    if depth > MAX_DEPTH {
        return nil
    }

    out := make([]*candidate, 0)
    for _, child := range container.Children {
        if _, ok := IsTextChild(child); ok {
            continue
        }
        irChild := AsNode(child)
        if irChild == nil {
            continue
        }

        cand := candOf(irChild, ctx)
        if cand == nil {
            out = append(out, bandsOf(irChild, depth+1, ctx)...)
            continue
        }

        if ctx.pageH > 0 && cand.height > ctx.pageH*WRAPPER_FRAC {
            inner := bandsOf(irChild, depth+1, ctx)
            if acceptSplit(inner, cand) {
                out = append(out, inner...)
                continue
            }
        }

        out = append(out, cand)
    }
    return out
}

func candOf(node *IRNode, ctx *context) *candidate {
    bbox, ok := node.BBoxByVp[ctx.cw]
    if !ok || !node.VisibleByVp[ctx.cw] {
        return nil
    }

    if computed, ok := node.ComputedByVp[ctx.cw]; ok {
        display := computed["display"]
        if display == "inline" || display == "inline-block" || display == "none" {
            return nil
        }
    }

    if bbox.Width < float64(ctx.cw)*MIN_BAND_W_FRAC {
        return nil
    }

    minH := MIN_BAND_H
    if semanticTags[node.Tag] || navEvidence(node) {
        minH = MIN_SEMANTIC_H
    }
    if bbox.Height < minH {
        return nil
    }

    return &candidate{node: node, y: bbox.Y, width: bbox.Width, height: bbox.Height}
}

func acceptSplit(inner []*candidate, parent *candidate) bool {
    if len(inner) < 2 || len(inner) > MAX_BANDS_PER_SPLIT {
        return false
    }
    for i := 1; i < len(inner); i++ {
        prev := inner[i-1]
        cur := inner[i]
        overlap := prev.y + prev.height - cur.y
        if overlap > math.Min(prev.height, cur.height)*MAX_STACK_OVERLAP {
            return false
        }
    }

    var covered float64
    cursor := -math.MaxFloat64
    for _, c := range inner {
        top := math.Max(c.y, cursor)
        bottom := c.y + c.height
        if bottom > top {
            covered += bottom - top
        }
        cursor = math.Max(cursor, bottom)
    }
    return covered >= parent.height*MIN_CHILD_COVERAGE
}

func deduplicateBands(bands []*candidate) []*candidate {
    out := make([]*candidate, 0, len(bands))
    for _, c := range bands {
        dup := false
        for _, f := range out {
            if math.Abs(f.y-c.y) < 4 && math.Abs(f.height-c.height) < 4 {
                dup = true
                break
            }
        }
        if !dup {
            out = append(out, c)
        }
    }
    return out
}

func guessRoles(nodes []*IRNode, cw int, pageH float64) []string {
    roles := make([]string, len(nodes))
    heroClaimed := false

    for i, node := range nodes {
        switch node.Tag {
        case "nav":
            roles[i] = "navbar"
            continue
        case "header":
            if heroLikeHeader(node, cw) && !heroClaimed {
                heroClaimed = true
                roles[i] = "hero"
            } else {
                roles[i] = "header"
            }
            continue
        case "footer":
            roles[i] = "footer"
            continue
        case "aside":
            roles[i] = "aside"
            continue
        case "main":
            roles[i] = "main"
            continue
        }

        bbox := node.BBoxByVp[cw]
        if bbox == nil {
            roles[i] = "section"
            continue
        }

        if bbox.Height <= 140 && bbox.Y <= 160 && (i == 0 || navEvidence(node)) {
            roles[i] = "navbar"
            continue
        }

        if !heroClaimed && bbox.Y <= math.Max(900, pageH*0.25) {
            heroClaimed = true
            roles[i] = "hero"
            continue
        }

        if i == len(nodes)-1 {
            roles[i] = "footer"
            continue
        }

        roles[i] = "section"
    }

    return roles
}

func heroLikeHeader(node *IRNode, cw int) bool {
    bbox := node.BBoxByVp[cw]
    if bbox == nil || bbox.Height < 200 {
        return false
    }
    return subtreeHas(node, func(n *IRNode) bool { return n.Tag == "h1" }, 6)
}

func navEvidence(node *IRNode) bool {
    if node.Attrs["role"] == "banner" || node.Attrs["role"] == "navigation" {
        return true
    }
    return subtreeHas(node, func(n *IRNode) bool {
        return n.Tag == "nav" || n.Attrs["role"] == "navigation"
    }, 3)
}

func subtreeHas(node *IRNode, pred func(*IRNode) bool, depth int) bool {
    if depth < 0 {
        return false
    }
    for _, child := range node.Children {
        if _, ok := IsTextChild(child); ok {
            continue
        }
        irChild := AsNode(child)
        if irChild == nil {
            continue
        }
        if pred(irChild) || subtreeHas(irChild, pred, depth-1) {
            return true
        }
    }
    return false
}

func bboxesFor(node *IRNode) map[int]BBox {
    out := make(map[int]BBox)
    for vp, b := range node.BBoxByVp {
        out[vp] = BBox{
            X:      round(b.X),
            Y:      round(b.Y),
            Width:  round(b.Width),
            Height: round(b.Height),
        }
    }
    return out
}

func round(v float64) float64 {
    return math.Round(v*100) / 100
}
