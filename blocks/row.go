package blocks

import (
	"slices"

	"go.xchunk.org/beautify"
	"go.xchunk.org/beautify/dra"
)

// Cell is a child of a Row together with the way its width is chosen.
type Cell struct {
	child beautify.Composable
	kind  cellKind
	size  int // width of a fixed cell, weight of a flex cell
}

type cellKind uint8

const (
	fixedCell cellKind = iota
	autoCell
	flexCell
)

// Fixed returns a cell of exactly n cells. Negative values count as 0.
func Fixed(n int, c beautify.Composable) Cell {
	return Cell{c, fixedCell, max(n, 0)}
}

// Auto returns a cell as wide as its content.
func Auto(c beautify.Composable) Cell {
	return Cell{c, autoCell, 0}
}

// maxWeight limits flex weights so that dividing the width cannot overflow.
const maxWeight = 1_000_000

// Flex returns a cell that gets a share of the width that the other cells
// leave. Weights are limited to the range from 1 to 1,000,000. In an
// unbounded context the cell is as wide as its content.
func Flex(weight int, c beautify.Composable) Cell {
	return Cell{c, flexCell, min(max(weight, 1), maxWeight)}
}

// RowOption configures Row.
type RowOption func(*row)

// RowGap sets the number of blank cells between neighbors. The default is 0.
func RowGap(n int) RowOption {
	return func(r *row) { r.gap = max(n, 0) }
}

// RowAlign sets the vertical alignment of cells with different heights. The
// default is the top.
func RowAlign(a dra.VAlign) RowOption {
	return func(r *row) { r.valign = a }
}

type row struct {
	cells  []Cell
	gap    int
	valign dra.VAlign
}

// Row places children side by side. In a bounded context, Fixed cells take
// their width first, then Auto cells, and Flex cells share the rest. A cell
// that gets no width is omitted together with its gap. When the gaps alone
// exceed the bound, cells are dropped from the right.
func Row(children []Cell, opts ...RowOption) beautify.Composable {
	r := row{cells: slices.Clone(children)}
	applyOptions(&r, opts)
	return r
}

// Render implements beautify.Composable.
func (r row) Render(ctx beautify.Context) (dra.Model, error) {
	if noRoom(ctx) {
		return dra.Model{}, nil
	}
	for i, c := range r.cells {
		if c.child == nil {
			return dra.Model{}, childError("row", i, beautify.ErrNilComposable)
		}
	}
	cells := r.cells
	bound, bounded := ctx.Width()
	if bounded {
		// Division avoids overflow for huge gaps.
		for len(cells) > 1 && r.gap > bound/(len(cells)-1) {
			cells = cells[:len(cells)-1]
		}
	}
	left := bound // width not taken yet, only used when bounded
	if bounded && len(cells) > 0 {
		left -= r.gap * (len(cells) - 1)
	}

	models := make([]dra.Model, len(cells))
	widths := make([]int, len(cells))
	// place renders cell i with at most width cells.
	place := func(i, width int, exact bool) error {
		childCtx := ctx
		if bounded || exact {
			childCtx = ctx.WithWidth(width)
		}
		if exact && width == 0 {
			return nil
		}
		m, err := renderChild(childCtx, "row", i, cells[i].child)
		if err != nil {
			return err
		}
		models[i], widths[i] = m, m.Width()
		if exact {
			widths[i] = width
		}
		left -= widths[i]
		return nil
	}

	weights := 0
	for i, c := range cells {
		if c.kind == fixedCell {
			width := c.size
			if bounded {
				width = min(width, left)
			}
			if err := place(i, width, true); err != nil {
				return dra.Model{}, err
			}
		}
	}
	for i, c := range cells {
		switch {
		case c.kind == autoCell, c.kind == flexCell && !bounded:
			if err := place(i, left, false); err != nil {
				return dra.Model{}, err
			}
		case c.kind == flexCell:
			weights += c.size
		}
	}
	if bounded && weights > 0 {
		for i, width := range shares(cells, left, weights) {
			if cells[i].kind == flexCell {
				if err := place(i, width, true); err != nil {
					return dra.Model{}, err
				}
			}
		}
	}
	return clip(ctx, r.join(cells, models, widths)), nil
}

// shares divides total among the flex cells by weight. The remainder of the
// division is handed out one cell at a time from left to right. Other cells
// get 0.
func shares(cells []Cell, total, weights int) []int {
	out := make([]int, len(cells))
	rest := total
	for i, c := range cells {
		if c.kind == flexCell {
			out[i] = total * c.size / weights
			rest -= out[i]
		}
	}
	for i, c := range cells {
		if c.kind == flexCell && rest > 0 {
			out[i]++
			rest--
		}
	}
	return out
}

// join places the rendered cells that have a width side by side. Fixed and
// flex cells become rectangles of their width so that columns line up.
func (r row) join(cells []Cell, models []dra.Model, widths []int) dra.Model {
	height := 1
	for _, m := range models {
		height = max(height, m.Height())
	}
	var parts []dra.Model
	for i, m := range models {
		if widths[i] == 0 {
			continue
		}
		if cells[i].kind != autoCell {
			m = dra.Pad(dra.PadHeight(m, height, r.valign), widths[i], dra.AlignLeft)
		}
		if len(parts) > 0 {
			parts = append(parts, dra.Blank(r.gap, 1))
		}
		parts = append(parts, m)
	}
	return dra.Join(r.valign, parts...)
}
