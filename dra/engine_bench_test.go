package dra

import (
	"strings"
	"testing"
)

// fiveRuns is a line with five style runs.
func fiveRuns() Model {
	return NewFactory().MustNew(
		WithText("12:34:56 INFO server started on :8080"),
		WithPalette(Style{FG: BrightBlack}, Style{FG: Green, Attrs: Bold}, Style{FG: RGB(80, 160, 240)}),
		WithColormap(Runs(Run{8, 0}, Run{1, NoStyle}, Run{4, 1}, Run{1, NoStyle}, Run{23, 2})),
	)
}

// threeLevels is a blank filled three levels deep.
func threeLevels() Model {
	f := NewFactory()
	inner := f.MustNew(WithFrame("<", ">"), WithStyle(Style{FG: Red})).Fill(f.Text("value"))
	middle := f.MustNew(WithFrame("(", ")"), WithStyle(Style{BG: Blue})).Fill(inner)
	return f.MustNew(WithFrame("[", "]"), WithStyle(Style{Attrs: Bold})).Fill(middle)
}

func TestAppendDoesNotAllocate(t *testing.T) {
	models := map[string]Model{
		"plain":  NewFactory().Text("hello, world"),
		"runs":   fiveRuns(),
		"nested": threeLevels(),
	}
	buf := make([]byte, 0, 4096)
	for name, m := range models {
		for _, p := range allProfiles {
			e := NewEngine(p)
			if n := testing.AllocsPerRun(100, func() { buf = e.Append(buf[:0], m) }); n != 0 {
				t.Errorf("%s at %v: %v allocations, want 0", name, p, n)
			}
		}
	}
}

func BenchmarkAppend(b *testing.B) {
	models := []struct {
		name string
		m    Model
	}{
		{"Plain", NewFactory().Text(strings.Repeat("plain text ", 7))},
		{"FiveRuns", fiveRuns()},
		{"Nested", threeLevels()},
	}
	for _, bm := range models {
		b.Run(bm.name, func(b *testing.B) {
			e := NewEngine(ProfileTrueColor)
			var buf []byte
			b.ReportAllocs()
			for range b.N {
				buf = e.Append(buf[:0], bm.m)
			}
		})
	}
}
