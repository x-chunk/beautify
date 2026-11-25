package blocks

import (
	"testing"

	"go.xchunk.org/beautify"
)

func TestList(t *testing.T) {
	items := func(c ...beautify.Composable) []beautify.Composable { return c }
	two := items(raw("a"), raw("b\nc"))
	checkText(t, List(two), unbounded, "• a", "• b", "  c")
	checkText(t, List(two, ListBullet("->")), unbounded, "-> a", "-> b", "   c")
	checkText(t, List(two, ListBullet("")), unbounded, " a", " b", " c")
	checkText(t, List(two, ListBullet("*\n*")), unbounded, "* * a", "* * b", "    c")
	checkText(t, List(two, ListOrdered()), unbounded, "1. a", "2. b", "   c")
	checkText(t, List(two, ListOrdered(), ListStart(9)), unbounded, " 9. a", "10. b", "    c")
	checkText(t, List(two, ListOrdered(), ListStart(-1)), unbounded, "-1. a", " 0. b", "    c")
	checkText(t, List(two, ListStart(9), nil), unbounded, "• a", "• b", "  c")
	checkText(t, List(two, ListGap(1)), unbounded, "• a", "", "• b", "  c")
	checkText(t, List(two, ListGap(-1)), unbounded, "• a", "• b", "  c")
	checkText(t, List(items(wrapping("aa bb"))), 4, "• aa", "  bb")
	checkText(t, List(items(raw(""), raw("x"))), unbounded, "• ", "• x")
	checkText(t, List(items(raw("abc"))), 3, "• a")
	checkText(t, List(items(Text("abc"))), 2, "• ")
	checkText(t, List(items(Text("abc")), ListOrdered()), 1, "1")
	checkEmpty(t, List(nil), unbounded)
	checkEmpty(t, List(two), 0)

	in := items(raw("a"))
	list := List(in)
	in[0] = raw("X")
	checkText(t, list, unbounded, "• a")
}

func TestListContract(t *testing.T) {
	build := func(c beautify.Composable) beautify.Composable {
		return List([]beautify.Composable{raw("first"), c}, ListGap(1))
	}
	checkChildContract(t, build)
	checkChildContract(t, func(c beautify.Composable) beautify.Composable {
		return List([]beautify.Composable{c, c}, ListOrdered(), ListStart(99))
	})
	checkChildErrors(t, "list", 1, build)
	if got, want := renderWith(t, longFactory, List([]beautify.Composable{raw("a\nb")}, ListBullet("漢")), unbounded).String(), "漢 a\n   b"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	items := []beautify.Composable{Text("one", TextStyle(red)), Text("two\nlines")}
	checkGolden(t, "list_default", Styled(List(items), onBG), unbounded)
	checkGolden(t, "list_ordered", List(items, ListOrdered(), ListStart(9), ListMarkerStyle(blue)), unbounded)
	checkGolden(t, "list_bullet", List(items, ListBullet("->"), ListMarkerStyle(blue)), unbounded)
	checkGolden(t, "list_gap", List(items, ListGap(1)), unbounded)
}
