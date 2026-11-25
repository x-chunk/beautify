package width

import (
	"math/rand/v2"
	"strings"
	"testing"
	"unicode/utf8"
)

type cluster struct {
	text  string
	width int
}

func split(t *testing.T, s string, mode Mode) []cluster {
	t.Helper()
	var out []cluster
	for s != "" {
		size, w := Next(s, mode)
		if size < 1 || size > len(s) {
			t.Fatalf("Next(%q) size = %d", s, size)
		}
		out = append(out, cluster{s[:size], w})
		s = s[size:]
	}
	return out
}

func TestNextEmpty(t *testing.T) {
	for _, mode := range []Mode{Real, Long} {
		if size, w := Next("", mode); size != 0 || w != 0 {
			t.Errorf("Next(\"\", %d) = %d, %d, want 0, 0", mode, size, w)
		}
		if w := String("", mode); w != 0 {
			t.Errorf("String(\"\", %d) = %d, want 0", mode, w)
		}
	}
}

func TestNextLong(t *testing.T) {
	const (
		man, woman, girl = "\U0001F468", "\U0001F469", "\U0001F467"
		zwj              = "‍"
		riA, riB, riC    = "\U0001F1E6", "\U0001F1E7", "\U0001F1E8"
	)
	tests := []struct {
		name string
		in   string
		want []cluster
	}{
		{"ascii", "ab", []cluster{{"a", 1}, {"b", 1}}},
		{"combining mark", "éx", []cluster{{"é", 1}, {"x", 1}}},
		{"two combining marks", "é̂", []cluster{{"é̂", 1}}},
		{"spacing mark", "कः", []cluster{{"कः", 1}}},
		{"combining mark without base", "́a", []cluster{{"́", 0}, {"a", 1}}},
		{"cjk", "漢字", []cluster{{"漢", 2}, {"字", 2}}},
		{"hangul syllable", "한", []cluster{{"한", 2}}},
		{"skin tone", "\U0001F44D\U0001F3FD", []cluster{{"\U0001F44D\U0001F3FD", 2}}},
		{"family", man + zwj + woman + zwj + girl, []cluster{{man + zwj + woman + zwj + girl, 2}}},
		{"trailing zwj", man + zwj, []cluster{{man + zwj, 2}}},
		{"zwj before newline", man + zwj + "\nx", []cluster{{man + zwj, 2}, {"\n", 1}, {"x", 1}}},
		{"zwj before text", "a" + zwj + "b", []cluster{{"a" + zwj + "b", 1}}},
		{"extender after zwj rune", "a" + zwj + "b́c", []cluster{{"a" + zwj + "b́", 1}, {"c", 1}}},
		{"leading zwj is its own base", zwj + "a", []cluster{{zwj, 0}, {"a", 1}}},
		{"keycap", "1️⃣", []cluster{{"1️⃣", 2}}},
		{"emoji presentation", "❤️", []cluster{{"❤️", 2}}},
		{"text presentation", "❤︎", []cluster{{"❤︎", 1}}},
		{"text presentation of wide", "⌚︎", []cluster{{"⌚︎", 2}}},
		{"variation selector supplement", "漢\U000E0100", []cluster{{"漢\U000E0100", 2}}},
		{"tag sequence", "\U0001F3F4\U000E0067\U000E007F", []cluster{{"\U0001F3F4\U000E0067\U000E007F", 2}}},
		{"flag", riA + riB, []cluster{{riA + riB, 2}}},
		{"three indicators", riA + riB + riC, []cluster{{riA + riB, 2}, {riC, 1}}},
		{"lone indicator", riA + "x", []cluster{{riA, 1}, {"x", 1}}},
		{"flag with extender", riA + riB + "́x", []cluster{{riA + riB + "́", 2}, {"x", 1}}},
		{"zero-width space", "a​b", []cluster{{"a", 1}, {"​", 0}, {"b", 1}}},
		{"invalid bytes", "\xff\xfe", []cluster{{"\xff", 1}, {"\xfe", 1}}},
		{"invalid byte is not a base", "\xff́", []cluster{{"\xff", 1}, {"́", 0}}},
		{"invalid byte after zwj", "a" + zwj + "\xff", []cluster{{"a" + zwj, 1}, {"\xff", 1}}},
		{"newline is not a base", "\ń", []cluster{{"\n", 1}, {"́", 0}}},
		{"ascii then combining", "ab́", []cluster{{"a", 1}, {"b́", 1}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := split(t, tt.in, Long)
			if len(got) != len(tt.want) {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
			total := 0
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("cluster %d = %q, want %q", i, got[i], tt.want[i])
				}
				total += got[i].width
			}
			if w := String(tt.in, Long); w != total {
				t.Errorf("String() = %d, want %d", w, total)
			}
		})
	}
}

func TestNextRealCountsRunes(t *testing.T) {
	inputs := []string{
		"abc", "é", "漢字", "\U0001F468‍\U0001F469", "\U0001F1E6\U0001F1E7",
		"1️⃣", "\xff\xfe", "a​b", "́", "тест",
	}
	for _, in := range inputs {
		clusters := split(t, in, Real)
		if len(clusters) != utf8.RuneCountInString(in) {
			t.Errorf("%q: %d clusters, want %d", in, len(clusters), utf8.RuneCountInString(in))
		}
		for _, c := range clusters {
			if c.width != 1 {
				t.Errorf("%q: cluster %q has width %d, want 1", in, c.text, c.width)
			}
		}
		if w := String(in, Real); w != utf8.RuneCountInString(in) {
			t.Errorf("String(%q) = %d", in, w)
		}
	}
}

// alphabet holds runes that exercise every segmentation rule.
var alphabet = []string{
	"a", " ", "я", "漢", "́", "ः", "‍", "️", "︎",
	"\U0001F468", "\U0001F3FD", "\U0001F1E6", "\U0001F1E7", "ᄒ", "ᅡ",
	"​", "\n", "\xff", "­", "\U000E0067",
}

func randomText(rng *rand.Rand, n int) string {
	var b strings.Builder
	for range n {
		b.WriteString(alphabet[rng.IntN(len(alphabet))])
	}
	return b.String()
}

func TestNextCoversWholeString(t *testing.T) {
	rng := rand.New(rand.NewPCG(1, 2))
	for range 2000 {
		s := randomText(rng, rng.IntN(12))
		for _, mode := range []Mode{Real, Long} {
			var b strings.Builder
			for _, c := range split(t, s, mode) {
				b.WriteString(c.text)
			}
			if b.String() != s {
				t.Fatalf("mode %d: clusters of %q join to %q", mode, s, b.String())
			}
		}
	}
}

func TestSeparable(t *testing.T) {
	tests := []struct {
		name string
		a, b string
		want bool
	}{
		{"plain", "a", "b", true},
		{"empty left", "", "́", true},
		{"empty right", "a", "", true},
		{"combining mark", "a", "́", false},
		{"after joiner", "a‍", "b", false},
		{"before joiner", "a", "‍b", false},
		{"indicators", "\U0001F1E6", "\U0001F1E7", false},
		{"indicator and text", "\U0001F1E6", "a", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Separable(tt.a, tt.b, Long); got != tt.want {
				t.Errorf("Separable() = %v, want %v", got, tt.want)
			}
			if !Separable(tt.a, tt.b, Real) {
				t.Error("Separable() = false in Real Mode")
			}
		})
	}
}

func TestSeparableKeepsWidth(t *testing.T) {
	rng := rand.New(rand.NewPCG(3, 4))
	for range 5000 {
		a, b := randomText(rng, 1+rng.IntN(4)), randomText(rng, 1+rng.IntN(4))
		if !Separable(a, b, Long) {
			continue
		}
		if got, want := String(a+b, Long), String(a, Long)+String(b, Long); got != want {
			t.Fatalf("%q + %q: joined width %d, sum %d", a, b, got, want)
		}
	}
}

func FuzzNext(f *testing.F) {
	for _, s := range alphabet {
		f.Add(s + s)
	}
	f.Add("a‍\n\xff\U0001F1E6\U0001F1E7́")
	f.Fuzz(func(t *testing.T, s string) {
		for _, mode := range []Mode{Real, Long} {
			for rest := s; rest != ""; {
				size, w := Next(rest, mode)
				if size < 1 || size > len(rest) {
					t.Fatalf("mode %d: Next(%q) size = %d", mode, rest, size)
				}
				if w < 0 || w > 2 {
					t.Fatalf("mode %d: Next(%q) width = %d", mode, rest, w)
				}
				rest = rest[size:]
			}
		}
	})
}

func BenchmarkString(b *testing.B) {
	texts := []struct{ name, s string }{
		{"ASCII", strings.Repeat("hello, world ", 20)},
		{"Cyrillic", strings.Repeat("привет, мир ", 20)},
		{"CJK", strings.Repeat("漢字仮名交じり文", 20)},
		{"Emoji", strings.Repeat("\U0001F468‍\U0001F469‍\U0001F467\U0001F44D\U0001F3FD", 20)},
	}
	modes := []struct {
		name string
		mode Mode
	}{{"Real", Real}, {"Long", Long}}
	for _, text := range texts {
		for _, m := range modes {
			b.Run(text.name+"/"+m.name, func(b *testing.B) {
				b.SetBytes(int64(len(text.s)))
				b.ReportAllocs()
				for range b.N {
					String(text.s, m.mode)
				}
			})
		}
	}
}
