package dra

import (
	"math"
	"testing"
)

func TestGradientAt(t *testing.T) {
	black, white, red := RGB(0, 0, 0), RGB(255, 255, 255), RGB(255, 0, 0)
	tests := []struct {
		name       string
		g          Gradient
		x, y, w, h int
		want       Color
	}{
		{"zero value", Gradient{}, 0, 0, 5, 5, Color{}},
		{"no stops", NewGradient(AxisVertical), 1, 1, 5, 5, Color{}},
		{"one stop", NewGradient(AxisVertical, red), 3, 4, 5, 5, red},
		{"vertical start", NewGradient(AxisVertical, black, white), 4, 0, 5, 5, black},
		{"vertical middle", NewGradient(AxisVertical, black, white), 4, 2, 5, 5, RGB(128, 128, 128)},
		{"vertical end", NewGradient(AxisVertical, black, white), 0, 4, 5, 5, white},
		{"horizontal quarter", NewGradient(AxisHorizontal, black, white), 1, 9, 5, 5, RGB(64, 64, 64)},
		{"horizontal end", NewGradient(AxisHorizontal, black, white), 4, 0, 5, 5, white},
		{"diagonal start", NewGradient(AxisDiagonal, black, white), 0, 0, 5, 3, black},
		{"diagonal middle", NewGradient(AxisDiagonal, black, white), 2, 1, 5, 3, RGB(128, 128, 128)},
		{"diagonal end", NewGradient(AxisDiagonal, black, white), 4, 2, 5, 3, white},
		{"unknown axis is vertical", NewGradient(Axis(9), black, white), 0, 4, 5, 5, white},
		{"three stops hit the middle one", NewGradient(AxisHorizontal, black, red, white), 2, 0, 5, 1, red},
		{"three stops, first half", NewGradient(AxisHorizontal, black, red, white), 1, 0, 5, 1, RGB(128, 0, 0)},
		{"three stops, second half", NewGradient(AxisHorizontal, black, red, white), 3, 0, 5, 1, RGB(255, 128, 128)},
		{"falling channel", NewGradient(AxisHorizontal, white, black), 1, 0, 5, 1, RGB(191, 191, 191)},
		{"1 by 1", NewGradient(AxisDiagonal, black, white), 0, 0, 1, 1, black},
		{"one row, vertical", NewGradient(AxisVertical, black, white), 7, 0, 9, 1, black},
		{"one row, diagonal", NewGradient(AxisDiagonal, black, white), 8, 0, 9, 1, white},
		{"zero size", NewGradient(AxisVertical, black, white), 0, 3, 0, 0, black},
		{"negative size", NewGradient(AxisDiagonal, black, white), 2, 3, -4, -4, black},
		{"before the start", NewGradient(AxisVertical, black, white), 0, -3, 5, 5, black},
		{"after the end", NewGradient(AxisVertical, black, white), 0, 40, 5, 5, white},
		{"huge coordinates", NewGradient(AxisDiagonal, black, white), math.MaxInt, math.MaxInt, math.MaxInt, math.MaxInt, white},
		{"palette stops use the xterm values", NewGradient(AxisVertical, Red, ANSI256(231)), 0, 0, 1, 2, RGB(128, 0, 0)},
		{"unset and default stops are dropped", NewGradient(AxisVertical, Color{}, red, TerminalDefault), 0, 0, 1, 2, red},
		{"only dropped stops", NewGradient(AxisVertical, Color{}, TerminalDefault), 0, 0, 1, 2, Color{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.g.At(tt.x, tt.y, tt.w, tt.h); got != tt.want {
				t.Errorf("At() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestGradientIsZero(t *testing.T) {
	if !(Gradient{}).IsZero() || !NewGradient(AxisDiagonal, Color{}).IsZero() {
		t.Error("a gradient without stops is not zero")
	}
	if NewGradient(AxisVertical, Red).IsZero() {
		t.Error("a gradient with a stop is zero")
	}
}

func TestGradientCopiesStops(t *testing.T) {
	stops := []Color{RGB(1, 2, 3)}
	g := NewGradient(AxisVertical, stops...)
	stops[0] = RGB(9, 9, 9)
	if got := g.At(0, 0, 1, 1); got != RGB(1, 2, 3) {
		t.Errorf("At() = %+v", got)
	}
}

// TestGradientIsExact compares every color with exact rational arithmetic.
// Floating point is not good enough: its results differ between platforms
// that fuse multiplication and addition and those that do not.
func TestGradientIsExact(t *testing.T) {
	values := []int64{255, 0, 43, 128, 7}
	for stops := 2; stops <= len(values); stops++ {
		var colors []Color
		for _, v := range values[:stops] {
			colors = append(colors, RGB(uint8(v), 0, 0))
		}
		g := NewGradient(AxisHorizontal, colors...)
		for width := 2; width <= 120; width++ {
			den := int64(width - 1)
			for x := range width {
				pos := int64(x) * int64(stops-1)
				i, rem := pos/den, pos%den
				sum := values[i] * (den - rem)
				if rem > 0 {
					sum += values[i+1] * rem
				}
				want := RGB(uint8((2*sum+den)/(2*den)), 0, 0)
				if got := g.At(x, 0, width, 1); got != want {
					t.Errorf("%d stops, cell %d of %d: %+v, want %+v", stops, x, width, got, want)
				}
			}
		}
	}
}

func TestGradientExtremeCoordinates(t *testing.T) {
	black, white := RGB(0, 0, 0), RGB(255, 255, 255)
	for _, axis := range []Axis{AxisVertical, AxisHorizontal, AxisDiagonal} {
		g := NewGradient(axis, black, white)
		if got := g.At(math.MinInt, math.MinInt, math.MaxInt, math.MaxInt); got != black {
			t.Errorf("axis %d, start: %+v", axis, got)
		}
		if got := g.At(math.MaxInt, math.MaxInt, math.MaxInt, math.MaxInt); got != white {
			t.Errorf("axis %d, end: %+v", axis, got)
		}
		if got := g.At(math.MaxInt/2, math.MaxInt/2, math.MaxInt, math.MaxInt); got != RGB(127, 127, 127) && got != RGB(128, 128, 128) {
			t.Errorf("axis %d, middle: %+v", axis, got)
		}
		if got := g.At(5, 5, math.MinInt, math.MinInt); got != black {
			t.Errorf("axis %d, negative size: %+v", axis, got)
		}
	}
}
