package dra

import (
	"slices"
	"testing"
)

func TestRuns(t *testing.T) {
	tests := []struct {
		name string
		runs []Run
		want []int8
	}{
		{"none", nil, []int8{}},
		{"one", []Run{{3, 1}}, []int8{1, 1, 1}},
		{"several", []Run{{1, 0}, {2, NoStyle}, {1, 5}}, []int8{0, -1, -1, 5}},
		{"non-positive counts are skipped", []Run{{0, 1}, {-2, 1}, {1, 2}}, []int8{2}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Runs(tt.runs...)
			if !slices.Equal(got, tt.want) {
				t.Errorf("Runs() = %v, want %v", got, tt.want)
			}
			if got == nil {
				t.Error("Runs() = nil, which WithColormap reads as no colormap")
			}
		})
	}
}
