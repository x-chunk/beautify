package dra

import "go.xchunk.org/beautify/internal/ansi"

// Axis is the direction a Gradient runs in. Unknown values behave like
// AxisVertical.
type Axis uint8

// Axes of a gradient.
const (
	AxisVertical   Axis = iota // top to bottom
	AxisHorizontal             // left to right
	AxisDiagonal               // top left to bottom right
)

// Gradient is a color that changes with the position inside a rectangle. It
// is an immutable value. The zero value has no stops and yields the unset
// Color everywhere.
type Gradient struct {
	axis  Axis
	stops [][3]uint8
}

// NewGradient returns a gradient that runs along axis through stops, which
// are spread evenly. Palette colors are converted to RGB with the xterm
// defaults. Unset stops and TerminalDefault are dropped.
func NewGradient(axis Axis, stops ...Color) Gradient {
	g := Gradient{axis: axis}
	for _, c := range stops {
		if c.IsSet() && c != TerminalDefault {
			r, gr, b := ansi.ToRGB(c.c)
			g.stops = append(g.stops, [3]uint8{r, gr, b})
		}
	}
	return g
}

// IsZero reports whether the gradient has no stops.
func (g Gradient) IsZero() bool {
	return len(g.stops) == 0
}

// At returns the RGB color of the cell in column x and line y of a rectangle
// of width by height cells. Cells outside the rectangle take the color of
// the nearest end. Without stops the color is unset.
func (g Gradient) At(x, y, width, height int) Color {
	if len(g.stops) == 0 {
		return Color{}
	}
	num, den := g.position(x, y, width, height)
	num *= int64(len(g.stops) - 1)
	i := num / den
	rem := num % den
	from, to := g.stops[i], g.stops[min(i+1, int64(len(g.stops)-1))]
	var c [3]uint8
	for ch := range c {
		sum := int64(from[ch])*(den-rem) + int64(to[ch])*rem
		c[ch] = uint8((2*sum + den) / (2 * den))
	}
	return RGB(c[0], c[1], c[2])
}

// position maps a cell to the fraction num/den in the range 0-1 along the
// axis. den is positive and small enough for At to multiply by it.
//
// Integers keep the result exact and equal on all platforms; floating point
// is not, because some compilers fuse multiplication and addition.
func (g Gradient) position(x, y, width, height int) (num, den int64) {
	const (
		maxSum = 1 << 61 // two such values still fit into an int64
		maxDen = 1 << 30
	)
	px, py := int64(x), int64(y)
	dx, dy := int64(max(width, 1)-1), int64(max(height, 1)-1)
	for max(px, py, dx, dy) > maxSum || min(px, py) < -maxSum {
		// Scaling everything keeps the proportions of absurd sizes.
		px, py, dx, dy = px/2, py/2, dx/2, dy/2
	}
	switch g.axis {
	case AxisHorizontal:
		num, den = px, dx
	case AxisDiagonal:
		num, den = px+py, dx+dy
	default:
		num, den = py, dy
	}
	if den == 0 {
		return 0, 1
	}
	num = min(max(num, 0), den)
	for den > maxDen {
		num, den = num/2, den/2
	}
	return num, den
}
