package blocks

import (
	"testing"

	"go.xchunk.org/beautify"
	"go.xchunk.org/beautify/dra"
)

func TestSections(t *testing.T) {
	two := []Section{{"System", raw("OS      Ubuntu 24.04")}, {"Hardware", raw("CPU     Ryzen 7 5800X")}}
	checkText(t, Sections(two, SectionsCaption("host")), unbounded,
		"╭─ System ─────── host ─╮",
		"│ OS      Ubuntu 24.04  │",
		"├─ Hardware ────────────┤",
		"│ CPU     Ryzen 7 5800X │",
		"╰───────────────────────╯")
	plain := []Section{{"", raw("a")}, {"", raw("b\nccc")}, {"", raw("")}, {"", raw("d")}}
	checkText(t, Sections(plain), unbounded,
		"╭─────╮",
		"│ a   │",
		"├─────┤",
		"│ b   │",
		"│ ccc │",
		"├─────┤",
		"├─────┤",
		"│ d   │",
		"╰─────╯")
	checkText(t, Sections(plain[:2], SectionsPadding(0, 1), SectionsBorder(BorderASCII)), unbounded,
		"+---+",
		"|   |",
		"|a  |",
		"|   |",
		"+---+",
		"|   |",
		"|b  |",
		"|ccc|",
		"|   |",
		"+---+")
	checkText(t, Sections([]Section{{"long title", wrapping("aa bb")}, {"t", raw("c")}}, nil), 6,
		"╭ l… ╮",
		"│ aa │",
		"│ bb │",
		"├─ t ┤",
		"│ c  │",
		"╰────╯")
	checkText(t, Sections([]Section{{"x", raw("12345")}, {"y", raw("")}}, SectionsTitleAlign(dra.AlignRight)), unbounded,
		"╭─── x ─╮",
		"│ 12345 │",
		"├─── y ─┤",
		"╰───────╯")
	checkText(t, Sections([]Section{{"", raw("12345")}, {"", raw("b")}}, SectionsFooter(realFactory.Text("ok"))), unbounded,
		"╭───────╮",
		"│ 12345 │",
		"├───────┤",
		"│ b     │",
		"╰── ok ─╯")
	checkText(t, Sections([]Section{{"", raw("a")}, {"", raw("b")}}, SectionsBorder(BorderSet{LeftJoin: "漢字", RightJoin: ""}), SectionsPadding(-1, -1)), unbounded,
		"   ", " a ", "漢  ", " b ", "   ")
	checkText(t, Sections(plain[:2]), 1, "╭", "│", "├", "│", "│", "╰")
	checkEmpty(t, Sections(nil), unbounded)
	checkEmpty(t, Sections(plain), 0)

	in := []Section{{"", raw("a")}}
	s := Sections(in)
	in[0].Body = raw("X")
	checkText(t, s, unbounded, "╭───╮", "│ a │", "╰───╯")
}

func TestSectionsGradient(t *testing.T) {
	g := dra.NewGradient(dra.AxisVertical, dra.RGB(0, 0, 0), dra.RGB(100, 0, 0))
	m := render(t, Sections([]Section{{"", raw("a")}, {"t", raw("b")}}, SectionsBorderGradient(g),
		SectionsPadding(0, 0), SectionsBorderStyle(dra.Style{FG: dra.Green})), unbounded)
	if got, want := lineOf(m, 2), "\x1b[0;38;2;50;0;0m├─┤\x1b[0m"; got != want {
		t.Errorf("divider = %q, want %q", got, want)
	}
	if got, want := lineOf(m, 4), "\x1b[0;38;2;100;0;0m╰─╯\x1b[0m"; got != want {
		t.Errorf("bottom edge = %q, want %q", got, want)
	}
}

func TestSectionsContract(t *testing.T) {
	build := func(c beautify.Composable) beautify.Composable {
		return Sections([]Section{{"漢字", raw("first")}, {"title", c}, {"", c}}, SectionsCaption("cap"), SectionsPadding(1, 1),
			SectionsFooter(longFactory.Text("漢")), SectionsBorderGradient(dra.NewGradient(dra.AxisDiagonal, dra.Red, dra.Blue)))
	}
	checkChildContract(t, build)
	checkChildErrors(t, "sections", 1, build)
	for _, child := range children {
		for _, width := range []int{8, 14, unbounded} {
			checkRectangle(t, renderWith(t, longFactory, build(child), width))
		}
	}

	sections := []Section{{"System", Text("OS: Linux", TextStyle(red))}, {"Hardware", Text("CPU: some long name")}}
	footer := dra.Concat(realFactory.Styled("●", red), realFactory.Text(" ok"))
	for name, opts := range map[string][]SectionsOption{
		"default":      nil,
		"border":       {SectionsBorder(BorderDouble)},
		"border_style": {SectionsBorderStyle(blue), SectionsCaption("host"), SectionsFooter(footer)},
		"label_styles": {SectionsBorderStyle(blue), SectionsCaption("host"), SectionsTitleStyle(red), SectionsCaptionStyle(dra.Style{Attrs: dra.Underline})},
		"title_center": {SectionsTitleAlign(dra.AlignCenter)},
		"padding":      {SectionsPadding(2, 1)},
		"gradient":     {SectionsBorderGradient(dra.NewGradient(dra.AxisVertical, dra.RGB(255, 0, 0), dra.RGB(0, 0, 255))), SectionsFooter(footer)},
	} {
		checkGolden(t, "sections_"+name, Sections(sections, opts...), unbounded)
	}
	checkGolden(t, "sections_narrow", Sections(sections, SectionsCaption("host")), 14)
}
