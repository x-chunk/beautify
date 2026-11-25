package dra

import (
	"bytes"
	"errors"
	"io"
	"math/rand/v2"
	"regexp"
	"strings"
	"sync"
	"testing"
	"unicode/utf8"

	"go.xchunk.org/beautify/internal/golden"
)

var allProfiles = []Profile{
	ProfilePlain, ProfileNoColor, ProfileANSI16, ProfileANSI256, ProfileTrueColor,
}

func TestEngineReference(t *testing.T) {
	f := NewFactory()
	boldRed := f.Styled("hi", Style{FG: Red, Attrs: Bold})
	red := Style{FG: Red}
	tests := []struct {
		name    string
		m       Model
		profile Profile
		want    string
	}{
		{"unstyled", f.Text("hi"), ProfileTrueColor, "hi"},
		{"bold red", boldRed, ProfileTrueColor, "\x1b[0;1;31mhi\x1b[0m"},
		{"two lines", f.Styled("a\nb", red), ProfileTrueColor, "\x1b[0;31ma\x1b[0m\n\x1b[0;31mb\x1b[0m"},
		{"slot inherits base", f.MustNew(WithFrame("[", "]"), WithStyle(red)).Fill(f.Text("x")),
			ProfileTrueColor, "\x1b[0;31m[x]\x1b[0m"},
		{"styled slot", f.MustNew(WithFrame("[", "]")).Fill(f.Styled("x", Style{FG: Green})),
			ProfileTrueColor, "[\x1b[0;32mx\x1b[0m]"},
		{"trailing newline", f.Styled("a\n", red), ProfileTrueColor, "\x1b[0;31ma\x1b[0m\n"},
		{"no color", boldRed, ProfileNoColor, "\x1b[0;1mhi\x1b[0m"},
		{"plain", boldRed, ProfilePlain, "hi"},
		{"unknown profile", boldRed, Profile(99), "hi"},

		{"empty", Model{}, ProfileTrueColor, ""},
		{"only newlines with a base", f.Styled("\n\n", red), ProfileTrueColor, "\n\n"},
		{"styled spaces", f.Styled("  ", Style{BG: Blue}), ProfileTrueColor, "\x1b[0;44m  \x1b[0m"},
		{"unfilled slot", f.MustNew(WithFrame("[", "]"), WithStyle(red)), ProfileTrueColor, "\x1b[0;31m[]\x1b[0m"},
		{"slot base over parent", f.MustNew(WithFrame("[", "]"), WithStyle(Style{FG: Red, Attrs: Bold})).
			Fill(f.Styled("x", Style{FG: Green, Attrs: Underline})), ProfileTrueColor,
			"\x1b[0;1;31m[\x1b[0;1;4;32mx\x1b[0;1;31m]\x1b[0m"},
		{"terminal default in a red parent", f.MustNew(WithFrame("[", "]"), WithStyle(red)).
			Fill(f.Styled("x", Style{FG: TerminalDefault})), ProfileTrueColor,
			"\x1b[0;31m[\x1b[0mx\x1b[0;31m]\x1b[0m"},
		{"equal after downgrade", f.MustNew(WithText("ab"),
			WithPalette(Style{FG: RGB(255, 0, 0)}, Style{FG: RGB(254, 1, 1)}), WithColormap([]int8{0, 1})),
			ProfileANSI256, "\x1b[0;38;5;196mab\x1b[0m"},
		{"palette over base", f.MustNew(WithText("abc"), WithStyle(Style{FG: Red, Attrs: Bold}),
			WithPalette(Style{BG: Blue}), WithColormap([]int8{NoStyle, 0, NoStyle})), ProfileTrueColor,
			"\x1b[0;1;31ma\x1b[0;1;31;44mb\x1b[0;1;31mc\x1b[0m"},
		{"slot ignores the neighbors' style", f.MustNew(WithFrame("[", "]"),
			WithPalette(Style{FG: Red}), WithColormap([]int8{0, 0})).Fill(f.Text("x")), ProfileTrueColor,
			"\x1b[0;31m[\x1b[0mx\x1b[0;31m]\x1b[0m"},
		{"no switch for a run without visible text", f.MustNew(WithText("a\nb"),
			WithPalette(Style{FG: Red}), WithColormap([]int8{NoStyle, 0, NoStyle})), ProfileTrueColor, "a\nb"},
		{"sanitization off passes escapes", f.MustNew(WithText("\x1b[7mx"), WithSanitize(false)),
			ProfilePlain, "\x1b[7mx"},
		{"sanitized escape", f.Text("\x1b[7mx"), ProfileTrueColor, "�[7mx"},
		{"invalid utf-8", f.Styled("\xff", red), ProfileTrueColor, "\x1b[0;31m�\x1b[0m"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewEngine(tt.profile).String(tt.m); got != tt.want {
				t.Errorf("got  %q\nwant %q", got, tt.want)
			}
		})
	}
}

func TestEngineProfilesGolden(t *testing.T) {
	f := NewFactory()
	inner := f.MustNew(WithText("a\tb"), WithPalette(Style{FG: ANSI256(208)}, Style{BG: RGB(10, 20, 30)}),
		WithColormap(Runs(Run{1, 0}, Run{1, NoStyle}, Run{1, 1})), WithStyle(Style{Attrs: Italic}))
	m := f.MustNew(WithFrame("<\n", "\n>"), WithStyle(Style{FG: mustHex(t, "#ff8800"), BG: Blue, Attrs: Bold})).Fill(inner)
	for _, p := range allProfiles {
		t.Run(p.String(), func(t *testing.T) {
			golden.Assert(t, "engine_profile_"+p.String(), NewEngine(p).Append(nil, m))
		})
	}
}

// mustHex is Hex for constants in tests.
func mustHex(t *testing.T, s string) Color {
	t.Helper()
	c, err := Hex(s)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestEngineProfile(t *testing.T) {
	if got := NewEngine(ProfileANSI16).Profile(); got != ProfileANSI16 {
		t.Errorf("Profile() = %v", got)
	}
	var zero Engine
	if got := zero.String(NewFactory().Styled("x", Style{FG: Red})); got != "x" {
		t.Errorf("zero engine = %q", got)
	}
}

func TestEngineAppendKeepsPrefix(t *testing.T) {
	got := NewEngine(ProfileANSI16).Append([]byte("> "), NewFactory().Styled("x", Style{FG: Red}))
	if string(got) != "> \x1b[0;31mx\x1b[0m" {
		t.Errorf("got %q", got)
	}
}

func TestEngineDeepNesting(t *testing.T) {
	f := NewFactory(WithFrame("(", ")"))
	m := f.MustNew()
	for range 1000 {
		m = f.MustNew(WithStyle(Style{FG: Red})).Fill(m)
	}
	if m.Width() != 2002 {
		t.Errorf("width = %d, want 2002", m.Width())
	}
	want := "\x1b[0;31m" + strings.Repeat("(", 1001) + strings.Repeat(")", 1001) + "\x1b[0m"
	if got := NewEngine(ProfileANSI16).String(m); got != want {
		t.Errorf("wrong output of %d bytes", len(got))
	}
}

// countingWriter records calls and can fail or write short.
type countingWriter struct {
	bytes.Buffer
	calls int
	limit int // bytes accepted per call, negative for all
	err   error
}

func (w *countingWriter) Write(p []byte) (int, error) {
	w.calls++
	if w.limit >= 0 && len(p) > w.limit {
		p = p[:w.limit]
	}
	n, _ := w.Buffer.Write(p)
	return n, w.err
}

func TestEngineRender(t *testing.T) {
	f := NewFactory()
	e := NewEngine(ProfileANSI16)
	m := f.Styled("hi\nyou", Style{FG: Red})

	w := &countingWriter{limit: -1}
	n, err := e.Render(w, m)
	if err != nil || n != w.Len() || w.String() != e.String(m) || w.calls != 1 {
		t.Errorf("Render() = %d, %v, %d calls, output %q", n, err, w.calls, w.String())
	}

	w = &countingWriter{limit: -1}
	if n, err := e.Render(w, Model{}); n != 0 || err != nil || w.calls != 0 {
		t.Errorf("empty: Render() = %d, %v, %d calls", n, err, w.calls)
	}

	w = &countingWriter{limit: 3}
	if n, err := e.Render(w, m); n != 3 || !errors.Is(err, io.ErrShortWrite) {
		t.Errorf("short: Render() = %d, %v", n, err)
	}

	failure := errors.New("disk full")
	w = &countingWriter{limit: 2, err: failure}
	if n, err := e.Render(w, m); n != 2 || err != failure {
		t.Errorf("failing: Render() = %d, %v", n, err)
	}
}

var sgr = regexp.MustCompile("\x1b\\[[0-9;]*m")

// randomModel builds a model with random text, styles and up to depth
// levels of slot content.
func randomModel(rng *rand.Rand, depth int) Model {
	prefix, suffix := randomText(rng, 8), randomText(rng, 8)
	palette := []Style{randomStyle(rng), randomStyle(rng), randomStyle(rng)}
	cmap := make([]int8, utf8.RuneCountInString(prefix+suffix))
	for i := range cmap {
		cmap[i] = int8(rng.IntN(len(palette)+1) - 1)
	}
	m := NewFactory().MustNew(WithFrame(prefix, suffix), WithPalette(palette...),
		WithColormap(cmap), WithStyle(randomStyle(rng)), WithMode(Mode(rng.IntN(2))))
	if depth > 0 && rng.IntN(3) > 0 {
		m = m.Fill(randomModel(rng, depth-1))
	}
	return m
}

// checkOutput verifies the guarantees the Engine gives for any model without
// raw control characters.
func checkOutput(t *testing.T, m Model, p Profile) {
	t.Helper()
	out := NewEngine(p).String(m)
	if got := sgr.ReplaceAllString(out, ""); got != m.String() {
		t.Fatalf("%v: stripped %q, want %q", p, got, m.String())
	}
	if p == ProfilePlain && strings.Contains(out, "\x1b") {
		t.Fatalf("plain output %q contains an escape", out)
	}
	active := false
	for i := 0; i < len(out); i++ {
		switch out[i] {
		case 0x1b:
			end := i + strings.IndexByte(out[i:], 'm')
			active = out[i:end+1] != "\x1b[0m"
			i = end
		case '\n':
			if active {
				t.Fatalf("%v: style is active at a newline in %q", p, out)
			}
		}
	}
	if active {
		t.Fatalf("%v: output %q ends with an active style", p, out)
	}
}

func TestEngineProperties(t *testing.T) {
	rng := rand.New(rand.NewPCG(7, 8))
	for range 1500 {
		m := randomModel(rng, 3)
		for _, p := range allProfiles {
			checkOutput(t, m, p)
		}
	}
}

func FuzzEngine(f *testing.F) {
	f.Add("a\nb", []byte{0, 1, 2}, uint8(4))
	f.Add("\x1b[31m", []byte{}, uint8(0))
	f.Fuzz(func(t *testing.T, text string, entries []byte, profile uint8) {
		cmap := make([]int8, utf8.RuneCountInString(text))
		for i := range cmap {
			cmap[i] = NoStyle
			if i < len(entries) {
				cmap[i] = int8(entries[i]%4) - 1
			}
		}
		m := NewFactory().MustNew(WithText(text), WithColormap(cmap),
			WithPalette(Style{FG: Red}, Style{BG: RGB(1, 2, 3), Attrs: Bold}, Style{FG: TerminalDefault}))
		checkOutput(t, m, Profile(profile%6))
	})
}

func TestEngineConcurrent(t *testing.T) {
	m := randomModel(rand.New(rand.NewPCG(9, 9)), 3)
	e := NewEngine(ProfileTrueColor)
	want := e.String(m)
	var wg sync.WaitGroup
	for range 16 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if got := e.String(m); got != want {
				t.Errorf("got %q, want %q", got, want)
			}
		}()
	}
	wg.Wait()
}
