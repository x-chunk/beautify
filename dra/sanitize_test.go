package dra

import (
	"slices"
	"testing"
	"unicode/utf8"
)

func TestNormalizeText(t *testing.T) {
	const bad = "�"
	tests := []struct {
		name     string
		in       string
		clean    string // expected with sanitization
		rawClean string // expected without; empty means the same
	}{
		{"plain", "hello", "hello", ""},
		{"empty", "", "", ""},
		{"unicode", "привет 漢\U0001F600", "привет 漢\U0001F600", ""},
		{"newline", "a\nb", "a\nb", ""},
		{"crlf", "a\r\nb", "a\nb", ""},
		{"lone cr", "a\rb", "a" + bad + "b", "a\rb"},
		{"cr crlf", "a\r\r\nb", "a" + bad + "\nb", "a\r\nb"},
		{"trailing cr", "a\r", "a" + bad, "a\r"},
		{"tab", "a\tb", "a    b", ""},
		{"esc", "\x1b[31mred", bad + "[31mred", "\x1b[31mred"},
		{"nul", "\x00", bad, "\x00"},
		{"bell", "a\x07", "a" + bad, "a\x07"},
		{"del", "\x7f", bad, "\x7f"},
		{"c1 csi", "\u009b31m", bad + "31m", "\u009b31m"},
		{"c1 bounds", "\u0080\u009f ", bad + bad + " ", "\u0080\u009f "},
		{"bidi override", "a‮b", "a" + bad + "b", "a‮b"},
		{"bidi bounds", " ‪‮ ⁥⁦⁩⁪",
			bad + bad + bad + " ⁥" + bad + bad + "⁪",
			" ‪‮ ⁥⁦⁩⁪"},
		{"line separator", "a b", "a" + bad + "b", "a b"},
		{"invalid utf-8", "a\xff\xfeb", "a" + bad + bad + "b", ""},
		{"truncated sequence", "\xe2\x82", bad + bad, ""},
		{"c1 as a raw byte", "\x9b", bad, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := normalize(tt.in, nil, 4, true); got != tt.clean {
				t.Errorf("sanitized = %q, want %q", got, tt.clean)
			}
			want := tt.rawClean
			if want == "" {
				want = tt.clean
			}
			if got, _ := normalize(tt.in, nil, 4, false); got != want {
				t.Errorf("raw = %q, want %q", got, want)
			}
		})
	}
}

func TestNormalizeTabWidth(t *testing.T) {
	if got, _ := normalize("\ta\t", nil, 2, true); got != "  a  " {
		t.Errorf("got %q", got)
	}
}

func TestNormalizeColormap(t *testing.T) {
	tests := []struct {
		name string
		in   string
		cmap []int8
		want []int8
	}{
		{"one to one", "a\x1bb", []int8{0, 1, 2}, []int8{0, 1, 2}},
		{"tab repeats its entry", "a\tb", []int8{0, 1, 2}, []int8{0, 1, 1, 1, 1, 2}},
		{"crlf keeps the newline entry", "a\r\nb", []int8{0, 1, 2, 3}, []int8{0, 2, 3}},
		{"invalid byte is one rune", "\xff\xfe", []int8{0, 1}, []int8{0, 1}},
		{"multi-byte runes", "я漢", []int8{0, 1}, []int8{0, 1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text, got := normalize(tt.in, tt.cmap, 4, true)
			if !slices.Equal(got, tt.want) {
				t.Errorf("cmap = %v, want %v", got, tt.want)
			}
			if len(got) != utf8.RuneCountInString(text) {
				t.Errorf("%d entries for %d runes", len(got), utf8.RuneCountInString(text))
			}
		})
	}
}

func TestUnifyClusters(t *testing.T) {
	tests := []struct {
		name string
		in   string
		cmap []int8
		want []int8
	}{
		{"no clusters", "ab", []int8{0, 1}, []int8{0, 1}},
		{"combining mark", "éx", []int8{0, 1, 2}, []int8{0, 0, 2}},
		{"flag", "\U0001F1E6\U0001F1E7\U0001F1E8", []int8{0, 1, 2}, []int8{0, 0, 2}},
		{"nil", "é", nil, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := slices.Clone(tt.cmap)
			unifyClusters(tt.in, got)
			if !slices.Equal(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStringWidth(t *testing.T) {
	tests := []struct {
		name       string
		in         string
		real, long int
	}{
		{"empty", "", 0, 0},
		{"ascii", "hello", 5, 5},
		{"widest line", "a\nabc\nab", 3, 3},
		{"trailing newline", "ab\n", 2, 2},
		{"only newlines", "\n\n", 0, 0},
		{"tab", "a\tb", 6, 6},
		{"crlf", "ab\r\nc", 2, 2},
		{"cjk", "漢字", 2, 4},
		{"combining", "é", 2, 1},
		{"escape is visible", "\x1b[m", 3, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StringWidth(tt.in, ModeReal); got != tt.real {
				t.Errorf("real = %d, want %d", got, tt.real)
			}
			if got := StringWidth(tt.in, ModeLong); got != tt.long {
				t.Errorf("long = %d, want %d", got, tt.long)
			}
		})
	}
}
