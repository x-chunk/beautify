package beautify

import (
	"bytes"
	"testing"

	"go.xchunk.org/beautify/dra"
)

func TestDetect(t *testing.T) {
	type env map[string]string
	tests := []struct {
		name string
		tty  bool
		goos string
		env  env
		want dra.Profile
	}{
		{"not a terminal", false, "linux", env{"TERM": "xterm-256color", "COLORTERM": "truecolor"}, dra.ProfilePlain},
		{"not a terminal, forced", false, "linux", env{"FORCE_COLOR": "1"}, dra.ProfileANSI16},
		{"not a terminal, CLICOLOR_FORCE", false, "linux", env{"CLICOLOR_FORCE": "1", "TERM": "xterm-256color"}, dra.ProfileANSI256},
		{"FORCE_COLOR=0 does not force", false, "linux", env{"FORCE_COLOR": "0"}, dra.ProfilePlain},
		{"empty FORCE_COLOR does not force", false, "linux", env{"FORCE_COLOR": ""}, dra.ProfilePlain},
		{"CLICOLOR_FORCE=0 does not force", false, "linux", env{"CLICOLOR_FORCE": "0"}, dra.ProfilePlain},
		{"one of two force variables", false, "linux", env{"FORCE_COLOR": "0", "CLICOLOR_FORCE": "yes"}, dra.ProfileANSI16},

		{"dumb", true, "linux", env{"TERM": "dumb"}, dra.ProfilePlain},
		{"dumb wins over COLORTERM", true, "linux", env{"TERM": "dumb", "COLORTERM": "truecolor"}, dra.ProfilePlain},
		{"dumb wins over NO_COLOR", true, "linux", env{"TERM": "dumb", "NO_COLOR": "1"}, dra.ProfilePlain},
		{"forced dumb", true, "linux", env{"TERM": "dumb", "FORCE_COLOR": "1"}, dra.ProfileANSI16},
		{"forced dumb with COLORTERM", true, "linux", env{"TERM": "dumb", "FORCE_COLOR": "1", "COLORTERM": "truecolor"}, dra.ProfileTrueColor},

		{"NO_COLOR", true, "linux", env{"NO_COLOR": "1", "TERM": "xterm"}, dra.ProfileNoColor},
		{"NO_COLOR=0 still counts", true, "linux", env{"NO_COLOR": "0"}, dra.ProfileNoColor},
		{"empty NO_COLOR", true, "linux", env{"NO_COLOR": "", "TERM": "xterm"}, dra.ProfileANSI16},
		{"NO_COLOR wins over FORCE_COLOR", true, "linux", env{"NO_COLOR": "1", "FORCE_COLOR": "1"}, dra.ProfileNoColor},
		{"NO_COLOR and force without a terminal", false, "linux", env{"NO_COLOR": "1", "FORCE_COLOR": "1"}, dra.ProfileNoColor},
		{"NO_COLOR without a terminal", false, "linux", env{"NO_COLOR": "1"}, dra.ProfilePlain},
		{"NO_COLOR wins over COLORTERM", true, "linux", env{"NO_COLOR": "1", "COLORTERM": "truecolor"}, dra.ProfileNoColor},
		{"NO_COLOR comes before the Windows rule", true, "windows", env{"NO_COLOR": "1"}, dra.ProfileNoColor},

		{"legacy Windows console", true, "windows", nil, dra.ProfilePlain},
		{"legacy Windows console with COLORTERM", true, "windows", env{"COLORTERM": "truecolor"}, dra.ProfilePlain},
		{"Windows Terminal", true, "windows", env{"WT_SESSION": "id"}, dra.ProfileANSI16},
		{"Windows with TERM", true, "windows", env{"TERM": "xterm-256color"}, dra.ProfileANSI256},
		{"forced on Windows", true, "windows", env{"FORCE_COLOR": "1"}, dra.ProfileANSI16},
		{"no TERM elsewhere", true, "linux", nil, dra.ProfileANSI16},

		{"COLORTERM=truecolor", true, "linux", env{"COLORTERM": "truecolor", "TERM": "xterm-256color"}, dra.ProfileTrueColor},
		{"COLORTERM=24bit", true, "linux", env{"COLORTERM": "24bit"}, dra.ProfileTrueColor},
		{"COLORTERM ignores case", true, "linux", env{"COLORTERM": "TrueColor"}, dra.ProfileTrueColor},
		{"other COLORTERM", true, "linux", env{"COLORTERM": "yes", "TERM": "xterm"}, dra.ProfileANSI16},
		{"direct color", true, "linux", env{"TERM": "xterm-direct"}, dra.ProfileTrueColor},
		{"direct color in the middle", true, "linux", env{"TERM": "xterm-direct2"}, dra.ProfileANSI16},
		{"256 colors", true, "linux", env{"TERM": "screen-256color-bce"}, dra.ProfileANSI256},
		{"plain xterm", true, "linux", env{"TERM": "xterm"}, dra.ProfileANSI16},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getenv := func(key string) string { return tt.env[key] }
			if got := detect(tt.tty, tt.goos, getenv); got != tt.want {
				t.Errorf("detect() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDetectProfileOfBuffer(t *testing.T) {
	t.Setenv("FORCE_COLOR", "")
	t.Setenv("CLICOLOR_FORCE", "")
	if got := DetectProfile(&bytes.Buffer{}); got != dra.ProfilePlain {
		t.Errorf("DetectProfile() = %v", got)
	}
}
