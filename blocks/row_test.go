package blocks

import (
	"testing"

	"go.xchunk.org/beautify"
	"go.xchunk.org/beautify/dra"
)

func cells(c ...Cell) []Cell { return c }

func TestRow(t *testing.T) {
	a, b, c := raw("aaa"), raw("bb"), raw("c")
	tests := []struct {
		name  string
		row   beautify.Composable
		width int
		want  []string
	}{
		{"auto", Row(cells(Auto(a), Auto(b))), unbounded, []string{"aaabb"}},
		{"gap", Row(cells(Auto(a), Auto(b), Auto(c)), RowGap(2)), unbounded, []string{"aaa  bb  c"}},
		{"negative gap", Row(cells(Auto(a), Auto(b)), RowGap(-2), nil), unbounded, []string{"aaabb"}},
		{"fixed pads", Row(cells(Fixed(5, a), Auto(b))), unbounded, []string{"aaa  bb"}},
		{"fixed pads the last cell", Row(cells(Auto(a), Fixed(4, b))), unbounded, []string{"aaabb  "}},
		{"fixed cuts", Row(cells(Fixed(2, a), Auto(b))), unbounded, []string{"aabb"}},
		{"fixed with an empty child", Row(cells(Fixed(3, raw("")), Auto(b))), unbounded, []string{"   bb"}},
		{"fixed zero is omitted with its gap", Row(cells(Fixed(0, a), Auto(b), Fixed(-1, c)), RowGap(1)), unbounded, []string{"bb"}},
		{"empty auto is omitted with its gap", Row(cells(Auto(a), Auto(raw("")), Auto(c)), RowGap(1)), unbounded, []string{"aaa c"}},
		{"flex unbounded", Row(cells(Flex(1, a), Flex(2, b))), unbounded, []string{"aaabb"}},

		{"flex shares", Row(cells(Flex(1, a), Flex(1, b))), 10, []string{"aaa  bb   "}},
		{"flex weights", Row(cells(Flex(1, a), Flex(2, b))), 12, []string{"aaa bb      "}},
		{"flex remainder goes left first", Row(cells(Flex(1, a), Flex(1, b), Flex(1, c))), 11, []string{"aaa bb  c  "}},
		{"flex weights below 1", Row(cells(Flex(0, a), Flex(-5, b))), 8, []string{"aaa bb  "}},
		{"huge weights are limited", Row(cells(Flex(1<<62, a), Flex(1<<62, b))), 8, []string{"aaa bb  "}},
		{"huge gap", Row(cells(Auto(a), Auto(b), Auto(c)), RowGap(1<<62)), 8, []string{"aaa"}},
		{"flex gets the rest", Row(cells(Fixed(4, a), Flex(1, b), Auto(c)), RowGap(1)), 12, []string{"aaa  bb    c"}},
		{"flex without room is omitted", Row(cells(Fixed(5, a), Flex(1, b))), 5, []string{"aaa  "}},
		{"fixed is clamped in order", Row(cells(Fixed(4, a), Fixed(4, b))), 6, []string{"aaa bb"}},
		{"auto takes what is left in order", Row(cells(Auto(wrapping("xx yy")), Auto(wrapping("zz")))), 6, []string{"xx yyz", "     z"}},
		{"auto goes before flex", Row(cells(Flex(1, a), Auto(b))), 5, []string{"aaabb"}},
		{"gaps fill the bound", Row(cells(Auto(a), Auto(b), Auto(c)), RowGap(2)), 4, nil},
		{"cells are dropped from the right", Row(cells(Auto(a), Auto(b), Auto(c)), RowGap(4)), 5, []string{"a"}},
		{"one cell ignores the gap", Row(cells(Auto(a)), RowGap(9)), 3, []string{"aaa"}},
		{"children get their width", Row(cells(Fixed(4, wrapping("aa bb")), Flex(1, wrapping("cc dd")))), 7, []string{"aa  cc ", "bb  dd "}},
		{"no children", Row(nil), 5, nil},
		{"no room", Row(cells(Auto(a))), 0, nil},

		{"top", Row(cells(Auto(raw("x\ny\nz")), Auto(c))), unbounded, []string{"xc", "y", "z"}},
		{"middle", Row(cells(Auto(raw("x\ny\nz")), Auto(c)), RowAlign(dra.AlignMiddle)), unbounded, []string{"x", "yc", "z"}},
		{"bottom", Row(cells(Auto(raw("x\ny\nz")), Auto(c)), RowAlign(dra.AlignBottom)), unbounded, []string{"x", "y", "zc"}},
		{"padded cells are rectangles", Row(cells(Auto(raw("x\ny\nz")), Fixed(2, c)), RowAlign(dra.AlignMiddle)), unbounded, []string{"x  ", "yc ", "z  "}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.want == nil {
				checkEmpty(t, tt.row, tt.width)
				return
			}
			checkText(t, tt.row, tt.width, tt.want...)
		})
	}
}

func TestRowCopiesCells(t *testing.T) {
	in := cells(Auto(raw("a")), Auto(raw("b")))
	row := Row(in)
	in[0] = Auto(raw("X"))
	checkText(t, row, unbounded, "ab")
}

func TestRowErrors(t *testing.T) {
	checkChildErrors(t, "row", 1, func(c beautify.Composable) beautify.Composable {
		return Row(cells(Auto(raw("a")), Auto(c), Flex(1, raw("b"))))
	})
	checkChildErrors(t, "row", 0, func(c beautify.Composable) beautify.Composable {
		return Row(cells(Flex(1, c)))
	})
	checkChildErrors(t, "row", 0, func(c beautify.Composable) beautify.Composable {
		return Row(cells(Fixed(5, c)))
	})
}

func TestRowContract(t *testing.T) {
	build := func(c beautify.Composable) beautify.Composable {
		return Row(cells(Fixed(5, c), Auto(c), Flex(2, c), Flex(1, c)), RowGap(1), RowAlign(dra.AlignMiddle))
	}
	checkChildContract(t, build)
	rect := Row(cells(Fixed(5, children["wide"]), Flex(1, children["multi-line"])), RowGap(1))
	checkRectangle(t, renderWith(t, longFactory, rect, 20))
	styledCells := cells(Fixed(5, Text("a\nb", TextStyle(red))), Flex(1, Text("flex", TextStyle(blue))), Auto(Text("auto")))
	checkGolden(t, "row_default", Styled(Row(styledCells), onBG), 20)
	checkGolden(t, "row_gap", Row(styledCells, RowGap(2)), 20)
	checkGolden(t, "row_bottom", Row(styledCells, RowAlign(dra.AlignBottom)), 20)
}
