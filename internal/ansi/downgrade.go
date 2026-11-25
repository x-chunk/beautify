package ansi

// attrMask covers every defined attribute bit.
const attrMask = Strike<<1 - 1

// cubeLevels are the channel values of the 6x6x6 color cube.
var cubeLevels = [6]uint8{0, 95, 135, 175, 215, 255}

// palette16 holds the xterm defaults of the 16-color palette. Terminals
// remap these colors, so they are only used to find the nearest color.
var palette16 = [16][3]uint8{
	{0x00, 0x00, 0x00}, {0x80, 0x00, 0x00}, {0x00, 0x80, 0x00}, {0x80, 0x80, 0x00},
	{0x00, 0x00, 0x80}, {0x80, 0x00, 0x80}, {0x00, 0x80, 0x80}, {0xc0, 0xc0, 0xc0},
	{0x80, 0x80, 0x80}, {0xff, 0x00, 0x00}, {0x00, 0xff, 0x00}, {0xff, 0xff, 0x00},
	{0x00, 0x00, 0xff}, {0xff, 0x00, 0xff}, {0x00, 0xff, 0xff}, {0xff, 0xff, 0xff},
}

// Downgrade converts s to what level can express. The result is canonical:
// styles that look the same at this level compare equal. Unknown levels are
// treated as Plain.
func Downgrade(s Style, level Level) Style {
	switch level {
	case NoColor:
		return Style{Attrs: s.Attrs & attrMask}
	case ANSI16, ANSI256, TrueColor:
		return Style{
			FG:    downgradeColor(s.FG, level),
			BG:    downgradeColor(s.BG, level),
			Attrs: s.Attrs & attrMask,
		}
	default:
		return Style{}
	}
}

func downgradeColor(c Color, level Level) Color {
	if c.Kind == Indexed256 && c.Index < 16 {
		c.Kind = Indexed16
	}
	switch c.Kind {
	case Indexed16:
		return Color{Kind: Indexed16, Index: c.Index % 16}
	case Indexed256:
		if level == ANSI16 {
			return nearest16(ToRGB(c))
		}
		return Color{Kind: Indexed256, Index: c.Index}
	case RGB:
		switch level {
		case ANSI16:
			return nearest16(c.R, c.G, c.B)
		case ANSI256:
			return nearest256(c.R, c.G, c.B)
		}
		return Color{Kind: RGB, R: c.R, G: c.G, B: c.B}
	default:
		return Color{}
	}
}

// ToRGB returns the channel values of c. Palette colors use the xterm
// defaults. None and Default have no value and yield black.
func ToRGB(c Color) (r, g, b uint8) {
	switch c.Kind {
	case RGB:
		return c.R, c.G, c.B
	case Indexed16:
		p := palette16[c.Index%16]
		return p[0], p[1], p[2]
	case Indexed256:
		switch {
		case c.Index < 16:
			p := palette16[c.Index]
			return p[0], p[1], p[2]
		case c.Index < 232:
			n := c.Index - 16
			return cubeLevels[n/36], cubeLevels[n/6%6], cubeLevels[n%6]
		default:
			v := 8 + 10*(c.Index-232)
			return v, v, v
		}
	default:
		return 0, 0, 0
	}
}

// nearest16 returns the closest color of the 16-color palette. Ties go to
// the lower index.
func nearest16(r, g, b uint8) Color {
	best, bestDist := 0, -1
	for i, p := range palette16 {
		d := distance(r, g, b, p[0], p[1], p[2])
		if bestDist < 0 || d < bestDist {
			best, bestDist = i, d
		}
	}
	return Color{Kind: Indexed16, Index: uint8(best)}
}

// nearest256 returns the closest cube or gray color. Indexes 0-15 are never
// chosen because they depend on the terminal theme. On a tie the cube wins.
func nearest256(r, g, b uint8) Color {
	ri, gi, bi := nearestLevel(r), nearestLevel(g), nearestLevel(b)
	cubeDist := distance(r, g, b, cubeLevels[ri], cubeLevels[gi], cubeLevels[bi])

	gi24 := nearestGray(r, g, b)
	v := 8 + 10*gi24
	if distance(r, g, b, v, v, v) < cubeDist {
		return Color{Kind: Indexed256, Index: 232 + gi24}
	}
	return Color{Kind: Indexed256, Index: 16 + 36*ri + 6*gi + bi}
}

// nearestLevel returns the index of the closest cube level. Ties go to the
// lower level.
func nearestLevel(v uint8) uint8 {
	var best uint8
	for i := range uint8(len(cubeLevels)) {
		if abs(int(v)-int(cubeLevels[i])) < abs(int(v)-int(cubeLevels[best])) {
			best = i
		}
	}
	return best
}

// nearestGray returns the step of the gray ramp closest to the channel
// average. It compares sums instead of averages to stay exact. Ties go to
// the lower step.
func nearestGray(r, g, b uint8) uint8 {
	sum := int(r) + int(g) + int(b)
	var best uint8
	for i := range uint8(24) {
		if abs(sum-3*(8+10*int(i))) < abs(sum-3*(8+10*int(best))) {
			best = i
		}
	}
	return best
}

// distance returns the squared RGB distance.
func distance(r1, g1, b1, r2, g2, b2 uint8) int {
	dr, dg, db := int(r1)-int(r2), int(g1)-int(g2), int(b1)-int(b2)
	return dr*dr + dg*dg + db*db
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
