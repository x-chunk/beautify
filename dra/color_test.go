package dra

import (
	"errors"
	"testing"

	"go.xchunk.org/beautify/internal/ansi"
)

func TestColorConstructors(t *testing.T) {
	tests := []struct {
		name string
		got  Color
		want ansi.Color
	}{
		{"zero", Color{}, ansi.Color{}},
		{"ANSI", ANSI(9), ansi.Color{Kind: ansi.Indexed16, Index: 9}},
		{"ANSI256", ANSI256(200), ansi.Color{Kind: ansi.Indexed256, Index: 200}},
		{"RGB", RGB(1, 2, 3), ansi.Color{Kind: ansi.RGB, R: 1, G: 2, B: 3}},
		{"TerminalDefault", TerminalDefault, ansi.Color{Kind: ansi.Default}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got.c != tt.want {
				t.Errorf("got %+v, want %+v", tt.got.c, tt.want)
			}
			if got, want := tt.got.IsSet(), tt.name != "zero"; got != want {
				t.Errorf("IsSet() = %v, want %v", got, want)
			}
		})
	}
}

func TestANSIWrapsAround(t *testing.T) {
	if ANSI(17) != ANSI(1) {
		t.Error("ANSI(17) != ANSI(1)")
	}
}

func TestNamedColors(t *testing.T) {
	named := []Color{
		Black, Red, Green, Yellow, Blue, Magenta, Cyan, White,
		BrightBlack, BrightRed, BrightGreen, BrightYellow,
		BrightBlue, BrightMagenta, BrightCyan, BrightWhite,
	}
	for i, c := range named {
		if c != ANSI(uint8(i)) {
			t.Errorf("color %d = %+v", i, c)
		}
	}
}

func TestHex(t *testing.T) {
	valid := map[string]Color{
		"#ff8000": RGB(255, 128, 0),
		"ff8000":  RGB(255, 128, 0),
		"#FF8000": RGB(255, 128, 0),
		"#abc":    RGB(0xaa, 0xbb, 0xcc),
		"AbC":     RGB(0xaa, 0xbb, 0xcc),
		"000":     RGB(0, 0, 0),
	}
	for in, want := range valid {
		if got, err := Hex(in); err != nil || got != want {
			t.Errorf("Hex(%q) = %+v, %v, want %+v", in, got, err, want)
		}
	}
	invalid := []string{
		"", "#", "f", "ff", "ffff", "#fffff", "fffffff", "#ggg", "gggggg",
		"##fff", "+ff", "#-1-1-1", "fff ", "ффф", "яя",
	}
	for _, in := range invalid {
		if got, err := Hex(in); !errors.Is(err, ErrInvalidColor) || got != (Color{}) {
			t.Errorf("Hex(%q) = %+v, %v, want ErrInvalidColor", in, got, err)
		}
	}
}

func TestAttrBitsMatchEncoder(t *testing.T) {
	pairs := map[Attr]uint8{
		Bold: ansi.Bold, Faint: ansi.Faint, Italic: ansi.Italic,
		Underline: ansi.Underline, Blink: ansi.Blink,
		Reverse: ansi.Reverse, Strike: ansi.Strike,
	}
	for a, want := range pairs {
		if uint8(a) != want {
			t.Errorf("attribute %d != encoder bit %d", a, want)
		}
	}
}
