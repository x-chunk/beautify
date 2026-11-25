package blocks

import (
	"go.xchunk.org/beautify"
	"go.xchunk.org/beautify/dra"
)

// Node is an element of a Tree. A nil or empty label still occupies one
// line, so the structure stays visible.
type Node struct {
	Label    beautify.Composable
	Children []Node
}

// GuideSet holds the text that connects the nodes of a Tree. The fields
// should have equal widths; narrower ones are padded to the widest.
type GuideSet struct {
	Branch string // before a child that has siblings after it
	Last   string // before the last child
	Pipe   string // below Branch
	Space  string // below Last
}

// Provided guide sets.
var (
	// GuidesUnicode has square corners. It is the default.
	GuidesUnicode = GuideSet{Branch: "├── ", Last: "└── ", Pipe: "│   ", Space: "    "}
	// GuidesRounded has a round corner at the last child.
	GuidesRounded = GuideSet{Branch: "├── ", Last: "╰── ", Pipe: "│   ", Space: "    "}
	// GuidesASCII uses only ASCII characters.
	GuidesASCII = GuideSet{Branch: "|-- ", Last: "`-- ", Pipe: "|   ", Space: "    "}
)

// TreeOption configures Tree.
type TreeOption func(*tree)

// TreeGuides sets the text that connects the nodes. The default is
// GuidesUnicode.
func TreeGuides(set GuideSet) TreeOption {
	return func(t *tree) { t.guides = set }
}

// TreeGuideStyle sets the style of the guides.
func TreeGuideStyle(st dra.Style) TreeOption {
	return func(t *tree) { t.style = st }
}

// TreeMaxDepth limits the depth of the nodes that are shown; the root has
// depth 0. The children of a node at depth n are replaced by a single "…".
// With n <= 0 there is no limit.
func TreeMaxDepth(n int) TreeOption {
	return func(t *tree) { t.maxDepth = n }
}

type tree struct {
	root     Node
	guides   GuideSet
	style    dra.Style
	maxDepth int
}

// Tree shows a hierarchy of nodes connected by guides. In a bounded context
// a label gets the width that the guides leave, but at least 1.
func Tree(root Node, opts ...TreeOption) beautify.Composable {
	t := tree{root: cloneNode(root), guides: GuidesUnicode}
	applyOptions(&t, opts)
	return t
}

func cloneNode(n Node) Node {
	children := make([]Node, len(n.Children))
	for i, c := range n.Children {
		children[i] = cloneNode(c)
	}
	n.Children = children
	return n
}

// Render implements beautify.Composable.
func (t tree) Render(ctx beautify.Context) (dra.Model, error) {
	if noRoom(ctx) {
		return dra.Model{}, nil
	}
	w := treeWalker{ctx: ctx, maxDepth: t.maxDepth}
	w.setGuides(t.guides, t.style)
	if err := w.node(t.root, dra.Model{}, dra.Model{}, 0); err != nil {
		return dra.Model{}, err
	}
	return dra.Stack(dra.AlignLeft, w.rows...), nil
}

// treeWalker renders nodes in depth-first order. It works from the root down
// with ready prefixes and adds each label once: prefixing whole subtrees on
// the way up would touch every line once per level.
type treeWalker struct {
	ctx      beautify.Context
	maxDepth int

	branch, last, pipe, space dra.Model
	guideWidth                int

	index int // of the next label, for errors
	rows  []dra.Model
}

// setGuides builds the guides and pads them to the same width.
func (w *treeWalker) setGuides(set GuideSet, st dra.Style) {
	guides := []*dra.Model{&w.branch, &w.last, &w.pipe, &w.space}
	for i, s := range []string{set.Branch, set.Last, set.Pipe, set.Space} {
		*guides[i] = w.ctx.Factory().Styled(oneLine(s), st)
		w.guideWidth = max(w.guideWidth, guides[i].Width())
	}
	for _, g := range guides {
		if g.IsEmpty() {
			*g = dra.Blank(w.guideWidth, 1)
		}
		*g = dra.Pad(*g, w.guideWidth, dra.AlignLeft)
	}
}

// node adds the label of n and its subtree. first goes before the first line
// of the label, rest before its other lines and before the guides of the
// children.
func (w *treeWalker) node(n Node, first, rest dra.Model, depth int) error {
	label, err := w.label(n.Label, depth)
	if err != nil {
		return err
	}
	w.add(label, first, rest)
	if w.maxDepth > 0 && depth == w.maxDepth && len(n.Children) > 0 {
		w.add(w.ctx.Factory().Text(ellipsis), w.extend(rest, w.last), dra.Model{})
		return nil
	}
	for i, c := range n.Children {
		guide, below := w.branch, w.pipe
		if i == len(n.Children)-1 {
			guide, below = w.last, w.space
		}
		if err := w.node(c, w.extend(rest, guide), w.extend(rest, below), depth+1); err != nil {
			return err
		}
	}
	return nil
}

// label renders a label with the width that the guides at depth leave.
func (w *treeWalker) label(c beautify.Composable, depth int) (dra.Model, error) {
	index := w.index
	w.index++
	if c == nil {
		return dra.Blank(0, 1), nil
	}
	ctx := w.ctx
	if width, bounded := ctx.Width(); bounded {
		ctx = ctx.WithWidth(max(width-depth*w.guideWidth, 1))
	}
	m, err := renderChild(ctx, "tree", index, c)
	if err == nil && m.IsEmpty() {
		m = dra.Blank(0, 1)
	}
	return m, err
}

// add appends the lines of one label.
func (w *treeWalker) add(label, first, rest dra.Model) {
	w.rows = append(w.rows, clip(w.ctx, dra.PrefixLines(label, first, rest)))
}

// extend returns prefix followed by guide. Prefixes do not grow beyond the
// bound, which keeps deep trees linear.
func (w *treeWalker) extend(prefix, guide dra.Model) dra.Model {
	if width, bounded := w.ctx.Width(); bounded && prefix.Width() >= width {
		return prefix
	}
	return dra.Concat(prefix, guide)
}
