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
