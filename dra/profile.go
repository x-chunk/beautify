package dra

// Profile is the color capability the Engine renders for. Values outside
// the defined range are treated as ProfilePlain.
type Profile uint8

// Profiles in order of increasing capability. The zero value is
// ProfilePlain, so an uninitialized Engine emits plain text.
const (
	ProfilePlain     Profile = iota // no escape sequences at all
	ProfileNoColor                  // attributes only
	ProfileANSI16                   // 16-color palette
	ProfileANSI256                  // 256-color palette
	ProfileTrueColor                // 24-bit colors
)

// String returns a short name of the profile, such as "ansi256".
func (p Profile) String() string {
	switch p {
	case ProfileNoColor:
		return "no-color"
	case ProfileANSI16:
		return "ansi16"
	case ProfileANSI256:
		return "ansi256"
	case ProfileTrueColor:
		return "truecolor"
	default:
		return "plain"
	}
}
