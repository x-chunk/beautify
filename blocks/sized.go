package blocks

import (
	"go.xchunk.org/beautify"
	"go.xchunk.org/beautify/dra"
)

// SizedOption configures Sized. Negative sizes count as 0.
type SizedOption func(*sized)

// MinWidth pads the child to at least n cells.
func MinWidth(n int) SizedOption {
	return func(s *sized) { s.minWidth = n }
}

// MaxWidth limits the child to n cells.
func MaxWidth(n int) SizedOption {
	return func(s *sized) { s.maxWidth = limit{max(n, 0), true} }
}

// Width sets the minimum and the maximum width to n cells.
func Width(n int) SizedOption {
	return func(s *sized) {
		MinWidth(n)(s)
		MaxWidth(n)(s)
	}
}

// MinHeight pads the child to at least n lines.
func MinHeight(n int) SizedOption {
	return func(s *sized) { s.minHeight = n }
}

// MaxHeight limits the child to n lines.
func MaxHeight(n int) SizedOption {
	return func(s *sized) { s.maxHeight = limit{max(n, 0), true} }
}

// Height sets the minimum and the maximum height to n lines.
func Height(n int) SizedOption {
	return func(s *sized) {
		MinHeight(n)(s)
		MaxHeight(n)(s)
	}
}

// limit is an optional maximum.
type limit struct {
	n   int
	set bool
}

type sized struct {
	child               beautify.Composable
	minWidth, minHeight int
	maxWidth, maxHeight limit
}

// Sized constrains the size of child. A maximum wins over a minimum, and the
// bound of the context wins over both. An empty child without a minimum
// height stays empty.
func Sized(child beautify.Composable, opts ...SizedOption) beautify.Composable {
	s := sized{child: child}
	applyOptions(&s, opts)
	return s
}

// Render implements beautify.Composable.
func (s sized) Render(ctx beautify.Context) (dra.Model, error) {
	if noRoom(ctx) {
		return dra.Model{}, nil
	}
	if s.maxWidth.set {
		if w, bounded := ctx.Width(); !bounded || s.maxWidth.n < w {
			ctx = ctx.WithWidth(s.maxWidth.n)
		}
	}
	m, err := renderChild(ctx, "sized", 0, s.child)
	if err != nil {
		return dra.Model{}, err
	}
	m = dra.PadHeight(m, s.minHeight, dra.AlignTop)
	if s.maxHeight.set {
		m = dra.TruncateHeight(m, s.maxHeight.n)
	}
	width := s.minWidth
	if w, bounded := ctx.Width(); bounded {
		width = min(width, w)
	}
	return clip(ctx, dra.Pad(m, width, dra.AlignLeft)), nil
}
