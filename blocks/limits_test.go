package blocks

import (
	"math"
	"testing"

	"go.xchunk.org/beautify"
)

// TestHugeSizes verifies that sizes far beyond the bound neither overflow nor
// allocate what can never be shown.
func TestHugeSizes(t *testing.T) {
	const huge = math.MaxInt
	for name, c := range map[string]beautify.Composable{
		"padding":  Padding(raw("a"), PaddingX(huge)),
		"border":   Border(raw("a"), BorderPadding(huge, 0)),
		"sections": Sections([]Section{{"t", raw("a")}, {"u", raw("b")}}, SectionsPadding(huge, 0)),
		"table":    Table(TableRow(raw("a"), raw("b")), TablePadding(huge)),
		"row":      Row([]Cell{Fixed(huge, raw("a")), Fixed(huge, raw("b"))}, RowGap(huge)),
		"space":    Space(huge, 1),
		"sized":    Sized(raw("a"), Width(huge)),
		"rule":     Rule(RuleWidth(huge)),
	} {
		m, err := c.Render(context(realFactory, 10))
		if err != nil || m.Width() != 10 {
			t.Errorf("%s: width %d, %v, want 10", name, m.Width(), err)
		}
	}
}
