package blocks

import (
	"fmt"
	"strings"

	"go.xchunk.org/beautify"
	"go.xchunk.org/beautify/dra"
)

// ellipsis marks text that was cut.
const ellipsis = "…"

// noRoom reports whether ctx is bounded to a width of 0.
func noRoom(ctx beautify.Context) bool {
	w, bounded := ctx.Width()
	return bounded && w == 0
}

// shrink returns ctx with its bound reduced by n cells.
func shrink(ctx beautify.Context, n int) beautify.Context {
	if w, bounded := ctx.Width(); bounded {
		return ctx.WithWidth(w - n)
	}
	return ctx
}

// clip cuts m to the bound of ctx.
func clip(ctx beautify.Context, m dra.Model) dra.Model {
	if w, bounded := ctx.Width(); bounded && m.Width() > w {
		return dra.Truncate(m, w, dra.Model{})
	}
	return m
}

// renderChild renders a child and enforces the bound of ctx. Errors name the
// block and the index of the child.
func renderChild(ctx beautify.Context, block string, index int, c beautify.Composable) (dra.Model, error) {
	if c == nil {
		return dra.Model{}, childError(block, index, beautify.ErrNilComposable)
	}
	m, err := c.Render(ctx)
	if err != nil {
		return dra.Model{}, childError(block, index, err)
	}
	return clip(ctx, m), nil
}

func childError(block string, index int, err error) error {
	return fmt.Errorf("blocks: %s child %d: %w", block, index, err)
}

// applyOptions runs opts on cfg. Nil options are skipped.
func applyOptions[C any, O ~func(*C)](cfg *C, opts []O) {
	for _, opt := range opts {
		if opt != nil {
			opt(cfg)
		}
	}
}

// oneLine replaces line breaks with spaces, for labels that must not grow in
// height.
func oneLine(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, "\r\n", " "), "\n", " ")
}
