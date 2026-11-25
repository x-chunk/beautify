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
