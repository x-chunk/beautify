package blocks

import (
	"bytes"
	"testing"

	"go.xchunk.org/beautify"
	"go.xchunk.org/beautify/dra"
	"go.xchunk.org/beautify/internal/golden"
)

func TestTreeInBorderInRow(t *testing.T) {
	c := Row([]Cell{
		Fixed(6, Text("left\nside")),
		Flex(1, Border(Tree(sampleTree(), TreeGuideStyle(red)), BorderTitle("tree"), BorderColor(blue))),
		Auto(Badge("ok", onBG)),
	}, RowGap(1))
	checkGolden(t, "composition_tree_in_border_in_row", c, 40)
	checkContract(t, "tree in border in row", c)
}

func TestListsInTable(t *testing.T) {
	c := Table(
		TableHeaders(Text("Fruit"), Text("Notes")),
		TableRow(Text("Apple"), List([]beautify.Composable{Text("red or green"), Text("sweet")})),
		TableRow(Text("Kiwi"), List([]beautify.Composable{Text("fuzzy")}, ListOrdered())),
	)
	checkGolden(t, "composition_lists_in_table", c, unbounded)
	checkGolden(t, "composition_lists_in_table_narrow", c, 22)
	checkContract(t, "lists in table", c)
}

func TestBorderInBorder(t *testing.T) {
	c := Border(Border(Text("inner text"), BorderTitle("in"), BorderStyle(BorderDouble)), BorderTitle("outer title"), BorderColor(red))
	checkGolden(t, "composition_border_in_border", c, unbounded)
	checkGolden(t, "composition_border_in_border_narrow", c, 12)
	checkContract(t, "border in border", c)
}

func TestStyledInheritanceThroughBlocks(t *testing.T) {
	c := Styled(Column([]beautify.Composable{
		Border(Text("a\nb", TextStyle(red))),
		KeyValue([]Pair{{"key", Text("value", TextStyle(dra.Style{BG: dra.TerminalDefault}))}}),
	}), onBG)
	checkGolden(t, "composition_inheritance", c, unbounded)
}

func TestOneSectionEqualsBorder(t *testing.T) {
	gradient := dra.NewGradient(dra.AxisDiagonal, dra.Red, dra.Blue)
	footer := dra.Concat(realFactory.Styled("●", red), realFactory.Text(" ok"))
	titleStyle, captionStyle := dra.Style{FG: dra.Yellow}, dra.Style{Attrs: dra.Italic}
	e := dra.NewEngine(dra.ProfileTrueColor)
	for setName, set := range borderSets {
		for childName, child := range children {
			for _, align := range []dra.Align{dra.AlignLeft, dra.AlignCenter, dra.AlignRight} {
				for _, width := range []int{1, 5, 12, 30, unbounded} {
					border := render(t, Border(child,
						BorderStyle(set), BorderColor(blue), BorderGradient(gradient),
						BorderTitle("title"), BorderTitleAlign(align), BorderTitleStyle(titleStyle),
						BorderCaption("cap"), BorderCaptionStyle(captionStyle),
						BorderFooter(footer), BorderPadding(2, 1)), width)
					section := render(t, Sections([]Section{{"title", child}},
						SectionsBorder(set), SectionsBorderStyle(blue), SectionsBorderGradient(gradient),
						SectionsTitleAlign(align), SectionsTitleStyle(titleStyle),
						SectionsCaption("cap"), SectionsCaptionStyle(captionStyle),
						SectionsFooter(footer), SectionsPadding(2, 1)), width)
					if a, b := e.String(border), e.String(section); a != b {
						t.Errorf("%s, %s, align %d, width %d:\nborder   %q\nsections %q", setName, childName, align, width, a, b)
					}
				}
			}
			if a, b := e.String(render(t, Border(child, BorderStyle(set)), unbounded)),
				e.String(render(t, Sections([]Section{{"", child}}, SectionsBorder(set)), unbounded)); a != b {
				t.Errorf("%s, %s, defaults: border %q, sections %q", setName, childName, a, b)
			}
		}
	}
}

func TestComposerRendersBlocks(t *testing.T) {
	tpl := beautify.NewTemplate(
		Border(Text("A report"), BorderTitle("report"), BorderColor(blue)),
		KeyValue([]Pair{{"Status", Badge("ok", onBG)}, {"Owner", Text("a long name that wraps at the bound")}}),
		Rule(RuleStyle(red)),
		Table(TableHeaders(Text("Name"), Text("Qty")), TableRow(Text("Apple"), Text("1"))),
		List([]beautify.Composable{Text("first"), Text("second")}, ListOrdered()),
	).With(beautify.WithDirection(beautify.Stacked))
	for _, p := range []dra.Profile{dra.ProfilePlain, dra.ProfileNoColor, dra.ProfileANSI16, dra.ProfileANSI256, dra.ProfileTrueColor} {
		var buf bytes.Buffer
		c := beautify.NewComposer(&buf, beautify.WithWidth(30), beautify.WithProfile(p))
		if err := c.Render(tpl); err != nil {
			t.Fatal(err)
		}
		golden.Assert(t, "composition_composer_"+p.String(), buf.Bytes())
	}
}

// TestMixedModes places content measured in one mode into blocks that
// measure in the other.
func TestMixedModes(t *testing.T) {
	c := Border(Column([]beautify.Composable{
		Model(longFactory.Text("漢字")), Model(realFactory.Text("漢字")), Text("text"),
	}))
	for _, f := range []*dra.Factory{realFactory, longFactory} {
		checkRectangle(t, renderWith(t, f, c, unbounded))
	}
	checkContract(t, "mixed modes", c)
}
