package beautify

import (
	"io"
	"os"
	"runtime"
	"strings"

	"go.xchunk.org/beautify/dra"
	"go.xchunk.org/beautify/internal/term"
)

// DetectProfile returns the color profile suitable for w, based on whether w
// is a terminal and on the environment. It honors NO_COLOR, FORCE_COLOR,
// CLICOLOR_FORCE, COLORTERM and TERM; NO_COLOR wins over the variables that
// force color.
func DetectProfile(w io.Writer) dra.Profile {
	return detect(term.IsTerminal(w), runtime.GOOS, os.Getenv)
}

// detect is the decision behind DetectProfile as a pure function. The first
// rule that matches wins.
func detect(isTerminal bool, goos string, getenv func(string) string) dra.Profile {
	termName := getenv("TERM")
	forced := isForced(getenv("FORCE_COLOR")) || isForced(getenv("CLICOLOR_FORCE"))
	switch {
	case !isTerminal && !forced:
		return dra.ProfilePlain
	case termName == "dumb" && !forced:
		return dra.ProfilePlain
	case getenv("NO_COLOR") != "":
		return dra.ProfileNoColor
	case goos == "windows" && getenv("WT_SESSION") == "" && termName == "" && !forced:
		// The legacy console needs a mode switch that the library does not
		// perform, and would show escape sequences as text.
		return dra.ProfilePlain
	}
	switch strings.ToLower(getenv("COLORTERM")) {
	case "truecolor", "24bit":
		return dra.ProfileTrueColor
	}
	switch {
	case strings.HasSuffix(termName, "-direct"):
		return dra.ProfileTrueColor
	case strings.Contains(termName, "256color"):
		return dra.ProfileANSI256
	default:
		return dra.ProfileANSI16
	}
}

// isForced reports whether the value of a variable forces color on.
func isForced(value string) bool {
	return value != "" && value != "0"
}
