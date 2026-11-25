package blocks

import (
	"go.xchunk.org/beautify"
	"go.xchunk.org/beautify/dra"
)

type model struct {
	m dra.Model
}

// Model returns a block that shows a prepared model, cut to the bound.
func Model(m dra.Model) beautify.Composable {
	return model{m}
}

// Render implements beautify.Composable.
func (m model) Render(ctx beautify.Context) (dra.Model, error) {
	if noRoom(ctx) {
		return dra.Model{}, nil
	}
	return clip(ctx, m.m), nil
}
