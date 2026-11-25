package blocks

import (
	"go.xchunk.org/beautify"
	"go.xchunk.org/beautify/dra"
)

type styled struct {
	child beautify.Composable
	style dra.Style
}

// Styled applies st under child: the styles of the child win, and st fills
// the colors they leave unset and adds its attributes.
func Styled(child beautify.Composable, st dra.Style) beautify.Composable {
	return styled{child, st}
}

// Render implements beautify.Composable.
func (s styled) Render(ctx beautify.Context) (dra.Model, error) {
	if noRoom(ctx) {
		return dra.Model{}, nil
	}
	m, err := renderChild(ctx, "styled", 0, s.child)
	if err != nil {
		return dra.Model{}, err
	}
	return dra.Restyle(m, s.style), nil
}
