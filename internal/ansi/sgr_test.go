package ansi

import "testing"

func idx16(n uint8) Color     { return Color{Kind: Indexed16, Index: n} }
func idx256(n uint8) Color    { return Color{Kind: Indexed256, Index: n} }
func rgb(r, g, b uint8) Color { return Color{Kind: RGB, R: r, G: g, B: b} }

func TestAppendSGR(t *testing.T) {
	tests := []struct {
		name  string
		style Style
		want  string
	}{
		{"empty", Style{}, "\x1b[0m"},
		{"bold red", Style{FG: idx16(1), Attrs: Bold}, "\x1b[0;1;31m"},
		{"bright fg, bg", Style{FG: idx16(9), BG: idx16(4)}, "\x1b[0;91;44m"},
		{"bold underline 256", Style{FG: idx256(196), Attrs: Bold | Underline}, "\x1b[0;1;4;38;5;196m"},
		{"rgb both", Style{FG: rgb(255, 128, 0), BG: rgb(0, 0, 0)}, "\x1b[0;38;2;255;128;0;48;2;0;0;0m"},
		{"fg 0", Style{FG: idx16(0)}, "\x1b[0;30m"},
		{"fg 7", Style{FG: idx16(7)}, "\x1b[0;37m"},
		{"fg 8", Style{FG: idx16(8)}, "\x1b[0;90m"},
		{"fg 15", Style{FG: idx16(15)}, "\x1b[0;97m"},
		{"bg 0", Style{BG: idx16(0)}, "\x1b[0;40m"},
		{"bg 7", Style{BG: idx16(7)}, "\x1b[0;47m"},
		{"bg 8", Style{BG: idx16(8)}, "\x1b[0;100m"},
		{"bg 15", Style{BG: idx16(15)}, "\x1b[0;107m"},
		{"bg 256", Style{BG: idx256(16)}, "\x1b[0;48;5;16m"},
		{"no leading zeros", Style{FG: rgb(1, 20, 100)}, "\x1b[0;38;2;1;20;100m"},
		{"all attributes in order", Style{Attrs: Strike | Reverse | Blink | Underline | Italic | Faint | Bold}, "\x1b[0;1;2;3;4;5;7;9m"},
		{"default emits nothing", Style{FG: Color{Kind: Default}, BG: Color{Kind: Default}}, "\x1b[0m"},
		{"attributes only", Style{Attrs: Italic}, "\x1b[0;3m"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := string(AppendSGR([]byte("x"), tt.style)); got != "x"+tt.want {
				t.Errorf("AppendSGR() = %q, want %q", got, "x"+tt.want)
			}
		})
	}
}

func TestAppendReset(t *testing.T) {
	if got := string(AppendReset([]byte("x"))); got != "x\x1b[0m" {
		t.Errorf("AppendReset() = %q", got)
	}
}

func TestAppendDoesNotAllocate(t *testing.T) {
	buf := make([]byte, 0, 128)
	s := Style{FG: rgb(255, 128, 64), BG: idx256(200), Attrs: Bold | Underline}
	allocs := testing.AllocsPerRun(100, func() {
		buf = AppendSGR(buf[:0], s)
		buf = AppendReset(buf)
	})
	if allocs != 0 {
		t.Errorf("allocations = %v, want 0", allocs)
	}
}
