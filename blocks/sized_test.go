package blocks

import (
	"testing"

	"go.xchunk.org/beautify"
)

func TestSized(t *testing.T) {
	c := raw("a\nbbb")
	checkText(t, Sized(c), unbounded, "a", "bbb")
	checkText(t, Sized(c, MinWidth(5)), unbounded, "a    ", "bbb  ")
	checkText(t, Sized(c, MaxWidth(2)), unbounded, "a", "bb")
	checkText(t, Sized(c, Width(2)), unbounded, "a ", "bb")
	checkText(t, Sized(c, Width(4)), unbounded, "a   ", "bbb ")
	checkText(t, Sized(c, MinHeight(3)), unbounded, "a", "bbb", "")
	checkText(t, Sized(c, MaxHeight(1)), unbounded, "a")
	checkText(t, Sized(c, Height(3), Width(1)), unbounded, "a", "b", " ")
	checkText(t, Sized(c, Height(1)), unbounded, "a")
	checkText(t, Sized(c, nil), unbounded, "a", "bbb")

	// The child gets the smaller of the bound and the maximum.
	checkText(t, Sized(wrapping("aa bb"), MaxWidth(3)), unbounded, "aa", "bb")
	checkText(t, Sized(wrapping("aa bb"), MaxWidth(9)), 3, "aa", "bb")
	checkText(t, Sized(wrapping("aa bb"), MaxWidth(3)), 9, "aa", "bb")

	// The bound wins over the minimum, the maximum wins over the minimum.
	checkText(t, Sized(c, MinWidth(9)), 4, "a   ", "bbb ")
	checkText(t, Sized(c, MinWidth(1<<60)), 4, "a   ", "bbb ")
	checkText(t, Sized(c, MinWidth(9), MaxWidth(2)), unbounded, "a ", "bb")
	checkText(t, Sized(c, MinHeight(9), MaxHeight(1)), unbounded, "a")

	checkText(t, Sized(raw(""), Width(2), Height(2)), unbounded, "  ", "  ")
	checkEmpty(t, Sized(raw(""), Width(2)), unbounded)
	checkEmpty(t, Sized(c, MaxHeight(0)), unbounded)
	checkEmpty(t, Sized(c, Height(-1)), unbounded)
	checkText(t, Sized(c, MaxWidth(0)), unbounded, "", "")
	checkText(t, Sized(c, MaxWidth(-3)), 10, "", "")
	checkEmpty(t, Sized(Text("a"), MaxWidth(0)), unbounded)
	checkEmpty(t, Sized(c, Width(3)), 0)
}

func TestSizedContract(t *testing.T) {
	for _, opts := range [][]SizedOption{nil, {Width(7), Height(2)}, {MinWidth(30)}, {MaxWidth(4), MinHeight(3)}} {
		build := func(c beautify.Composable) beautify.Composable { return Sized(c, opts...) }
		checkChildContract(t, build)
		checkChildErrors(t, "sized", 0, build)
	}
	build := func(c beautify.Composable) beautify.Composable { return Sized(c, Width(9), Height(3)) }
	checkRectangle(t, renderWith(t, longFactory, build(children["wide"]), unbounded))
	checkGolden(t, "sized_grow", Styled(Sized(Text("a\nbbb", TextStyle(red)), Width(5), Height(3)), onBG), unbounded)
	checkGolden(t, "sized_shrink", Sized(Text("aaaa bbbb cccc", TextStyle(red)), MaxWidth(6), MaxHeight(2)), unbounded)
}
