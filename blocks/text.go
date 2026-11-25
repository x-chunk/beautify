package blocks

import (
	"go.xchunk.org/beautify"
	"go.xchunk.org/beautify/dra"
)

// Overflow selects what Text does with lines that are wider than the bound.
// Unknown values behave like OverflowWrap.
type Overflow uint8

const (
	// OverflowWrap wraps lines at word boundaries. It is the default.
	OverflowWrap Overflow = iota
	// OverflowTruncate cuts lines and marks the cut with an ellipsis.
	OverflowTruncate
	// OverflowKeep never re-flows the text. Lines are still cut at the
	// bound, without a mark.
	OverflowKeep
)

// TextOption configures Text.
type TextOption func(*text)

// TextStyle sets the style of the text.
func TextStyle(st dra.Style) TextOption {
	return func(t *text) { t.style = st }
}

// TextOverflow sets how lines wider than the bound are handled.
func TextOverflow(o Overflow) TextOption {
	return func(t *text) { t.overflow = o }
}

// TextEllipsis sets the mark that OverflowTruncate puts at a cut. The
// default is "…".
func TextEllipsis(s string) TextOption {
	return func(t *text) { t.ellipsis = s }
}

type text struct {
	content  string
	style    dra.Style
	overflow Overflow
	ellipsis string
}

// Text returns a block of text. Line breaks in s are kept. In a bounded
// context lines that are too wide are wrapped unless TextOverflow says
// otherwise.
func Text(s string, opts ...TextOption) beautify.Composable {
	t := text{content: s, ellipsis: ellipsis}
	applyOptions(&t, opts)
	return t
}

// Render implements beautify.Composable.
func (t text) Render(ctx beautify.Context) (dra.Model, error) {
	if noRoom(ctx) {
		return dra.Model{}, nil
	}
	m := ctx.Factory().Styled(t.content, t.style)
	if w, bounded := ctx.Width(); bounded {
		switch t.overflow {
		case OverflowTruncate:
			m = dra.Truncate(m, w, ctx.Factory().Styled(t.ellipsis, t.style))
		case OverflowKeep:
		default:
			m = dra.Wrap(m, w, dra.WrapWords)
		}
	}
	// Wrapping leaves a wide cluster alone on its line at width 1.
	return clip(ctx, m), nil
}
