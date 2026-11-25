package dra

import (
	"math/rand/v2"
	"testing"

	"go.xchunk.org/beautify/internal/golden"
)

const family = "\U0001F468‍\U0001F469‍\U0001F467"

func TestTruncate(t *testing.T) {
	dots := text("...")
	tests := []struct {
		name          string
		m             Model
		width         int
		tail          Model
		want          string
		wantW, height int
	}{
		{"empty", Model{}, 3, dots, "", 0, 0},
		{"fits", text("abc"), 3, dots, "abc", 3, 1},
		{"hard cut", text("abcdef"), 4, Model{}, "abcd", 4, 1},
		{"tail", text("abcdef"), 5, dots, "ab...", 5, 1},
		{"tail is exactly the width", text("abcdef"), 3, dots, "...", 3, 1},
		{"tail is wider than the width", text("abcdef"), 2, dots, "ab", 2, 1},
		{"width 1", text("abc"), 1, text("…"), "…", 1, 1},
		{"width 0", text("abc\n\nd"), 0, dots, "\n\n", 0, 3},
		{"width 0 drops zero-width clusters", long("​a"), 0, long("​"), "", 0, 1},
		{"negative", text("abc"), -4, dots, "", 0, 1},
		{"only long lines get the tail", text("ab\nabcd\nabc"), 3, text("~"), "ab\nab~\nabc", 3, 3},
		{"empty lines", text("\n"), 1, dots, "\n", 0, 2},
		{"multi-line tail", text("abcd"), 3, text("~\nX"), "ab~", 3, 1},
		{"unfilled slot", blankFrame(), 1, Model{}, "<", 1, 1},
		{"filled slot", blankFrame().Fill(text("abc")), 3, Model{}, "<ab", 3, 1},
		{"cut inside a wide cluster", long("a漢字"), 4, Model{}, "a漢", 3, 1},
		{"wide cluster does not fit the tail", long("漢字漢"), 4, long("…"), "漢…", 3, 1},
		{"wide cluster at width 1", long("漢"), 1, Model{}, "", 0, 1},
		{"nothing is kept after the cut", long("a漢b"), 2, Model{}, "a", 1, 1},
		{"joiner sequence stays whole", long("a" + family + "b"), 3, Model{}, "a" + family, 3, 1},
		{"joiner sequence goes whole", long("a" + family + "b"), 2, Model{}, "a", 1, 1},
		{"flag", long("\U0001F1E6\U0001F1E7\U0001F1E8"), 2, Model{}, "\U0001F1E6\U0001F1E7", 2, 1},
		{"combining mark stays with its base", long("éé"), 1, Model{}, "é", 1, 1},
		{"tab is four cells", text("a\tb"), 3, Model{}, "a  ", 3, 1},
		{"invalid bytes", text("\xff\xfe"), 1, Model{}, "�", 1, 1},
		{"mixed modes", Concat(text("漢"), long("漢漢")), 3, Model{}, "漢漢", 3, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkModel(t, Truncate(tt.m, tt.width, tt.tail), tt.want, tt.wantW, tt.height)
		})
	}
}

func TestTruncateHeight(t *testing.T) {
	m := text("a\nbb\nccc")
	checkModel(t, TruncateHeight(m, 2), "a\nbb", 2, 2)
	checkModel(t, TruncateHeight(m, 3), "a\nbb\nccc", 3, 3)
	checkModel(t, TruncateHeight(m, 9), "a\nbb\nccc", 3, 3)
	checkModel(t, TruncateHeight(m, 0), "", 0, 0)
	checkModel(t, TruncateHeight(m, -1), "", 0, 0)
	checkModel(t, TruncateHeight(Model{}, 2), "", 0, 0)
	checkModel(t, TruncateHeight(blankFrame(), 1), "<>", 2, 1)
}

func TestTruncateGolden(t *testing.T) {
	f := NewFactory()
	m := Concat(f.Styled("red", Style{FG: Red}), f.Styled("green", Style{FG: Green}))
	golden.Assert(t, "truncate_tail", NewEngine(ProfileANSI16).Append(nil,
		Truncate(m, 6, f.Styled("…", Style{Attrs: Faint}))))
}

func TestTruncateProperties(t *testing.T) {
	rng := rand.New(rand.NewPCG(19, 20))
	for range 1500 {
		m, tail := randomModel(rng, 1), randomModel(rng, 0)
		w := rng.IntN(10) - 1
		got := Truncate(m, w, tail)
		if got.Height() != m.Height() {
			t.Fatalf("height %d, want %d", got.Height(), m.Height())
		}
		if got.Width() > max(w, 0) {
			t.Fatalf("Truncate(%q, %d, %q) = %q is %d wide", m.String(), w, tail.String(), got.String(), got.Width())
		}
		checkOutput(t, got, ProfileANSI16)
	}
}

func FuzzTruncate(f *testing.F) {
	f.Add("a漢́\nfamily", "…", 3, true)
	f.Fuzz(func(t *testing.T, s, tail string, w int, inLong bool) {
		f := NewFactory()
		if inLong {
			f = f.With(WithMode(ModeLong))
		}
		m := f.Text(s)
		got := Truncate(m, w, f.Text(tail))
		if got.Width() > max(w, 0) || got.Height() != m.Height() {
			t.Fatalf("%dx%d for width %d and height %d", got.Width(), got.Height(), w, m.Height())
		}
	})
}

func BenchmarkTruncate(b *testing.B) {
	m, tail := hundredLines(), text("…")
	b.ReportAllocs()
	for range b.N {
		Truncate(m, 20, tail)
	}
}
