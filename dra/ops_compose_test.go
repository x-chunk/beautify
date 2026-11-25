package dra

import (
	"math/rand/v2"
	"testing"

	"go.xchunk.org/beautify/internal/golden"
)

// text is an unstyled Real Mode model.
func text(s string) Model {
	return NewFactory().Text(s)
}

// long is an unstyled Long Mode model.
func long(s string) Model {
	return NewFactory(WithMode(ModeLong)).Text(s)
}

// blankFrame is a model with an unfilled slot.
func blankFrame() Model {
	return NewFactory().MustNew(WithFrame("<", ">"))
}

// checkModel compares the visible text and size of the result of an
// operation.
func checkModel(t *testing.T, got Model, want string, width, height int) {
	t.Helper()
	if got.String() != want || got.Width() != width || got.Height() != height {
		t.Errorf("got %q %dx%d, want %q %dx%d",
			got.String(), got.Width(), got.Height(), want, width, height)
	}
	if got.HasSlot() {
		t.Error("the result has a slot")
	}
	if c := composed(flatten(&got)); c.size != got.size {
		t.Errorf("cached size %+v, actual %+v", got.size, c.size)
	}
}

func TestLines(t *testing.T) {
	tests := []struct {
		name string
		m    Model
		want []string
	}{
		{"empty", Model{}, nil},
		{"one", text("ab"), []string{"ab"}},
		{"empty lines", text("\na\n"), []string{"", "a", ""}},
		{"only newlines", text("\n"), []string{"", ""}},
		{"unfilled slot", blankFrame(), []string{"<>"}},
		{"filled slot", blankFrame().Fill(text("a\nb")), []string{"<a", "b>"}},
		{"composed", Blank(1, 2), []string{" ", " "}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Lines(tt.m)
			if len(got) != len(tt.want) || (got == nil) != (tt.want == nil) {
				t.Fatalf("got %d lines, want %d", len(got), len(tt.want))
			}
			for i, l := range got {
				checkModel(t, l, tt.want[i], len(tt.want[i]), 1)
			}
		})
	}
}

func TestBlank(t *testing.T) {
	checkModel(t, Blank(3, 2), "   \n   ", 3, 2)
	checkModel(t, Blank(0, 1), "", 0, 1)
	checkModel(t, Blank(0, 3), "\n\n", 0, 3)
	checkModel(t, Blank(3, 0), "", 0, 0)
	checkModel(t, Blank(-1, 2), "\n", 0, 2)
	checkModel(t, Blank(2, -2), "", 0, 0)
	if got := Blank(100, 1); got.Width() != 100 || len(got.String()) != 100 {
		t.Errorf("Blank(100, 1) is %d wide", got.Width())
	}
}

func TestConcat(t *testing.T) {
	tests := []struct {
		name          string
		ms            []Model
		want          string
		width, height int
	}{
		{"none", nil, "", 0, 0},
		{"one", []Model{text("a")}, "a", 1, 1},
		{"two", []Model{text("a"), text("b")}, "ab", 2, 1},
		{"flow", []Model{text("a\nbb"), text("c\nd")}, "a\nbbc\nd", 3, 3},
		{"only empty", []Model{{}, text("")}, "", 0, 0},
		{"empty is skipped", []Model{text("a"), {}, text("b")}, "ab", 2, 1},
		{"empty line is not empty", []Model{Blank(0, 1)}, "", 0, 1},
		{"newlines", []Model{text("\n"), text("\n")}, "\n\n", 0, 3},
		{"unfilled slot", []Model{blankFrame(), blankFrame()}, "<><>", 4, 1},
		{"mixed modes", []Model{text("漢"), long("漢")}, "漢漢", 3, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkModel(t, Concat(tt.ms...), tt.want, tt.width, tt.height)
		})
	}
}

func TestConcatGolden(t *testing.T) {
	f := NewFactory()
	m := Concat(
		f.Styled("red\nover", Style{FG: Red}),
		f.Styled(" two ", Style{BG: Blue}),
		f.Text("plain\nend"),
	)
	boxed := f.MustNew(WithFrame("[", "]"), WithStyle(Style{Attrs: Bold})).Fill(m)
	golden.Assert(t, "concat_styled", NewEngine(ProfileANSI16).Append(nil, boxed))
}

func TestLinesRoundTrip(t *testing.T) {
	rng := rand.New(rand.NewPCG(13, 14))
	for range 500 {
		m := randomModel(rng, 2)
		var b builder
		for _, l := range Lines(m) {
			b.addLine(firstLine(l))
			b.endLine()
		}
		if got := b.model(); got.String() != m.String() || got.size != m.size {
			t.Fatalf("got %q, want %q", got.String(), m.String())
		}
	}
}
