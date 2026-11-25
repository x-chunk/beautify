package blocks

import (
	"go.xchunk.org/beautify"
	"go.xchunk.org/beautify/dra"
)

// PaddingOption configures Padding. Negative values count as 0.
type PaddingOption func(*padding)

// PaddingAll sets the padding on all four sides.
func PaddingAll(n int) PaddingOption {
	return PaddingSides(n, n, n, n)
}

// PaddingX sets the padding on the left and the right.
func PaddingX(n int) PaddingOption {
	return func(p *padding) { p.left, p.right = max(n, 0), max(n, 0) }
}

// PaddingY sets the padding above and below.
func PaddingY(n int) PaddingOption {
	return func(p *padding) { p.top, p.bottom = max(n, 0), max(n, 0) }
}

// PaddingSides sets the padding of each side.
func PaddingSides(top, right, bottom, left int) PaddingOption {
	return func(p *padding) {
		p.top, p.right = max(top, 0), max(right, 0)
		p.bottom, p.left = max(bottom, 0), max(left, 0)
	}
}

type padding struct {
	child                    beautify.Composable
	top, right, bottom, left int
}

// Padding surrounds child with blank space. The result is a rectangle. Around
// an empty child it renders only the space above and below.
func Padding(child beautify.Composable, opts ...PaddingOption) beautify.Composable {
	p := padding{child: child}
	applyOptions(&p, opts)
	return p
}

// Render implements beautify.Composable.
func (p padding) Render(ctx beautify.Context) (dra.Model, error) {
	if noRoom(ctx) {
		return dra.Model{}, nil
	}
	m, err := renderChild(shrink(ctx, p.left+p.right), "padding", 0, p.child)
	if err != nil {
		return dra.Model{}, err
	}
	m = pad(m, p.left, p.right)
	width := m.Width()
	if m.IsEmpty() {
		width = p.left + p.right
	}
	m = dra.Stack(dra.AlignLeft, dra.Blank(width, p.top), m, dra.Blank(width, p.bottom))
	return clip(ctx, m), nil
}

// pad makes m a rectangle with blank columns on both sides.
func pad(m dra.Model, left, right int) dra.Model {
	m = dra.Pad(m, m.Width(), dra.AlignLeft)
	l, r := dra.Blank(left, 1), dra.Blank(right, 1)
	return dra.SuffixLines(dra.PrefixLines(m, l, l), r, r)
}
