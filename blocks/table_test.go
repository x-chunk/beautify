package blocks

import (
	"errors"
	"strings"
	"testing"

	"go.xchunk.org/beautify"
	"go.xchunk.org/beautify/dra"
)

func texts(ss ...string) []beautify.Composable {
	var out []beautify.Composable
	for _, s := range ss {
		out = append(out, wrapping(s))
	}
	return out
}

func TestTable(t *testing.T) {
	headers := TableHeaders(texts("Name", "Qty")...)
	apple, kiwi := TableRow(texts("Apple", "1")...), TableRow(texts("Kiwi", "12")...)

	checkText(t, Table(headers, apple, kiwi), unbounded,
		"╭───────┬─────╮",
		"│ Name  │ Qty │",
		"├───────┼─────┤",
		"│ Apple │ 1   │",
		"│ Kiwi  │ 12  │",
		"╰───────┴─────╯")
	checkText(t, Table(headers, apple), 40,
		"╭───────┬─────╮",
		"│ Name  │ Qty │",
		"├───────┼─────┤",
		"│ Apple │ 1   │",
		"╰───────┴─────╯")
	checkText(t, Table(headers), unbounded,
		"╭──────┬─────╮",
		"│ Name │ Qty │",
		"╰──────┴─────╯")
	checkText(t, Table(apple, kiwi), unbounded,
		"╭───────┬────╮",
		"│ Apple │ 1  │",
		"│ Kiwi  │ 12 │",
		"╰───────┴────╯")
	checkText(t, Table(headers, apple, TableBorder(TableASCII)), unbounded,
		"+-------+-----+",
		"| Name  | Qty |",
		"+-------+-----+",
		"| Apple | 1   |",
		"+-------+-----+")
	checkText(t, Table(headers, apple, TableBorder(TableMinimal)), unbounded,
		" Name   Qty ",
		"────────────",
		" Apple  1   ")
	checkText(t, Table(headers, apple, TableBorder(TableNone)), unbounded,
		" Name   Qty ",
		" Apple  1   ")
	checkText(t, Table(headers, apple, TablePadding(0)), unbounded,
		"╭─────┬───╮",
		"│Name │Qty│",
		"├─────┼───┤",
		"│Apple│1  │",
		"╰─────┴───╯")
	checkText(t, Table(apple, TablePadding(-4), nil), unbounded,
		"╭─────┬─╮",
		"│Apple│1│",
		"╰─────┴─╯")
	checkText(t, Table(TableRows([][]beautify.Composable{texts("a"), texts("b", "c", "d")}), TableRow(texts("e", "f")...)), unbounded,
		"╭───┬───┬───╮",
		"│ a │   │   │",
		"│ b │ c │ d │",
		"│ e │ f │   │",
		"╰───┴───┴───╯")
	checkText(t, Table(headers, TableRow(nil, raw("x\ny"))), unbounded,
		"╭──────┬─────╮",
		"│ Name │ Qty │",
		"├──────┼─────┤",
		"│      │ x   │",
		"│      │ y   │",
		"╰──────┴─────╯")
	checkText(t, Table(TableRow(nil)), unbounded, "╭──╮", "│  │", "╰──╯")
	checkText(t, Table(TableRow(texts("a")...), TableRow()), unbounded, "╭───╮", "│ a │", "│   │", "╰───╯")
	checkEmpty(t, Table(), unbounded)
	checkEmpty(t, Table(TableHeaders(), TableRow(), TableRows(nil)), unbounded)
	checkEmpty(t, Table(headers, apple), 0)
}

func TestTableColumns(t *testing.T) {
	r := TableRow(texts("a", "bbb")...)
	s := TableRow(texts("ccc", "d")...)
	checkText(t, Table(r, s, TableColumns(ColumnSpec{Align: dra.AlignRight}, ColumnSpec{Width: 5, Align: dra.AlignCenter}, ColumnSpec{Width: 9})), unbounded,
		"╭─────┬───────╮",
		"│   a │  bbb  │",
		"│ ccc │   d   │",
		"╰─────┴───────╯")
	checkText(t, Table(r, TableColumns(ColumnSpec{MinWidth: 4}, ColumnSpec{MaxWidth: 2})), unbounded,
		"╭──────┬────╮",
		"│ a    │ bb │",
		"│      │ b  │",
		"╰──────┴────╯")
	checkText(t, Table(r, TableColumns(ColumnSpec{MinWidth: 4, MaxWidth: 2}, ColumnSpec{Width: -5})), unbounded,
		"╭────┬─────╮",
		"│ a  │ bbb │",
		"╰────┴─────╯")
	checkText(t, Table(r, TableColumns(ColumnSpec{Width: 2}), TableColumns(ColumnSpec{Width: 1})), unbounded,
		"╭───┬─────╮",
		"│ a │ bbb │",
		"╰───┴─────╯")
}

func TestTableShrink(t *testing.T) {
	r := TableRow(texts("aaaa bbbb", "cccc")...)
	checkText(t, Table(r), 19,
		"╭──────────┬──────╮",
		"│ aaaa     │ cccc │",
		"│ bbbb     │      │",
		"╰──────────┴──────╯")
	// Both columns are equally wide from here on; the leftmost shrinks first.
	checkText(t, Table(r), 14,
		"╭─────┬──────╮",
		"│ aaa │ cccc │",
		"│ a   │      │",
		"│ bbb │      │",
		"│ b   │      │",
		"╰─────┴──────╯")
	checkText(t, Table(r, TableColumns(ColumnSpec{MinWidth: 4})), 14,
		"╭──────┬─────╮",
		"│ aaaa │ ccc │",
		"│ bbbb │ c   │",
		"╰──────┴─────╯")
	checkText(t, Table(r, TableColumns(ColumnSpec{Width: 4})), 14,
		"╭──────┬─────╮",
		"│ aaaa │ ccc │",
		"│ bbbb │ c   │",
		"╰──────┴─────╯")
	// Nothing can shrink any more: the table is cut.
	checkText(t, Table(TableRow(texts("a", "b")...)), 7,
		"╭───┬──",
		"│ a │ b",
		"╰───┴──")
	checkText(t, Table(TableRow(texts("abc")...), TableColumns(ColumnSpec{Width: 1 << 50})), 7,
		"╭──────",
		"│ abc  ",
		"╰──────")
}

func TestTableCellsFillTheirColumn(t *testing.T) {
	checkText(t, Table(TableRow(wrapping("wide cell")), TableRow(Rule(RuleWidth(2)))), unbounded,
		"╭───────────╮",
		"│ wide cell │",
		"│ ───────── │",
		"╰───────────╯")
}

func TestTableCustomBorder(t *testing.T) {
	set := TableBorderSet{Top: "=-", Left: "", Column: "!", HeaderLeft: ">"}
	checkText(t, Table(TableHeaders(texts("h", "i")...), TableRow(texts("a", "b")...), TableBorder(set), TablePadding(0)), unbounded,
		" = =", " h!i", ">   ", " a!b")
}

func TestTableErrors(t *testing.T) {
	tests := map[string]beautify.Composable{
		"table header 1":     Table(TableHeaders(raw("a"), failing())),
		"table row 1 cell 0": Table(TableRow(raw("a")), TableRow(failing())),
	}
	for want, c := range tests {
		for _, width := range []int{unbounded, 20} {
			m, err := c.Render(context(realFactory, width))
			if !errors.Is(err, errChild) || !strings.Contains(err.Error(), want) || !m.IsEmpty() {
				t.Errorf("error = %v, want %q; model %q", err, want, m.String())
			}
		}
	}
	// A column of fixed width renders its cells only once.
	fixed := Table(TableRow(failing()), TableColumns(ColumnSpec{Width: 3}))
	if _, err := fixed.Render(context(realFactory, 20)); !errors.Is(err, errChild) {
		t.Errorf("error = %v", err)
	}
}

func TestTableCopiesInput(t *testing.T) {
	cells := texts("a", "b")
	rows := [][]beautify.Composable{texts("c", "d")}
	specs := []ColumnSpec{{}, {}}
	table := Table(TableHeaders(cells...), TableRows(rows), TableRow(cells...), TableColumns(specs...))
	want := render(t, table, unbounded).String()
	cells[0], rows[0][0], specs[0] = raw("X"), raw("X"), ColumnSpec{Width: 9}
	checkText(t, table, unbounded, want)
}

func TestTableOptionIsReusable(t *testing.T) {
	r := TableRow(texts("a")...)
	checkText(t, Table(r, r), unbounded, "╭───╮", "│ a │", "│ a │", "╰───╯")
	checkText(t, Table(r), unbounded, "╭───╮", "│ a │", "╰───╯")
}

func TestTableContract(t *testing.T) {
	build := func(c beautify.Composable) beautify.Composable {
		return Table(TableHeaders(raw("漢字"), c), TableRow(c, c, nil), TableRow(raw("x")))
	}
	checkChildContract(t, build)
	checkChildContract(t, func(c beautify.Composable) beautify.Composable {
		return Table(TableRow(c, c), TableBorder(TableMinimal), TableColumns(ColumnSpec{MinWidth: 6}, ColumnSpec{Width: 5}))
	})
	for _, child := range children {
		for _, width := range []int{18, 30, unbounded} {
			checkRectangle(t, renderWith(t, longFactory, build(child), width))
		}
	}

	headers := TableHeaders(Text("Name"), Text("Qty", TextStyle(red)))
	rows := TableRows([][]beautify.Composable{
		{Text("Apple", TextStyle(blue)), Text("1")}, {Text("Kiwi\nfruit"), Text("12")},
	})
	for name, set := range map[string]TableBorderSet{
		"rounded": TableRounded, "square": TableSquare, "ascii": TableASCII, "minimal": TableMinimal, "none": TableNone,
	} {
		checkGolden(t, "table_set_"+name, Table(headers, rows, TableBorder(set), TableBorderStyle(red)), unbounded)
	}
	checkGolden(t, "table_default", Styled(Table(headers, rows), onBG), unbounded)
	checkGolden(t, "table_header_style", Table(headers, rows, TableHeaderStyle(dra.Style{BG: dra.Blue, Attrs: dra.Bold})), unbounded)
	checkGolden(t, "table_padding", Table(headers, rows, TablePadding(2)), unbounded)
	checkGolden(t, "table_columns", Table(headers, rows, TableColumns(ColumnSpec{Width: 8}, ColumnSpec{Align: dra.AlignRight})), unbounded)
	checkGolden(t, "table_narrow", Table(headers, rows), 12)
}
