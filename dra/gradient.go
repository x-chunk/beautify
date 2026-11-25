package dra

import (
	"math"

	"go.xchunk.org/beautify/internal/ansi"
)

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
	pos := g.position(x, y, width, height) * float64(len(g.stops)-1)
	i := min(int(pos), len(g.stops)-1)
	from, to := g.stops[i], g.stops[min(i+1, len(g.stops)-1)]
	frac := pos - float64(i)
	var c [3]uint8
	for ch := range c {
		a, b := float64(from[ch]), float64(to[ch])
		c[ch] = uint8(math.Round(a + (b-a)*frac))
	}
	return RGB(c[0], c[1], c[2])
}

// position maps a cell to the range 0-1 along the axis. It works on floats
// so that extreme coordinates cannot overflow.
func (g Gradient) position(x, y, width, height int) float64 {
	dx, dy := math.Max(float64(width)-1, 0), math.Max(float64(height)-1, 0)
	num, den := float64(y), dy
	switch g.axis {
	case AxisHorizontal:
		num, den = float64(x), dx
	case AxisDiagonal:
		num, den = float64(x)+float64(y), dx+dy
	}
	if den == 0 {
		return 0
	}
	return math.Min(math.Max(num/den, 0), 1)
}
