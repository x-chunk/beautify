package dra

import "testing"

func TestProfileString(t *testing.T) {
	tests := map[Profile]string{
		ProfilePlain:     "plain",
		ProfileNoColor:   "no-color",
		ProfileANSI16:    "ansi16",
		ProfileANSI256:   "ansi256",
		ProfileTrueColor: "truecolor",
		Profile(200):     "plain",
	}
	for p, want := range tests {
		if got := p.String(); got != want {
			t.Errorf("Profile(%d).String() = %q, want %q", p, got, want)
		}
	}
	var zero Profile
	if zero != ProfilePlain {
		t.Error("zero profile is not ProfilePlain")
	}
}

func TestZeroModeIsReal(t *testing.T) {
	var zero Mode
	if zero != ModeReal {
		t.Error("zero mode is not ModeReal")
	}
}
