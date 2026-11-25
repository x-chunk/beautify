package blocks

import (
	"go.xchunk.org/beautify"
	"go.xchunk.org/beautify/dra"
)

type badge struct {
	text  string
	style dra.Style
}

// Badge returns a short label with one space on each side and st applied to
// all of it. Line breaks in text become spaces.
func Badge(text string, st dra.Style) beautify.Composable {
	return badge{oneLine(text), st}
}

// Render implements beautify.Composable.
func (b badge) Render(ctx beautify.Context) (dra.Model, error) {
	if noRoom(ctx) {
		return dra.Model{}, nil
	}
	f := ctx.Factory()
	text := f.Styled(b.text, b.style)
	if w, bounded := ctx.Width(); bounded {
		text = dra.Truncate(text, w-2, f.Styled(ellipsis, b.style))
	}
	space := f.Styled(" ", b.style)
	return clip(ctx, dra.Concat(space, text, space)), nil
}
