package dra

import (
	"math/rand/v2"
	"testing"
)

func TestStyleOver(t *testing.T) {
	parent := Style{FG: Red, BG: Blue, Attrs: Bold}
	tests := []struct {
		name          string
		child, parent Style
		want          Style
	}{
		{"both zero", Style{}, Style{}, Style{}},
		{"unset inherits", Style{}, parent, parent},
		{"set wins", Style{FG: Green}, parent, Style{FG: Green, BG: Blue, Attrs: Bold}},
		{"attributes combine", Style{Attrs: Italic}, parent, Style{FG: Red, BG: Blue, Attrs: Bold | Italic}},
		{"zero parent", parent, Style{}, parent},
		{"terminal default stops inheritance", Style{FG: TerminalDefault}, parent, Style{FG: TerminalDefault, BG: Blue, Attrs: Bold}},
		{"terminal default is inherited", Style{}, Style{BG: TerminalDefault}, Style{BG: TerminalDefault}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.child.Over(tt.parent); got != tt.want {
				t.Errorf("Over() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func randomColor(rng *rand.Rand) Color {
	switch rng.IntN(5) {
	case 0:
		return Color{}
	case 1:
		return TerminalDefault
	case 2:
		return ANSI(uint8(rng.IntN(16)))
	case 3:
		return ANSI256(uint8(rng.IntN(256)))
	default:
		return RGB(uint8(rng.IntN(256)), uint8(rng.IntN(256)), uint8(rng.IntN(256)))
	}
}

func randomStyle(rng *rand.Rand) Style {
	return Style{FG: randomColor(rng), BG: randomColor(rng), Attrs: Attr(rng.IntN(128))}
}

func TestStyleOverIsAssociative(t *testing.T) {
	rng := rand.New(rand.NewPCG(1, 1))
	for range 2000 {
		a, b, c := randomStyle(rng), randomStyle(rng), randomStyle(rng)
		if left, right := a.Over(b.Over(c)), a.Over(b).Over(c); left != right {
			t.Fatalf("a=%+v b=%+v c=%+v: %+v != %+v", a, b, c, left, right)
		}
	}
}

func TestStyleIsZero(t *testing.T) {
	if !(Style{}).IsZero() {
		t.Error("zero style is not zero")
	}
	for _, s := range []Style{{FG: Red}, {BG: TerminalDefault}, {Attrs: Bold}} {
		if s.IsZero() {
			t.Errorf("%+v is zero", s)
		}
	}
}

func TestStyleBuilders(t *testing.T) {
	if got := Foreground(Red); got != (Style{FG: Red}) {
		t.Errorf("Foreground() = %+v", got)
	}
	if got := Background(Red); got != (Style{BG: Red}) {
		t.Errorf("Background() = %+v", got)
	}
	base := Style{Attrs: Bold}
	got := base.WithFG(Red).WithBG(Blue).WithAttrs(Italic)
	if want := (Style{FG: Red, BG: Blue, Attrs: Bold | Italic}); got != want {
		t.Errorf("builders = %+v, want %+v", got, want)
	}
	if base != (Style{Attrs: Bold}) {
		t.Errorf("receiver changed to %+v", base)
	}
}
