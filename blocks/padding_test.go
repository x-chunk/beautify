package blocks

import (
	"testing"

	"go.xchunk.org/beautify"
)

func TestPadding(t *testing.T) {
	c := raw("a\nbbb")
	checkText(t, Padding(c), unbounded, "a  ", "bbb")
	checkText(t, Padding(c, PaddingAll(1)), unbounded, "     ", " a   ", " bbb ", "     ")
	checkText(t, Padding(c, PaddingX(2)), unbounded, "  a    ", "  bbb  ")
	checkText(t, Padding(c, PaddingY(1)), unbounded, "   ", "a  ", "bbb", "   ")
	checkText(t, Padding(c, PaddingSides(1, 2, 0, 3)), unbounded, "        ", "   a    ", "   bbb  ")
	checkText(t, Padding(c, PaddingAll(1), PaddingX(0)), unbounded, "   ", "a  ", "bbb", "   ")
	checkText(t, Padding(c, PaddingAll(-2)), unbounded, "a  ", "bbb")
	checkText(t, Padding(c, PaddingSides(-1, 1, -1, -1), nil), unbounded, "a   ", "bbb ")

	// The child gets what the padding leaves.
	checkText(t, Padding(wrapping("aa bb"), PaddingX(1)), 5, " aa ", " bb ")
	checkText(t, Padding(c, PaddingX(1)), 4, " a  ", " bb ")
	checkText(t, Padding(Text("abc"), PaddingX(2)), 3)
	checkText(t, Padding(Text("abc"), PaddingAll(2)), 3, "   ", "   ", "   ", "   ")

	checkText(t, Padding(raw(""), PaddingAll(1)), unbounded, "  ", "  ")
	checkEmpty(t, Padding(raw(""), PaddingX(1)), unbounded)
	checkEmpty(t, Padding(raw("")), unbounded)
	checkEmpty(t, Padding(c, PaddingAll(1)), 0)
}

func TestPaddingGolden(t *testing.T) {
	checkGolden(t, "padding_all", Styled(Padding(Text("a\nbbb", TextStyle(red)), PaddingAll(1)), onBG), unbounded)
	checkGolden(t, "padding_sides", Styled(Padding(Text("a"), PaddingSides(0, 1, 1, 3)), onBG), unbounded)
}

func TestPaddingContract(t *testing.T) {
	build := func(c beautify.Composable) beautify.Composable { return Padding(c, PaddingAll(2)) }
	checkChildContract(t, build)
	checkChildErrors(t, "padding", 0, build)
	checkRectangle(t, renderWith(t, longFactory, build(children["wide"]), unbounded))
	checkRectangle(t, renderWith(t, longFactory, build(children["multi-line"]), 9))
}
