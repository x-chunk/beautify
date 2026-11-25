package width

import "testing"

func TestWideTableIsSortedAndDisjoint(t *testing.T) {
	for i, r := range wideRanges {
		if r.lo > r.hi {
			t.Errorf("range %d: %X > %X", i, r.lo, r.hi)
		}
		if i > 0 && r.lo <= wideRanges[i-1].hi {
			t.Errorf("range %d (%X) overlaps or precedes range %d", i, r.lo, i-1)
		}
	}
}

func TestRuneWidth(t *testing.T) {
	tests := []struct {
		name string
		r    rune
		want int
	}{
		{"ascii", 'a', 1},
		{"cyrillic", 'я', 1},
		{"soft hyphen", 0x00AD, 1},
		{"combining acute (Mn)", 0x0301, 0},
		{"enclosing circle (Me)", 0x20DD, 0},
		{"spacing mark (Mc)", 0x0903, 1},
		{"zero-width space", 0x200B, 0},
		{"zero-width joiner", 0x200D, 0},
		{"word joiner", 0x2060, 0},
		{"bom", 0xFEFF, 0},
		{"hangul vowel", 0x1160, 0},
		{"hangul final", 0x11FF, 0},
		{"hangul extended", 0xD7B0, 0},
		{"hangul leading", 0x1100, 2},
		{"hangul leading end", 0x115F, 2},
		{"cjk", '漢', 2},
		{"before first wide", 0x10FF, 1},
		{"ideographic half fill space", 0x303F, 1},
		{"ideographic space", 0x3000, 2},
		{"fullwidth", 0xFF01, 2},
		{"halfwidth katakana", 0xFF66, 1},
		{"emoji", 0x1F600, 2},
		{"between emoji blocks", 0x1F650, 1},
		{"watch", 0x231A, 2},
		{"after watch", 0x231C, 1},
		{"star", 0x2B50, 2},
		{"heart", 0x2764, 1},
		{"ambiguous", 0x00A7, 1},
		{"regional indicator", 0x1F1E6, 1},
		{"plane 2", 0x20000, 2},
		{"plane 3 end", 0x3FFFD, 2},
		{"after plane 3", 0x3FFFE, 1},
		{"replacement", 0xFFFD, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := runeWidth(tt.r); got != tt.want {
				t.Errorf("runeWidth(%U) = %d, want %d", tt.r, got, tt.want)
			}
		})
	}
}
