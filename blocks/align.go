package blocks

import (
	"go.xchunk.org/beautify"
	"go.xchunk.org/beautify/dra"
)

type align struct {
	child beautify.Composable
	h     dra.Align
}

// Align places every line of child within the bound. In an unbounded context
// lines are aligned within the width of the child.
func Align(child beautify.Composable, h dra.Align) beautify.Composable {
	return align{child, h}
}

// Render implements beautify.Composable.
func (a align) Render(ctx beautify.Context) (dra.Model, error) {
	if noRoom(ctx) {
		return dra.Model{}, nil
	}
	m, err := renderChild(ctx, "align", 0, a.child)
	if err != nil {
		return dra.Model{}, err
	}
	width, bounded := ctx.Width()
	if !bounded {
		width = m.Width()
	}
	return dra.Pad(m, width, a.h), nil
}
