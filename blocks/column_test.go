package blocks

import (
	"testing"

	"go.xchunk.org/beautify"
	"go.xchunk.org/beautify/dra"
)

func TestColumn(t *testing.T) {
	items := func(c ...beautify.Composable) []beautify.Composable { return c }
	a, b := raw("a"), raw("bbb")
	checkText(t, Column(items(a, b)), unbounded, "a", "bbb")
	checkText(t, Column(items(a, b), ColumnGap(1)), unbounded, "a", "", "bbb")
	checkText(t, Column(items(a, b), ColumnGap(-1), nil), unbounded, "a", "bbb")
	checkText(t, Column(items(raw(""), a, raw(""), b, raw("")), ColumnGap(1)), unbounded, "a", "", "bbb")
	checkText(t, Column(items(a, raw("\n"), b), ColumnGap(1)), unbounded, "a", "", "", "", "", "bbb")
	checkText(t, Column(items(a, b), ColumnAlign(dra.AlignRight)), unbounded, "  a", "bbb")
	checkText(t, Column(items(a, b), ColumnAlign(dra.AlignCenter)), unbounded, " a", "bbb")
	checkText(t, Column(items(a, b)), 2, "a", "bb")
	checkText(t, Column(items(wrapping("aa bb"), a)), 2, "aa", "bb", "a")
	checkEmpty(t, Column(nil), unbounded)
	checkEmpty(t, Column(items(raw(""), raw("")), ColumnGap(2)), unbounded)
	checkEmpty(t, Column(items(a)), 0)

	in := items(a, b)
	col := Column(in)
	in[0] = raw("X")
	checkText(t, col, unbounded, "a", "bbb")
}

func TestColumnContract(t *testing.T) {
	build := func(c beautify.Composable) beautify.Composable {
		return Column([]beautify.Composable{raw("first"), c, c}, ColumnGap(1), ColumnAlign(dra.AlignCenter))
	}
	checkChildContract(t, build)
	checkChildErrors(t, "column", 1, build)
	items := []beautify.Composable{Text("a", TextStyle(red)), Text("bbb", TextStyle(blue))}
	checkGolden(t, "column_default", Styled(Column(items), onBG), unbounded)
	checkGolden(t, "column_gap", Column(items, ColumnGap(1)), unbounded)
	checkGolden(t, "column_right", Column(items, ColumnAlign(dra.AlignRight)), unbounded)
}
