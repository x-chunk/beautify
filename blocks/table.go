package blocks

import (
	"fmt"
	"slices"

	"go.xchunk.org/beautify"
	"go.xchunk.org/beautify/dra"
)

// TableBorderSet holds the characters a Table is drawn with. Each should be
// one cell wide: wider ones are cut.
//
// A horizontal line whose fields are all empty is not drawn, and neither is a
// vertical line (the left edge, the column lines or the right edge) whose
// fields are all empty. Other empty fields are drawn as spaces.
type TableBorderSet struct {
	TopLeft, Top, TopJoin, TopRight             string // line above the table
	HeaderLeft, Header, HeaderJoin, HeaderRight string // line under the header
	Left, Column, Right                         string // edges of a row
	BottomLeft, Bottom, BottomJoin, BottomRight string // line below the table
}

// Provided table border sets.
var (
	// TableRounded has light lines and round corners. It is the default.
	TableRounded = TableBorderSet{
		TopLeft: "╭", Top: "─", TopJoin: "┬", TopRight: "╮",
		HeaderLeft: "├", Header: "─", HeaderJoin: "┼", HeaderRight: "┤",
		Left: "│", Column: "│", Right: "│",
		BottomLeft: "╰", Bottom: "─", BottomJoin: "┴", BottomRight: "╯",
	}
	// TableSquare has light lines and square corners.
	TableSquare = TableBorderSet{
		TopLeft: "┌", Top: "─", TopJoin: "┬", TopRight: "┐",
		HeaderLeft: "├", Header: "─", HeaderJoin: "┼", HeaderRight: "┤",
		Left: "│", Column: "│", Right: "│",
		BottomLeft: "└", Bottom: "─", BottomJoin: "┴", BottomRight: "┘",
	}
	// TableASCII uses only ASCII characters.
	TableASCII = TableBorderSet{
		TopLeft: "+", Top: "-", TopJoin: "+", TopRight: "+",
		HeaderLeft: "+", Header: "-", HeaderJoin: "+", HeaderRight: "+",
		Left: "|", Column: "|", Right: "|",
		BottomLeft: "+", Bottom: "-", BottomJoin: "+", BottomRight: "+",
	}
	// TableMinimal has only a line under the header. With the default
	// padding its columns are separated by two spaces.
	TableMinimal = TableBorderSet{Header: "─"}
	// TableNone draws no lines.
	TableNone = TableBorderSet{}
)

// hasLeft, hasColumns and hasRight report which vertical lines are drawn.
func (s *TableBorderSet) hasLeft() bool {
	return s.TopLeft+s.HeaderLeft+s.Left+s.BottomLeft != ""
}

func (s *TableBorderSet) hasColumns() bool {
	return s.TopJoin+s.HeaderJoin+s.Column+s.BottomJoin != ""
}

func (s *TableBorderSet) hasRight() bool {
	return s.TopRight+s.HeaderRight+s.Right+s.BottomRight != ""
}

// ColumnSpec describes a column of a Table. Zero and negative sizes are
// unset.
type ColumnSpec struct {
	Width    int // exact content width, or 0 to fit the content
	MinWidth int // lower limit of a column that fits its content
	MaxWidth int // upper limit of a column that fits its content; wins over MinWidth
	Align    dra.Align
}

// TableOption configures Table.
type TableOption func(*table)

// TableHeaders sets the header row. A nil cell is empty.
func TableHeaders(cells ...beautify.Composable) TableOption {
	cells = slices.Clone(cells)
	return func(t *table) { t.headers = cells }
}

// TableRow adds a row. A nil cell is empty.
func TableRow(cells ...beautify.Composable) TableOption {
	cells = slices.Clone(cells)
	return func(t *table) { t.rows = append(t.rows, cells) }
}

// TableRows adds several rows. A nil cell is empty.
func TableRows(rows [][]beautify.Composable) TableOption {
	rows = slices.Clone(rows)
	for i := range rows {
		rows[i] = slices.Clone(rows[i])
	}
	return func(t *table) { t.rows = append(t.rows, rows...) }
}

// TableColumns describes the columns from left to right. Columns without a
// spec fit their content, and extra specs are ignored.
func TableColumns(specs ...ColumnSpec) TableOption {
	specs = slices.Clone(specs)
	return func(t *table) { t.specs = specs }
}

// TableBorder sets the characters of the lines. The default is TableRounded.
func TableBorder(set TableBorderSet) TableOption {
	return func(t *table) { t.set = set }
}

// TableBorderStyle sets the style of the lines.
func TableBorderStyle(st dra.Style) TableOption {
	return func(t *table) { t.borderStyle = st }
}

// TableHeaderStyle sets a style under the header cells, their padding
// included.
func TableHeaderStyle(st dra.Style) TableOption {
	return func(t *table) { t.headerStyle = st }
}

// TablePadding sets the blank cells on both sides of the content of every
// cell. The default is 1.
func TablePadding(n int) TableOption {
	return func(t *table) { t.padding = max(n, 0) }
}

type table struct {
	headers     []beautify.Composable
	rows        [][]beautify.Composable
	specs       []ColumnSpec
	set         TableBorderSet
	borderStyle dra.Style
	headerStyle dra.Style
	padding     int
}

// Table shows rows of cells in columns. Rows can have different lengths;
// missing cells are empty.
//
// Columns are as wide as their content. When that exceeds the bound, the
// widest columns shrink first and their cells are rendered again at the
// smaller width. Columns do not shrink below their minimum width or below 1,
// and columns of fixed width do not shrink; a table that is still too wide
// is cut.
func Table(opts ...TableOption) beautify.Composable {
	t := table{set: TableRounded, padding: 1}
	applyOptions(&t, opts)
	return t
}

// Render implements beautify.Composable.
func (t table) Render(ctx beautify.Context) (dra.Model, error) {
	if noRoom(ctx) {
		return dra.Model{}, nil
	}
	columns := len(t.headers)
	for _, r := range t.rows {
		columns = max(columns, len(r))
	}
	if columns == 0 {
		return dra.Model{}, nil
	}
	widths, err := t.widths(ctx, columns)
	if err != nil {
		return dra.Model{}, err
	}

	f := ctx.Factory()
	parts := []dra.Model{t.line(f, widths, t.set.TopLeft, t.set.Top, t.set.TopJoin, t.set.TopRight)}
	if len(t.headers) > 0 {
		header, err := t.row(ctx, -1, widths)
		if err != nil {
			return dra.Model{}, err
		}
		parts = append(parts, header)
		if len(t.rows) > 0 {
			parts = append(parts, t.line(f, widths, t.set.HeaderLeft, t.set.Header, t.set.HeaderJoin, t.set.HeaderRight))
		}
	}
	for i := range t.rows {
		row, err := t.row(ctx, i, widths)
		if err != nil {
			return dra.Model{}, err
		}
		parts = append(parts, row)
	}
	parts = append(parts, t.line(f, widths, t.set.BottomLeft, t.set.Bottom, t.set.BottomJoin, t.set.BottomRight))
	return clip(ctx, dra.Stack(dra.AlignLeft, parts...)), nil
}

// spec returns the spec of a column without negative sizes.
func (t *table) spec(column int) ColumnSpec {
	if column >= len(t.specs) {
		return ColumnSpec{}
	}
	s := t.specs[column]
	s.Width, s.MinWidth, s.MaxWidth = max(s.Width, 0), max(s.MinWidth, 0), max(s.MaxWidth, 0)
	return s
}

// cells returns the cells of a row; row -1 is the header.
func (t *table) cells(row int) []beautify.Composable {
	if row < 0 {
		return t.headers
	}
	return t.rows[row]
}

// cell renders one cell. Missing and nil cells are empty.
func (t *table) cell(ctx beautify.Context, row, column int) (dra.Model, error) {
	cells := t.cells(row)
	if column >= len(cells) || cells[column] == nil {
		return dra.Model{}, nil
	}
	m, err := cells[column].Render(ctx)
	if err != nil {
		if row < 0 {
			return dra.Model{}, fmt.Errorf("blocks: table header %d: %w", column, err)
		}
		return dra.Model{}, fmt.Errorf("blocks: table row %d cell %d: %w", row, column, err)
	}
	return clip(ctx, m), nil
}

// widths chooses the content width of every column.
func (t *table) widths(ctx beautify.Context, columns int) ([]int, error) {
	widths := make([]int, columns)
	for c := range widths {
		w, err := t.naturalWidth(ctx, c)
		if err != nil {
			return nil, err
		}
		widths[c] = w
	}
	bound, bounded := ctx.Width()
	if !bounded {
		return widths, nil
	}

	avail := bound - 2*t.padding*columns
	if t.set.hasLeft() {
		avail--
	}
	if t.set.hasRight() {
		avail--
	}
	if t.set.hasColumns() {
		avail -= columns - 1
	}
	total := 0
	for c := range widths {
		// No column needs more than the bound. Limiting it here keeps the
		// number of steps below small and does not change their result.
		widths[c] = min(widths[c], bound)
		total += widths[c]
	}
	for ; total > avail; total-- {
		widest := -1
		for c, w := range widths {
			s := t.spec(c)
			if s.Width == 0 && w > max(s.MinWidth, 1) && (widest < 0 || w > widths[widest]) {
				widest = c
			}
		}
		if widest < 0 {
			break
		}
		widths[widest]--
	}
	return widths, nil
}

// naturalWidth returns the width of a column that has all the room it needs.
func (t *table) naturalWidth(ctx beautify.Context, column int) (int, error) {
	s := t.spec(column)
	if s.Width > 0 {
		return s.Width, nil
	}
	width := s.MinWidth
	for row := -1; row < len(t.rows); row++ {
		m, err := t.cell(ctx.Unbounded(), row, column)
		if err != nil {
			return 0, err
		}
		width = max(width, m.Width())
	}
	if s.MaxWidth > 0 {
		width = min(width, s.MaxWidth)
	}
	return width, nil
}

// row renders a row at the final widths; row -1 is the header.
func (t *table) row(ctx beautify.Context, row int, widths []int) (dra.Model, error) {
	cells := make([]dra.Model, len(widths))
	height := 1
	for c, w := range widths {
		m, err := t.cell(ctx.WithWidth(w), row, c)
		if err != nil {
			return dra.Model{}, err
		}
		cells[c] = m
		height = max(height, m.Height())
	}
	for c, w := range widths {
		m := dra.Pad(dra.PadHeight(cells[c], height, dra.AlignTop), w, t.spec(c).Align)
		m = pad(m, t.padding, t.padding)
		if row < 0 {
			m = dra.Restyle(m, t.headerStyle)
		}
		cells[c] = m
	}
	edge := func(s string) dra.Model {
		return vertical(glyph(ctx.Factory(), s, t.borderStyle), height)
	}
	return t.join(cells, edge(t.set.Left), edge(t.set.Column), edge(t.set.Right)), nil
}

// line draws a horizontal line, or nothing when all its fields are empty.
func (t *table) line(f *dra.Factory, widths []int, left, fill, join, right string) dra.Model {
	if left+fill+join+right == "" {
		return dra.Model{}
	}
	cell := func(s string) dra.Model { return glyph(f, s, t.borderStyle) }
	runs := make([]dra.Model, len(widths))
	for c, w := range widths {
		runs[c] = dra.Repeat(cell(fill), w+2*t.padding)
	}
	return t.join(runs, cell(left), cell(join), cell(right))
}

// join places columns side by side with the vertical lines of the set.
func (t *table) join(columns []dra.Model, left, between, right dra.Model) dra.Model {
	parts := make([]dra.Model, 0, 2*len(columns)+1)
	if t.set.hasLeft() {
		parts = append(parts, left)
	}
	for c, m := range columns {
		if c > 0 && t.set.hasColumns() {
			parts = append(parts, between)
		}
		parts = append(parts, m)
	}
	if t.set.hasRight() {
		parts = append(parts, right)
	}
	return dra.Join(dra.AlignTop, parts...)
}
