package blocks

import (
	"strings"
	"testing"

	"go.xchunk.org/beautify"
	"go.xchunk.org/beautify/dra"
)

var borderSets = map[string]BorderSet{
	"rounded": BorderRounded, "square": BorderSquare, "double": BorderDouble,
	"heavy": BorderHeavy, "ascii": BorderASCII,
}

func TestBorder(t *testing.T) {
	c := raw("a\nbbb")
	checkText(t, Border(c), unbounded,
		"╭─────╮",
		"│ a   │",
		"│ bbb │",
		"╰─────╯")
	checkText(t, Border(c), 40,
		"╭─────╮",
		"│ a   │",
		"│ bbb │",
		"╰─────╯")
	checkText(t, Border(c, BorderPadding(0, 0)), unbounded,
		"╭───╮",
		"│a  │",
		"│bbb│",
		"╰───╯")
	checkText(t, Border(raw("a"), BorderPadding(2, 1)), unbounded,
		"╭─────╮",
		"│     │",
		"│  a  │",
		"│     │",
		"╰─────╯")
	checkText(t, Border(raw("a"), BorderPadding(-1, -1), nil), unbounded,
		"╭─╮",
		"│a│",
		"╰─╯")
	checkText(t, Border(wrapping("aa bb")), 6,
		"╭────╮",
		"│ aa │",
		"│ bb │",
		"╰────╯")
	checkText(t, Border(raw("")), unbounded, "╭──╮", "╰──╯")
	checkText(t, Border(raw(""), BorderPadding(0, 3)), unbounded, "╭╮", "╰╯")
	checkText(t, Border(raw("a\n")), unbounded,
		"╭───╮",
		"│ a │",
		"│   │",
		"╰───╯")
	checkText(t, Border(raw("x")), 1, "╭", "│", "╰")
	checkText(t, Border(Text("x")), 1, "╭", "╰")
	checkText(t, Border(Text("x"), BorderPadding(0, 0)), 2, "╭╮", "╰╯")
	checkEmpty(t, Border(c), 0)
}

func TestBorderTitle(t *testing.T) {
	edge := func(c beautify.Composable, width int) string {
		return strings.Split(render(t, c, width).String(), "\n")[0]
	}
	ten := raw("0123456789") // inner width 12
	tests := []struct {
		name string
		opts []BorderOption
		want string
	}{
		{"left", []BorderOption{BorderTitle("hi")}, "╭─ hi ───────╮"},
		{"center", []BorderOption{BorderTitle("hi"), BorderTitleAlign(dra.AlignCenter)}, "╭──── hi ────╮"},
		{"center, odd", []BorderOption{BorderTitle("hey"), BorderTitleAlign(dra.AlignCenter)}, "╭─── hey ────╮"},
		{"right", []BorderOption{BorderTitle("hi"), BorderTitleAlign(dra.AlignRight)}, "╭─────── hi ─╮"},
		{"empty", []BorderOption{BorderTitle("")}, "╭────────────╮"},
		{"one edge cell left", []BorderOption{BorderTitle("123456789")}, "╭─ 123456789 ╮"},
		{"one edge cell right", []BorderOption{BorderTitle("123456789"), BorderTitleAlign(dra.AlignRight)}, "╭ 123456789 ─╮"},
		{"fills the edge", []BorderOption{BorderTitle("1234567890")}, "╭ 1234567890 ╮"},
		{"truncated", []BorderOption{BorderTitle("12345678901")}, "╭ 123456789… ╮"},
		{"newlines", []BorderOption{BorderTitle("a\nb\r\nc")}, "╭─ a b c ────╮"},
		{"caption", []BorderOption{BorderCaption("ok")}, "╭─────── ok ─╮"},
		{"caption ignores the title alignment", []BorderOption{BorderCaption("ok"), BorderTitleAlign(dra.AlignCenter)}, "╭─────── ok ─╮"},
		{"title and caption", []BorderOption{BorderTitle("hi"), BorderCaption("ok")}, "╭─ hi ── ok ─╮"},
		{"one cell between", []BorderOption{BorderTitle("hi"), BorderCaption("abc")}, "╭─ hi ─ abc ─╮"},
		{"caption does not fit", []BorderOption{BorderTitle("hi"), BorderCaption("abcd")}, "╭─ hi ───────╮"},
		{"centered title and caption", []BorderOption{BorderTitle("h"), BorderTitleAlign(dra.AlignCenter), BorderCaption("k")}, "╭──── h ─ k ─╮"},
		{"right title excludes the caption", []BorderOption{BorderTitle("h"), BorderTitleAlign(dra.AlignRight), BorderCaption("k")}, "╭──────── h ─╮"},
		{"caption alone fills the edge", []BorderOption{BorderCaption("1234567890")}, "╭ 1234567890 ╮"},
		{"caption is never truncated", []BorderOption{BorderCaption("12345678901")}, "╭────────────╮"},
		{"caption newlines", []BorderOption{BorderCaption("a\nb")}, "╭────── a b ─╮"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := edge(Border(ten, tt.opts...), unbounded); got != tt.want {
				t.Errorf("got  %s\nwant %s", got, tt.want)
			}
		})
	}
	// Labels never widen the box.
	narrow := []struct {
		c    beautify.Composable
		want string
	}{
		{Border(raw("abc"), BorderTitle("long title")), "╭ lo… ╮"},
		{Border(raw("a"), BorderTitle("long title")), "╭ … ╮"},
		{Border(raw("a"), BorderTitle("t"), BorderPadding(0, 0)), "╭─╮"},
		{Border(raw(""), BorderTitle("t"), BorderCaption("c")), "╭──╮"},
	}
	for _, tt := range narrow {
		if got := edge(tt.c, unbounded); got != tt.want {
			t.Errorf("got %s, want %s", got, tt.want)
		}
	}
}

func TestBorderFooter(t *testing.T) {
	ten := raw("0123456789") // inner width 12
	tests := []struct {
		name   string
		footer dra.Model
		want   string
	}{
		{"none", dra.Model{}, "╰────────────╯"},
		{"right end", realFactory.Text("ok"), "╰─────── ok ─╯"},
		{"first line only", realFactory.Text("ok\nmore"), "╰─────── ok ─╯"},
		{"one edge cell on each side", realFactory.Text("12345678"), "╰─ 12345678 ─╯"},
		{"does not fit", realFactory.Text("123456789"), "╰────────────╯"},
		{"blank line", dra.Blank(0, 1), "╰────────────╯"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines := strings.Split(render(t, Border(ten, BorderFooter(tt.footer)), unbounded).String(), "\n")
			if got := lines[len(lines)-1]; got != tt.want {
				t.Errorf("got  %s\nwant %s", got, tt.want)
			}
		})
	}
}

func TestBorderCustomSet(t *testing.T) {
	set := BorderSet{TopLeft: "ab", Top: "", TopRight: "\n", Left: "漢", Right: "|", BottomLeft: "x\ny", Bottom: "-", BottomRight: "+"}
	m := renderWith(t, longFactory, Border(raw("z"), BorderStyle(set), BorderPadding(0, 0)), unbounded)
	if got, want := m.String(), "a  \n z|\nx-+"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	checkText(t, Border(raw("z"), BorderStyle(BorderSet{})), unbounded, "     ", "  z  ", "     ")
}

func TestBorderSets(t *testing.T) {
	joins := map[string][2]string{
		"rounded": {"├", "┤"}, "square": {"├", "┤"}, "double": {"╠", "╣"}, "heavy": {"┣", "┫"}, "ascii": {"+", "+"},
	}
	for name, set := range borderSets {
		checkGolden(t, "border_set_"+name, Border(Text("a\nbb"), BorderStyle(set)), unbounded)
		if got := [2]string{set.LeftJoin, set.RightJoin}; got != joins[name] {
			t.Errorf("%s joins = %q, want %q", name, got, joins[name])
		}
	}
}

func TestBorderGolden(t *testing.T) {
	body := Text("content\nof the wider box", TextStyle(red))
	footer := dra.Concat(realFactory.Styled("●", red), realFactory.Text(" "), realFactory.Styled("●", dra.Style{FG: dra.Green}))
	gradient := BorderGradient(dra.NewGradient(dra.AxisHorizontal, dra.RGB(255, 0, 0), dra.RGB(0, 0, 255)))
	labels := []BorderOption{BorderTitle("name"), BorderCaption("ok"), BorderFooter(footer)}
	with := func(opts ...BorderOption) []BorderOption { return append(opts, labels...) }
	tests := map[string][]BorderOption{
		"default":        nil,
		"color":          {BorderColor(blue)},
		"padding":        {BorderPadding(2, 1)},
		"labels":         labels,
		"labels_color":   with(BorderColor(blue.WithBG(dra.Yellow))),
		"label_styles":   with(BorderColor(blue), BorderTitleStyle(red), BorderCaptionStyle(dra.Style{Attrs: dra.Underline})),
		"title_center":   {BorderTitle("name"), BorderTitleAlign(dra.AlignCenter)},
		"gradient":       {gradient},
		"gradient_color": with(gradient, BorderColor(blue.WithBG(dra.Yellow)), BorderTitleStyle(red)),
	}
	for name, opts := range tests {
		checkGolden(t, "border_"+name, Border(body, opts...), unbounded)
	}
	checkGolden(t, "border_narrow", Border(body, BorderTitle("long title"), BorderColor(red)), 8)
}

func TestBorderInheritsBackground(t *testing.T) {
	checkGolden(t, "border_on_background", Styled(Border(Text("a\nbb")), onBG), unbounded)
}

// lineOf returns line y of m rendered in true color.
func lineOf(m dra.Model, y int) string {
	return dra.NewEngine(dra.ProfileTrueColor).String(dra.Lines(m)[y])
}

func TestBorderGradient(t *testing.T) {
	red, blue := dra.RGB(255, 0, 0), dra.RGB(0, 0, 255)
	g := BorderGradient(dra.NewGradient(dra.AxisVertical, red, blue))
	m := render(t, Border(raw("x"), g, BorderPadding(0, 0), BorderColor(dra.Style{FG: dra.Green, Attrs: dra.Bold})), unbounded)
	for y, want := range []string{
		"\x1b[0;1;38;2;255;0;0m╭─╮\x1b[0m",
		"\x1b[0;1;38;2;128;0;128m│\x1b[0mx\x1b[0;1;38;2;128;0;128m│\x1b[0m",
		"\x1b[0;1;38;2;0;0;255m╰─╯\x1b[0m",
	} {
		if got := lineOf(m, y); got != want {
			t.Errorf("line %d: got %q, want %q", y, got, want)
		}
	}

	// Labels follow the gradient unless they have their own foreground.
	footer := dra.Concat(realFactory.Text("a"), realFactory.Styled("b", dra.Style{FG: dra.Green}), realFactory.Styled("c", dra.Style{Attrs: dra.Italic}))
	m = render(t, Border(raw("0123456"), g, BorderTitle("t"), BorderTitleStyle(dra.Style{FG: dra.Green}), BorderCaption("c"),
		BorderCaptionStyle(dra.Style{Attrs: dra.Bold}), BorderFooter(footer)), unbounded)
	if got, want := lineOf(m, 0), "\x1b[0;38;2;255;0;0m╭─ \x1b[0;32mt\x1b[0;38;2;255;0;0m ─ \x1b[0;1;38;2;255;0;0mc\x1b[0;38;2;255;0;0m ─╮\x1b[0m"; got != want {
		t.Errorf("top edge:\n got %q\nwant %q", got, want)
	}
	if got, want := lineOf(m, 2), "\x1b[0;38;2;0;0;255m╰─── a\x1b[0;32mb\x1b[0;3;38;2;0;0;255mc\x1b[0;38;2;0;0;255m ─╯\x1b[0m"; got != want {
		t.Errorf("bottom edge:\n got %q\nwant %q", got, want)
	}

	// The footer color depends on the column.
	h := BorderGradient(dra.NewGradient(dra.AxisHorizontal, dra.RGB(0, 0, 0), dra.RGB(100, 0, 0)))
	m = render(t, Border(raw("0123456"), h, BorderFooter(realFactory.Text("ab"))), unbounded)
	if got := lineOf(m, 2); !strings.Contains(got, "38;2;60;0;0ma") || !strings.Contains(got, "38;2;70;0;0mb") {
		t.Errorf("bottom edge: %q", got)
	}

	// No stops means no gradient.
	plain := render(t, Border(raw("x"), BorderColor(dra.Style{FG: dra.Green})), unbounded)
	none := render(t, Border(raw("x"), BorderColor(dra.Style{FG: dra.Green}), BorderGradient(dra.Gradient{})), unbounded)
	e := dra.NewEngine(dra.ProfileTrueColor)
	if e.String(plain) != e.String(none) {
		t.Error("an empty gradient changes the output")
	}
}

func TestBorderContract(t *testing.T) {
	gradient := dra.NewGradient(dra.AxisDiagonal, dra.Red, dra.Blue)
	build := func(c beautify.Composable) beautify.Composable {
		return Border(c, BorderTitle("漢字 title"), BorderCaption("ok"), BorderFooter(longFactory.Text("漢")), BorderGradient(gradient))
	}
	checkChildContract(t, build)
	checkChildContract(t, func(c beautify.Composable) beautify.Composable { return Border(c, BorderPadding(3, 1)) })
	checkChildErrors(t, "border", 0, build)
	for name, child := range children {
		for _, width := range []int{6, 12, unbounded} {
			m := renderWith(t, longFactory, build(child), width)
			t.Run(name, func(t *testing.T) { checkRectangle(t, m) })
		}
	}
}
