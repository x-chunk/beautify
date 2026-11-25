package dra

import (
	"testing"

	"go.xchunk.org/beautify/internal/golden"
)

func TestPrefixLines(t *testing.T) {
	tests := []struct {
		name           string
		m, first, rest Model
		want           string
		width, height  int
	}{
		{"empty model", Model{}, text("> "), text("  "), "", 0, 0},
		{"one line", text("a"), text("> "), text("| "), "> a", 3, 1},
		{"several lines", text("a\nb\nc"), text("> "), text("| "), "> a\n| b\n| c", 3, 3},
		{"empty prefix", text("a\nb"), Model{}, Model{}, "a\nb", 1, 2},
		{"multi-line prefix", text("a\nb"), text(">\nX"), text("|\nX"), ">a\n|b", 2, 2},
		{"empty lines get a prefix", text("\n"), text(">"), text("|"), ">\n|", 1, 2},
		{"empty line", Blank(0, 1), text(">"), text("|"), ">", 1, 1},
		{"unfilled slots", blankFrame(), blankFrame(), text("|"), "<><>", 4, 1},
		{"mixed modes", long("漢"), text("漢"), Model{}, "漢漢", 3, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkModel(t, PrefixLines(tt.m, tt.first, tt.rest), tt.want, tt.width, tt.height)
		})
	}
}

func TestSuffixLines(t *testing.T) {
	tests := []struct {
		name           string
		m, first, rest Model
		want           string
		width, height  int
	}{
		{"empty model", Model{}, text("<"), text("|"), "", 0, 0},
		{"several lines", text("aa\nb\nc"), text("<"), text("|"), "aa<\nb|\nc|", 3, 3},
		{"empty suffix", text("a\nb"), Model{}, Model{}, "a\nb", 1, 2},
		{"multi-line suffix", text("a\nb"), text("<\nX"), text("|\nX"), "a<\nb|", 2, 2},
		{"empty lines get a suffix", text("\n"), text("<"), text("|"), "<\n|", 1, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkModel(t, SuffixLines(tt.m, tt.first, tt.rest), tt.want, tt.width, tt.height)
		})
	}
}

func TestPrefixLinesGolden(t *testing.T) {
	f := NewFactory()
	m := SuffixLines(
		PrefixLines(f.Styled("a\nb", Style{FG: Red}), f.Styled("> ", Style{Attrs: Bold}), f.Text("  ")),
		f.Styled(" <", Style{FG: Green}), f.Text(" |"))
	golden.Assert(t, "prefix_suffix_lines", NewEngine(ProfileANSI16).Append(nil, m))
}
