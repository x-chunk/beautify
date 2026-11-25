package blocks

import (
	"go.xchunk.org/beautify"
	"go.xchunk.org/beautify/dra"
)

type space struct {
	width, height int
}

// Space returns a blank area of the given size. Negative values count as 0.
// In a bounded context the width is reduced to the bound.
func Space(width, height int) beautify.Composable {
	return space{width, height}
}

// Render implements beautify.Composable.
func (s space) Render(ctx beautify.Context) (dra.Model, error) {
	if noRoom(ctx) {
		return dra.Model{}, nil
	}
	width := s.width
	if w, bounded := ctx.Width(); bounded {
		width = min(width, w)
	}
	return dra.Blank(width, s.height), nil
}
