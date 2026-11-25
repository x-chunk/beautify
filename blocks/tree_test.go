package blocks

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"go.xchunk.org/beautify"
	"go.xchunk.org/beautify/dra"
)

func node(label string, children ...Node) Node {
	return Node{Label: raw(label), Children: children}
}

func sampleTree() Node {
	return node("root",
		node("src", node("a.go"), node("b.go\n(new)", node("deep"))),
		node("docs\n(old)", node("x")),
		node("end"),
	)
}

func TestTree(t *testing.T) {
	checkText(t, Tree(sampleTree()), unbounded,
		"root",
		"├── src",
		"│   ├── a.go",
		"│   └── b.go",
		"│       (new)",
		"│       └── deep",
		"├── docs",
		"│   (old)",
		"│   └── x",
		"└── end")
	checkText(t, Tree(node("only")), unbounded, "only")
	checkText(t, Tree(node("two\nlines", node("c"))), unbounded, "two", "lines", "└── c")
	checkText(t, Tree(Node{Children: []Node{{}, node("")}}), unbounded, "", "├── ", "└── ")
	checkText(t, Tree(Node{}), unbounded, "")
	checkText(t, Tree(node("a", node("b", node("c"))), TreeGuides(GuidesASCII)), unbounded,
		"a", "`-- b", "    `-- c")
	checkText(t, Tree(node("a", node("b", node("c")), node("d")), TreeGuides(GuidesRounded)), unbounded,
		"a", "├── b", "│   ╰── c", "╰── d")
}

func TestTreeCustomGuides(t *testing.T) {
	set := GuideSet{Branch: "+-", Last: "`", Pipe: "|", Space: ""}
	checkText(t, Tree(node("a", node("b", node("c")), node("d", node("e"))), TreeGuides(set)), unbounded,
		"a", "+-b", "| ` c", "` d", "  ` e")
	checkText(t, Tree(node("a", node("b")), TreeGuides(GuideSet{})), unbounded, "a", "b")
}

func TestTreeMaxDepth(t *testing.T) {
	checkText(t, Tree(sampleTree(), TreeMaxDepth(1)), unbounded,
		"root",
		"├── src",
		"│   └── …",
		"├── docs",
		"│   (old)",
		"│   └── …",
		"└── end")
	full := render(t, Tree(sampleTree()), unbounded).String()
	for _, n := range []int{0, -1, 3, 9} {
		checkText(t, Tree(sampleTree(), TreeMaxDepth(n)), unbounded, full)
	}
	// Labels that are not shown are not rendered.
	checkText(t, Tree(node("a", node("b", Node{Label: failing()})), TreeMaxDepth(1)), unbounded, "a", "└── b", "    └── …")
}

func TestTreeWidth(t *testing.T) {
	tree := Node{Label: wrapping("root node"), Children: []Node{
		{Label: wrapping("aa bb"), Children: []Node{{Label: wrapping("cc dd")}}},
	}}
	checkText(t, Tree(tree), 10, "root node", "└── aa bb", "    └── cc", "        dd")
	checkText(t, Tree(tree), 6, "root", "node", "└── aa", "    bb", "    └─", "      ", "      ", "      ")
	checkEmpty(t, Tree(tree), 0)

	var widths []int
	rec := beautify.ComposableFunc(func(ctx beautify.Context) (dra.Model, error) {
		w, _ := ctx.Width()
		widths = append(widths, w)
		return ctx.Factory().Text("x"), nil
	})
	if _, err := Tree(Node{rec, []Node{{rec, []Node{{rec, nil}}}}}).Render(context(realFactory, 6)); err != nil {
		t.Fatal(err)
	}
	if got := fmt.Sprint(widths); got != "[6 2 1]" {
		t.Errorf("label widths = %s, want [6 2 1]", got)
	}
}

func chain(depth int) Node {
	n := node("leaf")
	for range depth {
		n = node("n", n)
	}
	return n
}

func TestTreeDeepChain(t *testing.T) {
	start := time.Now()
	m := render(t, Tree(chain(10000)), 80)
	if m.Height() != 10001 || m.Width() > 80 {
		t.Errorf("size %dx%d", m.Width(), m.Height())
	}
	t.Logf("10000 levels took %v", time.Since(start))
	if m := render(t, Tree(chain(500)), unbounded); m.Width() != 4*500+len("leaf") || m.Height() != 501 {
		t.Errorf("unbounded size %dx%d", m.Width(), m.Height())
	}
}

func TestTreeCopiesNodes(t *testing.T) {
	root := node("a", node("b"), node("c"))
	tree := Tree(root)
	root.Children[0] = node("X")
	checkText(t, tree, unbounded, "a", "├── b", "└── c")
}

func TestTreeErrors(t *testing.T) {
	for i, tree := range []Node{
		{Label: failing()},
		node("a", Node{Label: failing()}),
		node("a", node("b", node("c")), Node{Label: failing()}),
	} {
		_, err := Tree(tree).Render(context(realFactory, 20))
		want := fmt.Sprintf("tree child %d", []int{0, 1, 3}[i])
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("error %v does not mention %q", err, want)
		}
	}
}

func TestTreeContract(t *testing.T) {
	checkChildContract(t, func(c beautify.Composable) beautify.Composable {
		return Tree(Node{c, []Node{{c, []Node{{c, nil}}}, {c, nil}}})
	})
	tree := Node{Text("root", TextStyle(red)), []Node{
		{Text("a\nb", TextStyle(blue)), []Node{{Text("c"), nil}}}, {Text("d"), nil},
	}}
	checkGolden(t, "tree_default", Styled(Tree(tree), onBG), unbounded)
	checkGolden(t, "tree_guide_style", Tree(tree, TreeGuideStyle(red)), unbounded)
	checkGolden(t, "tree_ascii", Tree(tree, TreeGuides(GuidesASCII)), unbounded)
	checkGolden(t, "tree_rounded", Tree(tree, TreeGuides(GuidesRounded)), unbounded)
	checkGolden(t, "tree_max_depth", Tree(tree, TreeMaxDepth(1), TreeGuideStyle(red)), unbounded)
}
