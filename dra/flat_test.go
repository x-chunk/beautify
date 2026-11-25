package dra

import (
	"fmt"
	"math/rand/v2"
	"testing"
)

// composedText builds a composed model from unstyled text.
func composedText(s string, mode Mode) Model {
	m := frame(s, "", mode)
	return composed(flatten(&m))
}

func TestBuilderMergesEqualSpans(t *testing.T) {
	red := Style{FG: Red}
	var b builder
	for _, s := range []span{
		{text: "a", width: 1}, {text: "b", width: 1}, {text: "c", width: 1},
		{text: "", style: red},
		{text: "d", style: red, width: 1},
		{text: "e", style: red, mode: ModeLong, width: 1},
	} {
		b.add(s)
	}
	b.endLine()
	got := fmt.Sprintf("%+v", b.model().lines[0].spans)
	want := fmt.Sprintf("%+v", []span{
		{text: "abc", width: 3},
		{text: "d", style: red, width: 1},
		{text: "e", style: red, mode: ModeLong, width: 1},
	})
	if got != want {
		t.Errorf("got  %s\nwant %s", got, want)
	}
}

func TestBuilderKeepsClustersApart(t *testing.T) {
	// Joined, the two halves would be measured as one cluster.
	a, z := span{text: "\U0001F468‍", mode: ModeLong, width: 2}, span{text: "\U0001F469", mode: ModeLong, width: 2}
	var b builder
	b.add(a)
	b.add(z)
	b.endLine()
	if m := b.model(); len(m.lines[0].spans) != 2 || m.Width() != 4 {
		t.Errorf("spans = %+v, width %d", m.lines[0].spans, m.Width())
	}
}

func TestComposedModel(t *testing.T) {
	tests := []struct {
		name          string
		in            string
		width, height int
	}{
		{"empty", "", 0, 0},
		{"one line", "abc", 3, 1},
		{"newlines", "\n\n", 0, 3},
		{"trailing newline", "ab\n", 2, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := composedText(tt.in, ModeReal)
			if m.String() != tt.in || m.Width() != tt.width || m.Height() != tt.height {
				t.Errorf("got %q, %dx%d", m.String(), m.Width(), m.Height())
			}
			if m.HasSlot() || m.IsFilled() {
				t.Error("a composed model has a slot")
			}
			filled := m.Fill(frame("x", "", ModeReal))
			if filled.String() != tt.in || filled.IsFilled() {
				t.Errorf("Fill() = %q", filled.String())
			}
		})
	}
}

func TestComposedMode(t *testing.T) {
	if composedText("a", ModeReal).Mode() != ModeReal || composedText("a", ModeLong).Mode() != ModeLong {
		t.Error("wrong mode")
	}
	mixed := frame("a", "", ModeReal).Fill(frame("漢", "", ModeLong))
	if got := composed(flatten(&mixed)); got.Mode() != ModeLong || got.Width() != 3 {
		t.Errorf("mode %v, width %d", got.Mode(), got.Width())
	}
}

func TestFlattenKeepsOutput(t *testing.T) {
	rng := rand.New(rand.NewPCG(11, 12))
	for range 1500 {
		m := randomModel(rng, 3)
		c := composed(flatten(&m))
		if c.Width() != m.Width() || c.Height() != m.Height() || c.size != m.size {
			t.Fatalf("%q: size %+v, want %+v", m.String(), c.size, m.size)
		}
		outer := NewFactory().MustNew(WithFrame("<", ">"), WithStyle(randomStyle(rng)))
		for _, p := range allProfiles {
			e := NewEngine(p)
			if got, want := e.String(c), e.String(m); got != want {
				t.Fatalf("%v: got %q, want %q", p, got, want)
			}
			// Styles stay relative, so inheritance still works.
			if got, want := e.String(outer.Fill(c)), e.String(outer.Fill(m)); got != want {
				t.Fatalf("%v in a slot: got %q, want %q", p, got, want)
			}
			checkOutput(t, c, p)
		}
	}
}

func TestFlattenEveryStyle(t *testing.T) {
	// A slot chain with more distinct styles than one palette can hold.
	m := frame("", "", ModeReal)
	for i := range 200 {
		m = NewFactory().MustNew(WithFrame("x", ""), WithPalette(Style{FG: RGB(uint8(i), 0, 0)}),
			WithColormap([]int8{0})).Fill(m)
	}
	c := composed(flatten(&m))
	if got := len(c.lines[0].spans); got != 200 {
		t.Errorf("%d spans, want 200", got)
	}
	e := NewEngine(ProfileTrueColor)
	if e.String(c) != e.String(m) {
		t.Error("output differs")
	}
}
