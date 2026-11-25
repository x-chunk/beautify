package blocks

import "testing"

func TestBadge(t *testing.T) {
	checkText(t, Badge("ok", onBG), unbounded, " ok ")
	checkText(t, Badge("ok", onBG), 4, " ok ")
	checkText(t, Badge("status", onBG), 6, " sta… ")
	checkText(t, Badge("status", onBG), 3, " … ")
	checkText(t, Badge("status", onBG), 2, "  ")
	checkText(t, Badge("status", onBG), 1, " ")
	checkEmpty(t, Badge("status", onBG), 0)
	checkText(t, Badge("a\nb\r\nc", onBG), unbounded, " a b c ")
	checkText(t, Badge("", onBG), unbounded, "  ")
	checkContract(t, "badge", Badge("status", onBG))
	checkContract(t, "wide badge", Badge("漢字 \x1b\t", onBG))
	checkGolden(t, "badge_default", Badge("ok", onBG), unbounded)
	checkGolden(t, "badge_truncated", Badge("status", blue), 6)
}
