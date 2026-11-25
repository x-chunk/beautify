package blocks

import (
	"testing"

	"go.xchunk.org/beautify"
	"go.xchunk.org/beautify/dra"
)

func TestStyled(t *testing.T) {
	checkText(t, Styled(raw("a\nb"), red), unbounded, "a", "b")
	checkText(t, Styled(raw("abc"), red), 2, "ab")
	checkEmpty(t, Styled(raw("abc"), red), 0)
	checkEmpty(t, Styled(raw(""), red), 5)

	m := render(t, Styled(Text("x", TextStyle(dra.Style{FG: dra.Green})), red.WithBG(dra.Blue)), unbounded)
	if got, want := dra.NewEngine(dra.ProfileANSI16).String(m), "\x1b[0;32;44mx\x1b[0m"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	checkGolden(t, "styled_background", Styled(Text("a\nbb"), onBG), unbounded)
	build := func(c beautify.Composable) beautify.Composable { return Styled(c, red) }
	checkChildContract(t, build)
	checkChildErrors(t, "styled", 0, build)
}
