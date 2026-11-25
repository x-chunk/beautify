package dra

import (
	"math/rand/v2"
	"strings"
	"testing"

	"go.xchunk.org/beautify/internal/golden"
)

func TestWrap(t *testing.T) {
	tests := []struct {
		name          string
		m             Model
		width         int
		mode          WrapMode
		want          string
		wantW, height int
	}{
		{"empty", Model{}, 3, WrapWords, "", 0, 0},
		{"zero width", text("a b"), 0, WrapWords, "a b", 3, 1},
		{"negative width", text("a b"), -1, WrapAnywhere, "a b", 3, 1},
		{"fits", text("a b"), 3, WrapWords, "a b", 3, 1},
		{"unfilled slot", blankFrame(), 1, WrapAnywhere, "<\n>", 1, 2},

		{"anywhere", text("abcdefg"), 3, WrapAnywhere, "abc\ndef\ng", 3, 3},
		{"anywhere keeps spaces", text("a bc d"), 2, WrapAnywhere, "a \nbc\n d", 2, 3},
		{"anywhere wide", long("a漢字b"), 3, WrapAnywhere, "a漢\n字b", 3, 2},
		{"anywhere wide at width 1", long("a漢b"), 1, WrapAnywhere, "a\n漢\nb", 2, 3},
		{"anywhere keeps clusters", long("éé"), 1, WrapAnywhere, "é\né", 1, 2},
		{"anywhere joiner sequence", long(family + family), 2, WrapAnywhere, family + "\n" + family, 2, 2},
		{"zero-width clusters never break", long("ab​c"), 2, WrapAnywhere, "ab​\nc", 2, 2},

		{"words", text("the quick brown fox"), 10, WrapWords, "the quick\nbrown fox", 9, 2},
		{"unknown mode wraps words", text("ab cd"), 4, WrapMode(9), "ab\ncd", 2, 2},
		{"word exactly fits", text("ab cd ef"), 5, WrapWords, "ab cd\nef", 5, 2},
		{"empty source lines", text("aa bb\n\ncc"), 2, WrapWords, "aa\nbb\n\ncc", 2, 4},
		{"long word", text("abcdefgh"), 3, WrapWords, "abc\ndef\ngh", 3, 3},
		{"long word after a short one", text("ab cdefgh"), 4, WrapWords, "ab\ncdef\ngh", 4, 3},
		{"last piece stays open", text("abcde f g"), 4, WrapWords, "abcd\ne f\ng", 4, 3},
		{"several spaces at a break", text("ab    cd"), 3, WrapWords, "ab\ncd", 2, 2},
		{"several spaces inside a line", text("a  b cccc"), 4, WrapWords, "a  b\ncccc", 4, 2},
		{"only spaces", text("     "), 3, WrapWords, "", 0, 1},
		{"only spaces that fit", text("   "), 3, WrapWords, "   ", 3, 1},
		{"indentation", text("  ab cd"), 4, WrapWords, "  ab\ncd", 4, 2},
		{"indentation as wide as the width", text("    ab"), 4, WrapWords, "ab", 2, 1},
		{"first word does not fit after the indentation", text("  abc d"), 4, WrapWords, "abc\nd", 3, 2},
		{"long word after the indentation", text(" abcdef"), 4, WrapWords, "abcd\nef", 4, 2},
		{"trailing spaces that fit", text("abc d  "), 3, WrapWords, "abc\nd  ", 3, 2},
		{"trailing spaces that do not fit", text("abc de    "), 3, WrapWords, "abc\nde", 3, 2},
		{"tab is four spaces", text("ab\tcd"), 5, WrapWords, "ab\ncd", 2, 2},
		{"no-break space is part of a word", text("a b c"), 3, WrapWords, "a b\nc", 3, 2},
		{"width 1", text("a bc"), 1, WrapWords, "a\nb\nc", 1, 3},
		{"wide clusters at width 1", long("漢 字"), 1, WrapWords, "漢\n字", 2, 2},
		{"wide word", long("漢字漢 a"), 5, WrapWords, "漢字\n漢 a", 4, 2},
		{"mixed modes", Concat(text("ab "), long("漢字")), 4, WrapWords, "ab\n漢字", 4, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkModel(t, Wrap(tt.m, tt.width, tt.mode), tt.want, tt.wantW, tt.height)
		})
	}
}

func TestWrapGolden(t *testing.T) {
	f := NewFactory()
	m := Concat(f.Styled("one tw", Style{FG: Red}), f.Styled("o three", Style{FG: Green, BG: Blue}), f.Text(" four"))
	e := NewEngine(ProfileANSI16)
	golden.Assert(t, "wrap_words_style_inside_word", e.Append(nil, Wrap(m, 6, WrapWords)))
	golden.Assert(t, "wrap_anywhere_styled", e.Append(nil, Wrap(m, 4, WrapAnywhere)))
}

// removeRunes returns s without the runes in cut.
func removeRunes(s, cut string) string {
	return strings.Map(func(r rune) rune {
		if strings.ContainsRune(cut, r) {
			return -1
		}
		return r
	}, s)
}

func TestWrapProperties(t *testing.T) {
	rng := rand.New(rand.NewPCG(21, 22))
	for range 2000 {
		w := 1 + rng.IntN(8)
		s := randomText(rng, 24)
		got := Wrap(text(s), w, WrapAnywhere)
		if got.Width() > w {
			t.Fatalf("WrapAnywhere(%q, %d) is %d wide", s, w, got.Width())
		}
		if removeRunes(got.String(), "\n") != removeRunes(s, "\n") {
			t.Fatalf("WrapAnywhere(%q, %d) = %q", s, w, got.String())
		}
		words := Wrap(text(s), w, WrapWords)
		if words.Width() > w {
			t.Fatalf("WrapWords(%q, %d) is %d wide", s, w, words.Width())
		}

		m := randomModel(rng, 1)
		for _, mode := range []WrapMode{WrapWords, WrapAnywhere} {
			got := Wrap(m, w, mode)
			if removeRunes(got.String(), " \n") != removeRunes(m.String(), " \n") {
				t.Fatalf("Wrap(%q, %d, %d) = %q", m.String(), w, mode, got.String())
			}
			// Only a wide cluster at width 1 may exceed the width.
			if got.Width() > max(w, 2) {
				t.Fatalf("Wrap(%q, %d, %d) is %d wide", m.String(), w, mode, got.Width())
			}
			if got.Height() < m.Height() {
				t.Fatalf("Wrap(%q, %d, %d) lost lines", m.String(), w, mode)
			}
			checkOutput(t, got, ProfileTrueColor)
		}
	}
}

func FuzzWrap(f *testing.F) {
	f.Add("the quick  brown\nfox", 5, uint8(0), false)
	f.Add("漢字 ́ a", 1, uint8(1), true)
	f.Fuzz(func(t *testing.T, s string, w int, mode uint8, inLong bool) {
		f := NewFactory()
		if inLong {
			f = f.With(WithMode(ModeLong))
		}
		m := f.Text(s)
		got := Wrap(m, w, WrapMode(mode))
		if w > 0 && got.Width() > max(w, 2) {
			t.Fatalf("width %d for %d", got.Width(), w)
		}
		if removeRunes(got.String(), " \n") != removeRunes(m.String(), " \n") {
			t.Fatalf("Wrap(%q, %d) = %q", m.String(), w, got.String())
		}
	})
}

func BenchmarkWrap(b *testing.B) {
	m := hundredLines()
	b.ReportAllocs()
	for range b.N {
		Wrap(m, 20, WrapWords)
	}
}
