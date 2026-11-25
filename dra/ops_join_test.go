package dra

import (
	"math/rand/v2"
	"testing"

	"go.xchunk.org/beautify/internal/golden"
)

func TestStack(t *testing.T) {
	tests := []struct {
		name          string
		align         Align
		ms            []Model
		want          string
		width, height int
	}{
		{"none", AlignLeft, nil, "", 0, 0},
		{"one", AlignLeft, []Model{text("a")}, "a", 1, 1},
		{"only empty", AlignCenter, []Model{{}, text("")}, "", 0, 0},
		{"left", AlignLeft, []Model{text("a"), text("bbbb"), text("cc")}, "a\nbbbb\ncc", 4, 3},
		{"center", AlignCenter, []Model{text("a"), text("bbbb"), text("cc")}, " a\nbbbb\n cc", 4, 3},
		{"right", AlignRight, []Model{text("a"), text("bbbb"), text("cc")}, "   a\nbbbb\n  cc", 4, 3},
		{"unknown align is left", Align(9), []Model{text("a"), text("bb")}, "a\nbb", 2, 2},
		{"lines align one by one", AlignRight, []Model{text("a\nbbb"), text("cc")}, "  a\nbbb\n cc", 3, 3},
		{"empty models add nothing", AlignLeft, []Model{text("a"), {}, text("b")}, "a\nb", 1, 2},
		{"empty lines count", AlignLeft, []Model{text("a"), Blank(0, 1), text("b")}, "a\n\nb", 1, 3},
		{"trailing newline", AlignLeft, []Model{text("a\n"), text("b")}, "a\n\nb", 1, 3},
		{"unfilled slot", AlignRight, []Model{blankFrame(), text("abc")}, " <>\nabc", 3, 2},
		{"mixed modes", AlignRight, []Model{long("漢"), text("漢")}, "漢\n 漢", 2, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkModel(t, Stack(tt.align, tt.ms...), tt.want, tt.width, tt.height)
		})
	}
}

func TestJoin(t *testing.T) {
	tall, short := text("a\nbb\nc\nd"), text("x")
	tests := []struct {
		name          string
		valign        VAlign
		ms            []Model
		want          string
		width, height int
	}{
		{"none", AlignTop, nil, "", 0, 0},
		{"one is not padded", AlignTop, []Model{text("a\nbb")}, "a\nbb", 2, 2},
		{"only empty", AlignTop, []Model{{}, {}}, "", 0, 0},
		{"one line each", AlignTop, []Model{text("a"), text("b")}, "ab", 2, 1},
		{"tallest first", AlignTop, []Model{tall, short}, "a x\nbb\nc \nd ", 3, 4},
		{"tallest last", AlignTop, []Model{short, tall}, "xa\n bb\n c\n d", 3, 4},
		{"tallest in the middle", AlignTop, []Model{short, tall, short}, "xa x\n bb\n c \n d ", 4, 4},
		{"middle", AlignMiddle, []Model{short, tall}, " a\nxbb\n c\n d", 3, 4},
		{"bottom", AlignBottom, []Model{short, tall}, " a\n bb\n c\nxd", 3, 4},
		{"unknown align is top", VAlign(9), []Model{short, tall}, "xa\n bb\n c\n d", 3, 4},
		{"empty is skipped", AlignTop, []Model{text("a"), {}, text("b")}, "ab", 2, 1},
		{"trailing spaces are kept", AlignTop, []Model{text("a  "), text("b  ")}, "a  b  ", 6, 1},
		{"zero-width line", AlignTop, []Model{Blank(0, 1), text("a")}, "a", 1, 1},
		{"unfilled slot", AlignTop, []Model{blankFrame(), text("a")}, "<>a", 3, 1},
		{"mixed modes", AlignTop, []Model{long("漢\na"), text("漢\nb")}, "漢漢\na b", 3, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkModel(t, Join(tt.valign, tt.ms...), tt.want, tt.width, tt.height)
		})
	}
}

func TestStackGolden(t *testing.T) {
	f := NewFactory()
	m := Stack(AlignCenter, f.Styled("wide line", Style{FG: Red}), f.Styled("x", Style{BG: Blue}))
	golden.Assert(t, "stack_center", NewEngine(ProfileANSI16).Append(nil, m))
	// Padding is unstyled, so it takes the background of the outer model.
	outer := f.MustNew(WithStyle(Style{BG: Green})).Fill(m)
	golden.Assert(t, "stack_center_in_slot", NewEngine(ProfileANSI16).Append(nil, outer))
}

func TestJoinGolden(t *testing.T) {
	f := NewFactory()
	m := Join(AlignMiddle, f.Styled("a\nb\nc", Style{FG: Red}), f.Styled("|", Style{Attrs: Bold}), f.Text("x\ny"))
	golden.Assert(t, "join_middle", NewEngine(ProfileANSI16).Append(nil, m))
}

func TestStackManyStyles(t *testing.T) {
	f := NewFactory()
	var ms []Model
	for i := range 300 {
		ms = append(ms, f.Styled("x", Style{FG: RGB(uint8(i), uint8(i/256), 7)}))
	}
	m := Stack(AlignLeft, ms...)
	seen := map[Style]bool{}
	for _, l := range m.lines {
		seen[l.spans[0].style] = true
	}
	if len(seen) != 300 || m.Height() != 300 {
		t.Errorf("%d styles, height %d", len(seen), m.Height())
	}
	checkOutput(t, m, ProfileTrueColor)
}

func TestStackAndJoinProperties(t *testing.T) {
	rng := rand.New(rand.NewPCG(15, 16))
	for range 800 {
		var ms []Model
		heights, tallest, widths, widest := 0, 0, 0, 0
		for range rng.IntN(4) {
			m := randomModel(rng, 1)
			ms = append(ms, m)
			heights += m.Height()
			tallest = max(tallest, m.Height())
			widths += m.Width()
			widest = max(widest, m.Width())
		}
		s := Stack(Align(rng.IntN(3)), ms...)
		if s.Height() != heights || s.Width() != widest {
			t.Fatalf("Stack: %dx%d, want %dx%d", s.Width(), s.Height(), widest, heights)
		}
		j := Join(VAlign(rng.IntN(3)), ms...)
		if j.Height() != tallest || j.Width() != widths {
			t.Fatalf("Join: %dx%d, want %dx%d", j.Width(), j.Height(), widths, tallest)
		}
		checkOutput(t, s, ProfileANSI256)
		checkOutput(t, j, ProfileANSI256)
		if len(ms) > 0 {
			if got := Stack(AlignLeft, Lines(ms[0])...); got.String() != ms[0].String() {
				t.Fatalf("Stack(Lines) = %q, want %q", got.String(), ms[0].String())
			}
		}
	}
}

func TestJoinOfSingleLinesEqualsConcat(t *testing.T) {
	ms := []Model{text("a"), Blank(2, 1), text("b")}
	if got, want := Join(AlignTop, ms...).String(), Concat(ms...).String(); got != want {
		t.Errorf("Join() = %q, Concat() = %q", got, want)
	}
}

func BenchmarkJoin(b *testing.B) {
	m := hundredLines()
	b.ReportAllocs()
	for range b.N {
		Join(AlignTop, m, m)
	}
}

// hundredLines is a composed model of 100 styled lines.
func hundredLines() Model {
	var b builder
	for range 100 {
		b.addLine(firstLine(fiveRuns()))
		b.endLine()
	}
	return b.model()
}
