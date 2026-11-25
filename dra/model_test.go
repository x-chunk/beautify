package dra

import (
	"math/rand/v2"
	"strings"
	"sync"
	"testing"
)

// frame builds an unstyled leaf.
func frame(prefix, suffix string, mode Mode) Model {
	return newLeaf(leaf{prefix: prefix, suffix: suffix, mode: mode})
}

func TestZeroModel(t *testing.T) {
	var m Model
	if m.Width() != 0 || m.Height() != 0 || !m.IsEmpty() || m.HasSlot() || m.IsFilled() || m.Mode() != ModeReal || m.String() != "" {
		t.Errorf("zero model is not empty: %+v", m)
	}
	if got := m.Fill(frame("x", "", ModeReal)); !got.IsEmpty() {
		t.Errorf("filling the zero model gives %q", got.String())
	}
}

func TestModelMetrics(t *testing.T) {
	tests := []struct {
		name           string
		prefix, suffix string
		slot           string // "-" means unfilled
		text           string
		width, height  int
	}{
		{"empty", "", "", "-", "", 0, 0},
		{"filled with empty", "", "", "", "", 0, 0},
		{"one line", "abc", "", "-", "abc", 3, 1},
		{"newline only", "\n", "", "-", "\n", 0, 2},
		{"trailing newline", "ab\n", "", "-", "ab\n", 2, 2},
		{"unfilled frame", "[", "]", "-", "[]", 2, 1},
		{"filled frame", "[", "]", "abc", "[abc]", 5, 1},
		{"frame filled with empty", "[", "]", "", "[]", 2, 1},
		{"only slot", "", "", "ab\nc", "ab\nc", 2, 2},
		{"multi-line value flows", "a\n", "\nb", "x\ny", "a\nx\ny\nb", 1, 4},
		{"value continues the prefix", "key: ", "!", "one\nthree", "key: one\nthree!", 8, 2},
		{"suffix only", "", "end", "a\n", "a\nend", 3, 2},
		{"widest line in the middle", "a", "b", "x\nlongest\ny", "ax\nlongest\nyb", 7, 3},
	}
	for _, tt := range tests {
		for _, mode := range []Mode{ModeReal, ModeLong} {
			t.Run(tt.name, func(t *testing.T) {
				m := frame(tt.prefix, tt.suffix, mode)
				if tt.slot != "-" {
					m = m.Fill(frame(tt.slot, "", mode))
				}
				if got := m.String(); got != tt.text {
					t.Errorf("String() = %q, want %q", got, tt.text)
				}
				if m.Width() != tt.width || m.Height() != tt.height {
					t.Errorf("size = %dx%d, want %dx%d", m.Width(), m.Height(), tt.width, tt.height)
				}
				if m.IsEmpty() != (tt.height == 0) {
					t.Errorf("IsEmpty() = %v", m.IsEmpty())
				}
				if m.Mode() != mode {
					t.Errorf("Mode() = %v, want %v", m.Mode(), mode)
				}
			})
		}
	}
}

func TestModelWideMetrics(t *testing.T) {
	const text = "漢字\né"
	if m := frame(text, "", ModeReal); m.Width() != 2 {
		t.Errorf("real width = %d, want 2", m.Width())
	}
	if m := frame(text, "", ModeLong); m.Width() != 4 {
		t.Errorf("long width = %d, want 4", m.Width())
	}
	// Every piece keeps its own mode, and a combining mark at the start of
	// the slot content is measured on its own.
	mixed := frame("漢", "", ModeLong).Fill(frame("漢́", "", ModeReal))
	if mixed.Width() != 4 {
		t.Errorf("mixed width = %d, want 4", mixed.Width())
	}
	if mixed.Mode() != ModeLong {
		t.Errorf("mixed mode = %v, want the leaf's mode", mixed.Mode())
	}
}

func TestNestedFill(t *testing.T) {
	inner := frame("<", ">", ModeReal).Fill(frame("x\ny", "", ModeReal))
	middle := frame("(", ")", ModeReal).Fill(inner)
	outer := frame("[", "]", ModeReal).Fill(middle)
	if got, want := outer.String(), "[(<x\ny>)]"; got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
	if outer.Width() != 4 || outer.Height() != 2 {
		t.Errorf("size = %dx%d, want 4x2", outer.Width(), outer.Height())
	}
}

func TestFill(t *testing.T) {
	blank := frame("[", "]", ModeReal)
	if !blank.HasSlot() || blank.IsFilled() {
		t.Errorf("blank: HasSlot=%v IsFilled=%v", blank.HasSlot(), blank.IsFilled())
	}
	a := blank.Fill(frame("a", "", ModeReal))
	if blank.IsFilled() || blank.String() != "[]" {
		t.Error("Fill changed its receiver")
	}
	if !a.IsFilled() || !a.HasSlot() {
		t.Errorf("filled: HasSlot=%v IsFilled=%v", a.HasSlot(), a.IsFilled())
	}
	if got := a.Fill(frame("bb", "", ModeReal)); got.String() != "[bb]" || got.Width() != 4 {
		t.Errorf("refill = %q, width %d", got.String(), got.Width())
	}
	empty := blank.Fill(Model{})
	if !empty.IsFilled() || empty.String() != "[]" || empty.Width() != 2 {
		t.Errorf("filled with empty = %q, IsFilled %v", empty.String(), empty.IsFilled())
	}
}

func TestFillConcurrently(t *testing.T) {
	blank := frame("[", "]", ModeReal)
	var wg sync.WaitGroup
	for i := range 32 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			content := strings.Repeat("x", i)
			if got := blank.Fill(frame(content, "", ModeReal)).String(); got != "["+content+"]" {
				t.Errorf("got %q", got)
			}
		}()
	}
	wg.Wait()
}

// alphabet holds runes that exercise measuring, wrapping and cutting.
var alphabet = []string{
	"a", "b", "c", " ", " ", "\n", "я", "漢", "́", "‍", "️",
	"\U0001F468", "\U0001F3FD", "\U0001F1E6", "\U0001F1E7", "ᄒ", "ᅡ",
}

func randomText(rng *rand.Rand, maxRunes int) string {
	var b strings.Builder
	for range rng.IntN(maxRunes + 1) {
		b.WriteString(alphabet[rng.IntN(len(alphabet))])
	}
	return b.String()
}

// sizeOf measures text independently of the cached metrics.
func sizeOf(text string, mode Mode) (width, height int) {
	if text == "" {
		return 0, 0
	}
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		width = max(width, measure(line, mode).max)
	}
	return width, len(lines)
}

func TestCachedMetricsMatchText(t *testing.T) {
	rng := rand.New(rand.NewPCG(5, 6))
	for range 2000 {
		for _, mode := range []Mode{ModeReal, ModeLong} {
			m := frame(randomText(rng, 12), "", mode)
			if w, h := sizeOf(m.String(), mode); m.Width() != w || m.Height() != h {
				t.Fatalf("%q: cached %dx%d, measured %dx%d", m.String(), m.Width(), m.Height(), w, h)
			}
		}
		// In Long Mode a cluster can form across the slot boundary in the
		// joined string, so nesting is compared in Real Mode only.
		m := frame(randomText(rng, 6), randomText(rng, 6), ModeReal)
		for range rng.IntN(4) {
			m = frame(randomText(rng, 6), randomText(rng, 6), ModeReal).Fill(m)
		}
		if w, h := sizeOf(m.String(), ModeReal); m.Width() != w || m.Height() != h {
			t.Fatalf("%q: cached %dx%d, measured %dx%d", m.String(), m.Width(), m.Height(), w, h)
		}
	}
}
