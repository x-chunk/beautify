package ansi

import "testing"

func TestDowngradeLevels(t *testing.T) {
	full := Style{FG: rgb(255, 0, 0), BG: idx256(21), Attrs: Bold | Strike}
	tests := []struct {
		name  string
		in    Style
		level Level
		want  Style
	}{
		{"plain drops everything", full, Plain, Style{}},
		{"unknown level is plain", full, TrueColor + 1, Style{}},
		{"no color keeps attributes", full, NoColor, Style{Attrs: Bold | Strike}},
		{"true color is unchanged", full, TrueColor, full},
		{"256 converts rgb only", Style{FG: rgb(255, 0, 0), BG: idx16(4)}, ANSI256, Style{FG: idx256(196), BG: idx16(4)}},
		{"256 keeps 256", Style{FG: idx256(21)}, ANSI256, Style{FG: idx256(21)}},
		{"16 converts rgb and 256", full, ANSI16, Style{FG: idx16(9), BG: idx16(12), Attrs: Bold | Strike}},
		{"16 keeps 16", Style{FG: idx16(5)}, ANSI16, Style{FG: idx16(5)}},
		{"unknown attribute bits are dropped", Style{Attrs: 1<<7 | Bold}, TrueColor, Style{Attrs: Bold}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Downgrade(tt.in, tt.level); got != tt.want {
				t.Errorf("Downgrade() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestDowngradeCanonicalizes(t *testing.T) {
	for _, level := range []Level{NoColor, ANSI16, ANSI256, TrueColor} {
		if got := Downgrade(Style{FG: Color{Kind: Default}, BG: Color{Kind: Default}}, level); got != (Style{}) {
			t.Errorf("level %d: Default = %+v, want None", level, got)
		}
		if got := Downgrade(Style{FG: Color{Kind: None, Index: 3, R: 1}}, level); got != (Style{}) {
			t.Errorf("level %d: None with stray fields = %+v, want None", level, got)
		}
	}
	for _, level := range []Level{ANSI16, ANSI256, TrueColor} {
		for n := range uint8(16) {
			if got := Downgrade(Style{FG: idx256(n)}, level).FG; got != idx16(n) {
				t.Errorf("level %d: Indexed256 %d = %+v, want Indexed16", level, n, got)
			}
		}
	}
}

func TestRGBTo256(t *testing.T) {
	tests := []struct {
		name string
		in   Color
		want uint8
	}{
		{"black", rgb(0, 0, 0), 16},
		{"white", rgb(255, 255, 255), 231},
		{"red", rgb(255, 0, 0), 196},
		{"exact cube color", rgb(95, 135, 175), 67},
		{"exact gray", rgb(8, 8, 8), 232},
		{"last gray", rgb(238, 238, 238), 255},
		{"channel tie goes to the lower level", rgb(115, 0, 0), 52},
		{"gray tie goes to the lower index", rgb(13, 13, 13), 232},
		{"distance between cube and gray", rgb(246, 246, 246), 255},
		{"cube wins above the ramp", rgb(248, 248, 248), 231},
		{"cube closer than gray", rgb(96, 96, 96), 59},
		// (4,4,4) is equally close to cube black and gray 8.
		{"cube wins a tie with gray", rgb(4, 4, 4), 16},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Downgrade(Style{FG: tt.in}, ANSI256).FG
			if got != idx256(tt.want) {
				t.Errorf("got %+v, want index %d", got, tt.want)
			}
		})
	}
}

func TestTo16(t *testing.T) {
	tests := []struct {
		in   Color
		want uint8
	}{
		{rgb(0, 0, 0), 0},
		{rgb(255, 255, 255), 15},
		{rgb(128, 0, 0), 1},
		{rgb(192, 192, 192), 7},
		{rgb(128, 128, 128), 8},
		{rgb(0, 255, 255), 14},
		{rgb(64, 0, 0), 0}, // equally close to black and maroon
		{idx256(196), 9},
		{idx256(232), 0},
		{idx256(255), 15},
	}
	for _, tt := range tests {
		if got := Downgrade(Style{BG: tt.in}, ANSI16).BG; got != idx16(tt.want) {
			t.Errorf("%+v -> %+v, want index %d", tt.in, got, tt.want)
		}
	}
}

func TestDowngradeExhaustive(t *testing.T) {
	for n := range 256 {
		c := idx256(uint8(n))
		got := Downgrade(Style{FG: c}, ANSI16).FG
		if got.Kind != Indexed16 || got.Index > 15 {
			t.Errorf("index %d -> %+v, want a valid Indexed16", n, got)
		}
		if n < 16 {
			continue
		}
		r, g, b := ToRGB(c)
		if got := Downgrade(Style{FG: rgb(r, g, b)}, ANSI256).FG; got != c {
			t.Errorf("index %d -> rgb(%d,%d,%d) -> %+v", n, r, g, b, got)
		}
	}
}

func TestNearbyColorsDowngradeToEqualStyles(t *testing.T) {
	a := Downgrade(Style{FG: rgb(255, 0, 0)}, ANSI256)
	b := Downgrade(Style{FG: rgb(254, 1, 1)}, ANSI256)
	if a != b {
		t.Errorf("%+v != %+v", a, b)
	}
}

func TestToRGB(t *testing.T) {
	tests := []struct {
		in      Color
		r, g, b uint8
	}{
		{rgb(1, 2, 3), 1, 2, 3},
		{idx16(1), 0x80, 0, 0},
		{idx16(17), 0x80, 0, 0},
		{idx256(9), 0xff, 0, 0},
		{idx256(16), 0, 0, 0},
		{idx256(67), 95, 135, 175},
		{idx256(255), 238, 238, 238},
		{Color{}, 0, 0, 0},
		{Color{Kind: Default}, 0, 0, 0},
	}
	for _, tt := range tests {
		if r, g, b := ToRGB(tt.in); r != tt.r || g != tt.g || b != tt.b {
			t.Errorf("ToRGB(%+v) = %d,%d,%d, want %d,%d,%d", tt.in, r, g, b, tt.r, tt.g, tt.b)
		}
	}
}
