package dra

import (
	"math/rand/v2"
	"testing"

	"go.xchunk.org/beautify/internal/golden"
)

func TestPad(t *testing.T) {
	tests := []struct {
		name          string
		m             Model
		width         int
		align         Align
		want          string
		wantW, height int
	}{
		{"empty stays empty", Model{}, 5, AlignLeft, "", 0, 0},
		{"left", text("a\nbbb"), 4, AlignLeft, "a   \nbbb ", 4, 2},
		{"right", text("a\nbbb"), 4, AlignRight, "   a\n bbb", 4, 2},
		{"center puts the extra cell right", text("a\nbbb"), 4, AlignCenter, " a  \nbbb ", 4, 2},
		{"unknown align is left", text("a"), 2, Align(9), "a ", 2, 1},
		{"narrower than the content", text("abc\na"), 2, AlignLeft, "abc\na ", 3, 2},
		{"zero", text("a"), 0, AlignRight, "a", 1, 1},
		{"negative", text("a"), -3, AlignCenter, "a", 1, 1},
		{"empty lines", text("\n"), 2, AlignLeft, "  \n  ", 2, 2},
		{"unfilled slot", blankFrame(), 3, AlignRight, " <>", 3, 1},
		{"wide runes", long("漢\na"), 3, AlignLeft, "漢 \na  ", 3, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkModel(t, Pad(tt.m, tt.width, tt.align), tt.want, tt.wantW, tt.height)
		})
	}
}

func TestPadHeight(t *testing.T) {
	tests := []struct {
		name          string
		m             Model
		height        int
		valign        VAlign
		want          string
		width, wantHt int
	}{
		{"empty", Model{}, 2, AlignTop, "\n", 0, 2},
		{"empty and zero", Model{}, 0, AlignTop, "", 0, 0},
		{"empty and negative", Model{}, -1, AlignTop, "", 0, 0},
		{"top", text("a"), 3, AlignTop, "a\n\n", 1, 3},
		{"bottom", text("a"), 3, AlignBottom, "\n\na", 1, 3},
		{"middle puts the extra line at the bottom", text("a"), 4, AlignMiddle, "\na\n\n", 1, 4},
		{"unknown valign is top", text("a"), 2, VAlign(9), "a\n", 1, 2},
		{"never removes lines", text("a\nb\nc"), 2, AlignTop, "a\nb\nc", 1, 3},
		{"negative", text("a"), -1, AlignBottom, "a", 1, 1},
		{"unfilled slot", blankFrame(), 2, AlignTop, "<>\n", 2, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkModel(t, PadHeight(tt.m, tt.height, tt.valign), tt.want, tt.width, tt.wantHt)
		})
	}
}

func TestRepeat(t *testing.T) {
	tests := []struct {
		name  string
		unit  Model
		width int
		want  string
		wantW int
	}{
		{"one cell", text("-"), 4, "----", 4},
		{"two cells", text("-="), 4, "-=-=", 4},
		{"remainder", text("-="), 5, "-=-=-", 5},
		{"narrower than the unit", text("abc"), 2, "ab", 2},
		{"zero", text("-"), 0, "", 0},
		{"negative", text("-"), -2, "", 0},
		{"empty unit", Model{}, 3, "   ", 3},
		{"zero-width unit", Blank(0, 1), 2, "  ", 2},
		{"wide unit", long("漢"), 4, "漢漢", 4},
		{"wide unit and odd width", long("漢"), 5, "漢漢 ", 5},
		{"wide unit that does not fit", long("漢"), 1, " ", 1},
		{"multi-line unit", text("-\n="), 3, "---", 3},
		{"unfilled slot", blankFrame(), 3, "<><", 3},
		{"joiner unit keeps its width", long("a‍"), 3, "a‍a‍a‍", 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkModel(t, Repeat(tt.unit, tt.width), tt.want, tt.wantW, 1)
		})
	}
}

func TestPadGolden(t *testing.T) {
	f := NewFactory()
	e := NewEngine(ProfileANSI16)
	m := Pad(f.Styled("a\nbbb", Style{FG: Red}), 5, AlignCenter)
	golden.Assert(t, "pad_center", e.Append(nil, m))
	outer := f.MustNew(WithStyle(Style{BG: Blue})).Fill(PadHeight(m, 3, AlignTop))
	golden.Assert(t, "pad_center_in_slot", e.Append(nil, outer))
	unit := Concat(f.Styled("=", Style{FG: Green}), f.Styled("-", Style{FG: Red}))
	golden.Assert(t, "repeat_styled", e.Append(nil, Repeat(unit, 5)))
}

func TestPadProperties(t *testing.T) {
	rng := rand.New(rand.NewPCG(17, 18))
	for range 800 {
		m, w := randomModel(rng, 1), rng.IntN(12)-1
		got := Pad(m, w, Align(rng.IntN(3)))
		if got.Height() != m.Height() {
			t.Fatalf("height %d, want %d", got.Height(), m.Height())
		}
		for _, l := range Lines(got) {
			if l.Width() < w {
				t.Fatalf("line %q is narrower than %d", l.String(), w)
			}
		}
		checkOutput(t, got, ProfileTrueColor)
		if r := Repeat(m, w); r.Width() != max(w, 0) || r.Height() != 1 {
			t.Fatalf("Repeat(%q, %d) is %dx%d", m.String(), w, r.Width(), r.Height())
		}
	}
}
