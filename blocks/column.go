package blocks

import (
	"slices"

	"go.xchunk.org/beautify"
	"go.xchunk.org/beautify/dra"
)

// ColumnOption configures Column.
type ColumnOption func(*column)

// ColumnGap sets the number of empty lines between neighbors. The default is
// 0.
func ColumnGap(n int) ColumnOption {
	return func(c *column) { c.gap = max(n, 0) }
}

// ColumnAlign sets how children are aligned with the widest one. The default
// is the left.
func ColumnAlign(a dra.Align) ColumnOption {
	return func(c *column) { c.align = a }
}

type column struct {
	children []beautify.Composable
	gap      int
	align    dra.Align
}

// Column places children under each other. Empty children are skipped
// together with their gap.
func Column(children []beautify.Composable, opts ...ColumnOption) beautify.Composable {
	c := column{children: slices.Clone(children)}
	applyOptions(&c, opts)
	return c
}

// Render implements beautify.Composable.
func (c column) Render(ctx beautify.Context) (dra.Model, error) {
	if noRoom(ctx) {
		return dra.Model{}, nil
	}
	parts := make([]dra.Model, 0, 2*len(c.children))
	for i, child := range c.children {
		m, err := renderChild(ctx, "column", i, child)
		if err != nil {
			return dra.Model{}, err
		}
		if m.IsEmpty() {
			continue
		}
		if len(parts) > 0 {
			parts = append(parts, dra.Blank(0, c.gap))
		}
		parts = append(parts, m)
	}
	return dra.Stack(c.align, parts...), nil
}
