package blocks

import (
	"testing"

	"go.xchunk.org/beautify"
)

func TestText(t *testing.T) {
	const s = "the quick brown fox"
	checkText(t, Text(s), unbounded, s)
	checkText(t, Text(s), 10, "the quick", "brown fox")
	checkText(t, Text(s, TextOverflow(OverflowTruncate)), 10, "the quick…")
	checkText(t, Text(s, TextOverflow(OverflowTruncate)), 1, "…")
	checkText(t, Text(s, TextOverflow(OverflowTruncate), TextEllipsis("...")), 10, "the qui...")
	checkText(t, Text(s, TextOverflow(OverflowTruncate), TextEllipsis("")), 4, "the ")
	checkText(t, Text(s, TextOverflow(OverflowKeep)), 10, "the quick ")
	checkText(t, Text(s, TextOverflow(OverflowKeep)), unbounded, s)
	checkText(t, Text(s, TextOverflow(Overflow(9))), 10, "the quick", "brown fox")
	checkText(t, Text("a b\nccc d"), 3, "a b", "ccc", "d")
	checkText(t, Text("abc\nd", TextOverflow(OverflowTruncate)), 2, "a…", "d")
	checkText(t, Text("a\n"), 5, "a", "")
	checkText(t, Text("\n\n"), 5, "", "", "")
	checkText(t, Text("a\tb\r\nc\x1b\xff"), unbounded, "a    b", "c\ufffd\ufffd")
	checkText(t, Text("x", nil), 5, "x")
	checkEmpty(t, Text(""), unbounded)
	checkEmpty(t, Text(""), 5)
	checkEmpty(t, Text(s), 0)
}

func TestTextWideAtWidthOne(t *testing.T) {
	if m := renderWith(t, longFactory, Text("漢a"), 1); m.Width() > 1 {
		t.Errorf("%q is %d wide", m.String(), m.Width())
	}
}

func TestTextGolden(t *testing.T) {
	checkGolden(t, "text_default", Text("the quick brown fox"), 10)
	checkGolden(t, "text_style", Text("the quick brown fox", TextStyle(blue)), 10)
	checkGolden(t, "text_truncate", Text("the quick brown fox", TextStyle(red), TextOverflow(OverflowTruncate)), 10)
	checkGolden(t, "text_ellipsis", Text("the quick brown fox", TextStyle(red), TextOverflow(OverflowTruncate), TextEllipsis("~")), 10)
	checkGolden(t, "text_keep", Text("the quick brown fox", TextStyle(red), TextOverflow(OverflowKeep)), 10)
}

func TestTextContract(t *testing.T) {
	texts := []string{"", "ab", "incomprehensibilities", "one\ntwo words\n\nlast", "漢字 e\u0301 \U0001F468\u200d\U0001F469", "\x1b[31m\t\r"}
	for _, s := range texts {
		for _, o := range []Overflow{OverflowWrap, OverflowTruncate, OverflowKeep} {
			checkContract(t, s, Text(s, TextOverflow(o)))
		}
	}
	var _ beautify.Composable = Text("")
}
