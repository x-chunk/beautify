package blocks

import (
	"testing"

	"go.xchunk.org/beautify"
	"go.xchunk.org/beautify/dra"
)

func TestAlign(t *testing.T) {
	c := raw("a\nbbb")
	checkText(t, Align(c, dra.AlignLeft), 5, "a    ", "bbb  ")
	checkText(t, Align(c, dra.AlignCenter), 6, "  a   ", " bbb  ")
	checkText(t, Align(c, dra.AlignRight), 5, "    a", "  bbb")
	checkText(t, Align(c, dra.AlignRight), unbounded, "  a", "bbb")
	checkText(t, Align(c, dra.AlignRight), 2, " a", "bb")
	checkEmpty(t, Align(c, dra.AlignRight), 0)
	checkEmpty(t, Align(raw(""), dra.AlignRight), 5)
	for _, a := range []dra.Align{dra.AlignLeft, dra.AlignCenter, dra.AlignRight} {
		build := func(c beautify.Composable) beautify.Composable { return Align(c, a) }
		checkChildContract(t, build)
		checkChildErrors(t, "align", 0, build)
		checkRectangle(t, renderWith(t, longFactory, build(children["wide"]), 40))
	}
	checkGolden(t, "align_center", Styled(Align(Text("a\nbbb", TextStyle(red)), dra.AlignCenter), onBG), 6)
	checkGolden(t, "align_right", Align(Text("a\nbbb", TextStyle(red)), dra.AlignRight), 6)
}
