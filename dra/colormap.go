package dra

// NoStyle is the colormap entry for "no own style": the rune has the base
// style of its model.
const NoStyle int8 = -1

// Run describes N consecutive runes that share palette index Index.
type Run struct {
	N     int
	Index int8
}

// Runs expands runs into a colormap. Runs with N <= 0 are skipped. The
// result is never nil.
func Runs(runs ...Run) []int8 {
	cmap := []int8{}
	for _, r := range runs {
		for range r.N {
			cmap = append(cmap, r.Index)
		}
	}
	return cmap
}
