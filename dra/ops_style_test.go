package dra

import (
	"math/rand/v2"
	"testing"

	"go.xchunk.org/beautify/internal/golden"
)

func TestRestyle(t *testing.T) {
	f := NewFactory()
	e := NewEngine(ProfileANSI16)
	tests := []struct {
		name string
		m    Model
		st   Style
		want string
	}{
		{"empty", Model{}, Style{FG: Red}, ""},
		{"zero style", f.Styled("a", Style{FG: Red}), Style{}, "\x1b[0;31ma\x1b[0m"},
		{"fills unset", text("a"), Style{FG: Red}, "\x1b[0;31ma\x1b[0m"},
		{"own foreground wins", f.Styled("a", Style{FG: Green}), Style{FG: Red, BG: Blue}, "\x1b[0;32;44ma\x1b[0m"},
		{"attributes add up", f.Styled("a", Style{Attrs: Bold}), Style{Attrs: Italic}, "\x1b[0;1;3ma\x1b[0m"},
		{"colors padding", Pad(text("a"), 2, AlignLeft), Style{BG: Blue}, "\x1b[0;44ma \x1b[0m"},
		{"unfilled slot", blankFrame(), Style{FG: Red}, "\x1b[0;31m<>\x1b[0m"},
		{"filled slot", blankFrame().Fill(f.Styled("x", Style{FG: Green})), Style{FG: Red},
			"\x1b[0;31m<\x1b[0;32mx\x1b[0;31m>\x1b[0m"},
		{"background per line", text("a\nb"), Style{BG: Blue}, "\x1b[0;44ma\x1b[0m\n\x1b[0;44mb\x1b[0m"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Restyle(tt.m, tt.st)
			if out := e.String(got); out != tt.want {
				t.Errorf("got %q, want %q", out, tt.want)
			}
			checkModel(t, got, tt.m.String(), tt.m.Width(), tt.m.Height())
		})
	}
}

func TestRestyleMergesSpans(t *testing.T) {
	f := NewFactory()
	m := Restyle(Concat(f.Text("a"), f.Styled("b", Style{FG: Red})), Style{FG: Red})
	if n := len(m.lines[0].spans); n != 1 {
		t.Errorf("%d spans, want 1", n)
	}
}

func TestRestyleGolden(t *testing.T) {
	f := NewFactory()
	m := Join(AlignTop, f.Styled("a\nb", Style{FG: Red}), f.Text("xx"))
	golden.Assert(t, "restyle", NewEngine(ProfileANSI16).Append(nil, Restyle(m, Style{FG: Green, BG: Blue, Attrs: Bold})))
}

func TestRestyleEqualsSlotInheritance(t *testing.T) {
	rng := rand.New(rand.NewPCG(23, 24))
	e := NewEngine(ProfileTrueColor)
	for range 500 {
		m, st := randomModel(rng, 2), randomStyle(rng)
		want := e.String(NewFactory().MustNew(WithStyle(st)).Fill(m))
		if got := e.String(Restyle(m, st)); got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	}
}

func TestRestyleFunc(t *testing.T) {
	f := NewFactory()
	e := NewEngine(ProfileANSI16)
	checker := func(x, y int) Style {
		if (x+y)%2 == 0 {
			return Style{FG: Red}
		}
		return Style{}
	}
	tests := []struct {
		name string
		m    Model
		fn   func(x, y int) Style
		want string
	}{
		{"nil function", f.Styled("a", Style{FG: Red}), nil, "\x1b[0;31ma\x1b[0m"},
		{"empty model", Model{}, checker, ""},
		{"zero style changes nothing", f.Styled("ab", Style{FG: Red}),
			func(int, int) Style { return Style{} }, "\x1b[0;31mab\x1b[0m"},
		{"by position", text("ab\ncd"), checker, "\x1b[0;31ma\x1b[0mb\nc\x1b[0;31md\x1b[0m"},
		{"wins over the model", f.Styled("ab", Style{FG: Green, BG: Blue}), checker,
			"\x1b[0;31;44ma\x1b[0;32;44mb\x1b[0m"},
		{"unfilled slot", blankFrame(), checker, "\x1b[0;31m<\x1b[0m>"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RestyleFunc(tt.m, tt.fn)
			if out := e.String(got); out != tt.want {
				t.Errorf("got %q, want %q", out, tt.want)
			}
			checkModel(t, got, tt.m.String(), tt.m.Width(), tt.m.Height())
		})
	}
}

func TestRestyleFuncCallsOncePerCluster(t *testing.T) {
	type pos struct{ x, y int }
	var got []pos
	m := Concat(long("漢é"+family), text("ab\n漢"))
	RestyleFunc(m, func(x, y int) Style {
		got = append(got, pos{x, y})
		return Style{}
	})
	want := []pos{{0, 0}, {2, 0}, {3, 0}, {5, 0}, {6, 0}, {0, 1}}
	if len(got) != len(want) {
		t.Fatalf("calls %v, want %v", got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("call %d at %v, want %v", i, got[i], want[i])
		}
	}
}

func TestRestyleFuncEveryCellDiffers(t *testing.T) {
	m := RestyleFunc(Blank(40, 20), func(x, y int) Style {
		return Style{BG: RGB(uint8(x), uint8(y), 0)}
	})
	for _, l := range m.lines {
		if len(l.spans) != 40 {
			t.Fatalf("%d spans, want 40", len(l.spans))
		}
	}
	checkOutput(t, m, ProfileTrueColor)
}

func TestRestyleFuncMergesEqualCells(t *testing.T) {
	f := NewFactory()
	m := RestyleFunc(Concat(f.Text("ab"), f.Styled("cd", Style{FG: Red})), func(x, y int) Style {
		return Style{FG: Green}
	})
	if n := len(m.lines[0].spans); n != 1 {
		t.Errorf("%d spans, want 1", n)
	}
}

func TestRestyleFuncGradientGolden(t *testing.T) {
	g := NewGradient(AxisDiagonal, RGB(0xff, 0x80, 0), RGB(0, 0x80, 0xff))
	m := text("#####\n#####\n#####")
	m = RestyleFunc(m, func(x, y int) Style { return Style{FG: g.At(x, y, 5, 3)} })
	for _, p := range allProfiles {
		golden.Assert(t, "gradient_"+p.String(), NewEngine(p).Append(nil, m))
	}
}

func TestRestyleFuncProperties(t *testing.T) {
	rng := rand.New(rand.NewPCG(25, 26))
	for range 500 {
		m := randomModel(rng, 2)
		styles := []Style{{}, randomStyle(rng), randomStyle(rng)}
		got := RestyleFunc(m, func(x, y int) Style { return styles[(x+2*y)%3] })
		if got.String() != m.String() || got.size != m.size {
			t.Fatalf("got %q %+v, want %q %+v", got.String(), got.size, m.String(), m.size)
		}
		checkOutput(t, got, ProfileTrueColor)
	}
}
