package schema

import (
	"fmt"
	"math"
)

// Section matches compiler/src/infer/sections.ts Section type.
type Section struct {
	ID       string            `json:"id"`      // e.g., section-001
	NodeID   string            `json:"nodeId"`  // IR node id (e.g., n42)
	Role     string            `json:"role"`    // navbar, hero, footer, section, etc.
	Order    int               `json:"order"`   // 0, 1, 2, ...
	BBoxByVp map[int]BBox      `json:"bboxByVp"`// Per-viewport bounding boxes
}

type BBox struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// Thresholds from compiler/src/infer/sections.ts
const (
	MIN_BAND_H        = 64
	MIN_SEMANTIC_H    = 24
	MIN_BAND_W_FRAC   = 0.55
	WRAPPER_FRAC      = 0.5
	MIN_CHILD_COVERAGE = 0.8
	MAX_STACK_OVERLAP = 0.3
	MAX_BANDS_PER_SPLIT = 20
	MAX_DEPTH         = 12
)

var SEMANTIC_TAGS = map[string]bool{
	"header": true, "nav": true, "main": true, "section": true,
	"article": true, "footer": true, "aside": true,
}

// DetectSections detects page sections from an IR root and returns ordered Section array.
// This is the direct port of compiler/src/infer/sections.ts:detectSections().
func DetectSections(root *IRNode, canonical int, perViewport map[int]ViewportInfo) []Section {
	pageH := perViewport[canonical].ScrollHeight
	nodes := detectSectionNodes(root, canonical, pageH)
	roles := guessRoles(nodes, canonical, pageH)

	var sections []Section
	for i, node := range nodes {
		sections = append(sections, Section{
			ID:       fmt.Sprintf("section-%03d", i+1),
			NodeID:   node.ID,
			Role:     roles[i],
			Order:    i,
			BBoxByVp: bboxesFor(node),
		})
	}
	return sections
}

type candidate struct {
	node   *IRNode
	y      float64
	width  float64
	height float64
}

func detectSectionNodes(root *IRNode, cw int, pageH float64) []*IRNode {
	ctx := context{cw: cw, pageH: pageH}
	bands := bandsOf(root, 0, &ctx)

	// Sort by y, deduplicate near-identical boxes
	sortBands(bands)
	final := deduplicateBands(bands)

	if len(final) == 0 {
		return []*IRNode{root}
	}

	var nodes []*IRNode
	for _, c := range final {
		nodes = append(nodes, c.node)
	}
	return nodes
}

type context struct {
	cw    int
	pageH float64
}

func candOf(node *IRNode, ctx *context) *candidate {
	bbox, ok := node.BBoxByVp[ctx.cw]
	if !ok || !node.VisibleByVp[ctx.cw] {
		return nil
	}

	display := node.ComputedByVp[ctx.cw]["display"]
	if isInlineOrNone(display) {
		return nil
	}

	if bbox.Width < float64(ctx.cw)*MIN_BAND_W_FRAC {
		return nil
	}

	minH := MIN_BAND_H
	if SEMANTIC_TAGS[node.Tag] || navEvidence(node) {
		minH = MIN_SEMANTIC_H
	}

	if bbox.Height < float64(minH) {
		return nil
	}

	return &candidate{node: node, y: bbox.Y, width: bbox.Width, height: bbox.Height}
}

func bandsOf(container *IRNode, depth int, ctx *context) []*candidate {
	if depth > MAX_DEPTH {
		return nil
	}

	var out []*candidate
	for _, child := range container.Children {
		if isText(child) {
			continue
		}

		irChild := child.(*IRNode)
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

func acceptSplit(inner []*candidate, parent *candidate) bool {
	if len(inner) < 2 || len(inner) > MAX_BANDS_PER_SPLIT {
		return false
	}

	// Check for excessive overlap
	for i := 1; i < len(inner); i++ {
		prev := inner[i-1]
		cur := inner[i]
		overlap := prev.y + prev.height - cur.y
		if overlap > math.Min(prev.height, cur.height)*MAX_STACK_OVERLAP {
			return false
		}
	}

	// Check coverage
	var covered float64
	cursor := -math.MaxFloat64
	for _, c := range inner {
		top := math.Max(c.y, cursor)
		bot := c.y + c.height
		if bot > top {
			covered += bot - top
		}
		cursor = math.Max(cursor, bot)
	}

	return covered >= parent.height*MIN_CHILD_COVERAGE
}

func sortBands(bands []*candidate) {
	// Simple bubble sort (Go's sort.Slice is available)
	for i := 1; i < len(bands); i++ {
		for j := 0; j < i; j++ {
			if bands[i].y < bands[j].y {
				bands[i], bands[j] = bands[j], bands[i]
			}
		}
	}
}

func deduplicateBands(bands []*candidate) []*candidate {
	var final []*candidate
	for _, c := range bands {
		dup := false
		for _, f := range final {
			if math.Abs(f.y-c.y) < 4 && math.Abs(f.height-c.height) < 4 {
				dup = true
				break
			}
		}
		if !dup {
			final = append(final, c)
		}
	}
	return final
}

func guessRoles(nodes []*IRNode, cw int, pageH float64) []string {
	roles := make([]string, len(nodes))
	heroClaimed := false

	for i, node := range nodes {
		if node.Tag == "nav" {
			roles[i] = "navbar"
			continue
		}
		if node.Tag == "header" {
			if !heroLikeHeader(node, cw) {
				roles[i] = "header"
			} else {
				heroClaimed = true
				roles[i] = "hero"
			}
			continue
		}
		if node.Tag == "footer" {
			roles[i] = "footer"
			continue
		}
		if node.Tag == "aside" {
			roles[i] = "aside"
			continue
		}
		if node.Tag == "main" {
			roles[i] = "main"
			continue
		}

		bbox := node.BBoxByVp[cw]
		if bbox == nil {
			roles[i] = "section"
			continue
		}

		h := bbox.Height
		y := bbox.Y

		// Thin top bar
		if h <= 140 && y <= 160 && (i == 0 || navEvidence(node)) {
			roles[i] = "navbar"
			continue
		}

		// Unclaimed hero
		if !heroClaimed && y <= math.Max(900, pageH*0.25) {
			heroClaimed = true
			roles[i] = "hero"
			continue
		}

		// Last section
		if i == len(nodes)-1 {
			roles[i] = "footer"
			continue
		}

		// Default
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
		if isText(child) {
			continue
		}
		irChild := child.(*IRNode)
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

func isInlineOrNone(display string) bool {
	return display == "inline" || display == "inline-block" || display == "none"
}

func round(v float64) float64 {
	return math.Round(v*100) / 100
}
