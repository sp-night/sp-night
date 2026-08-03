// Package color is the colour maths the SP Night audits are built on: sRGB,
// WCAG relative luminance and contrast, and Oklab/Oklch for the questions
// contrast cannot answer.
package color

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// RGB is an 8-bit-per-channel sRGB colour.
type RGB struct{ R, G, B uint8 }

// ParseHex reads "#rrggbb" (the leading # is optional).
func ParseHex(s string) (RGB, error) {
	h := strings.TrimPrefix(strings.TrimSpace(s), "#")
	if len(h) != 6 {
		return RGB{}, fmt.Errorf("invalid hex %q: want #rrggbb", s)
	}
	v, err := strconv.ParseUint(h, 16, 32)
	if err != nil {
		return RGB{}, fmt.Errorf("invalid hex %q: %w", s, err)
	}
	return RGB{uint8(v >> 16), uint8(v >> 8 & 0xff), uint8(v & 0xff)}, nil
}

// MustParseHex is ParseHex for call sites that run after Palette.Validate has
// already proved every hex parses. Emitting a silently wrong file is worse
// than crashing.
func MustParseHex(s string) RGB {
	c, err := ParseHex(s)
	if err != nil {
		panic(err)
	}
	return c
}

// Hex renders the colour as "#rrggbb", lowercase.
func (c RGB) Hex() string { return fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B) }

// linear converts an sRGB channel to linear light (WCAG 2.x).
func linear(ch uint8) float64 {
	v := float64(ch) / 255.0
	if v <= 0.04045 {
		return v / 12.92
	}
	return math.Pow((v+0.055)/1.055, 2.4)
}

// delinear is the inverse of linear: linear light back to an sRGB channel.
func delinear(v float64) uint8 {
	v = clamp01(v)
	if v <= 0.0031308 {
		v *= 12.92
	} else {
		v = 1.055*math.Pow(v, 1/2.4) - 0.055
	}
	return uint8(math.Round(v * 255))
}

// Luminance is WCAG relative luminance.
func (c RGB) Luminance() float64 {
	return 0.2126*linear(c.R) + 0.7152*linear(c.G) + 0.0722*linear(c.B)
}

// Contrast is the WCAG contrast ratio between two colours, 1.0 to 21.0.
func Contrast(a, b RGB) float64 {
	la, lb := a.Luminance(), b.Luminance()
	if la < lb {
		la, lb = lb, la
	}
	return (la + 0.05) / (lb + 0.05)
}

func clamp01(f float64) float64 { return math.Max(0, math.Min(1, f)) }

// Clamp01 clamps an alpha or ratio into [0,1] for the template helpers.
func Clamp01(f float64) float64 { return clamp01(f) }

func lerp(a, b uint8, t float64) uint8 {
	return uint8(math.Round(float64(a) + (float64(b)-float64(a))*t))
}

// Mix interpolates two colours in sRGB space. t=0 returns a, t=1 returns b.
func Mix(a, b RGB, t float64) RGB {
	t = clamp01(t)
	return RGB{lerp(a.R, b.R, t), lerp(a.G, b.G, t), lerp(a.B, b.B, t)}
}

// Lighten mixes towards white, Darken towards black.
func Lighten(c RGB, t float64) RGB { return Mix(c, RGB{255, 255, 255}, t) }
func Darken(c RGB, t float64) RGB  { return Mix(c, RGB{0, 0, 0}, t) }

// --- Oklab ---------------------------------------------------------------
//
// WCAG contrast only measures luminance: it tells you whether text is legible
// on a background, not whether two colours are distinguishable from each
// other. Two accents can sit at 1.00:1 against each other (identical
// luminance) and still be obvious — red and green. They can also have
// neighbouring hues and the same luminance, and then they really are hard to
// tell apart.
//
// Oklab is perceptually uniform: Euclidean distance in it approximates "how
// different do these colours look".

// Oklab is a colour in the Oklab perceptual space.
type Oklab struct{ L, A, B float64 }

// Oklab converts the colour into Oklab.
func (c RGB) Oklab() Oklab {
	r, g, b := linear(c.R), linear(c.G), linear(c.B)

	l := 0.4122214708*r + 0.5363325363*g + 0.0514459929*b
	m := 0.2119034982*r + 0.6806995451*g + 0.1073969566*b
	s := 0.0883024619*r + 0.2817188376*g + 0.6299787005*b

	l, m, s = math.Cbrt(l), math.Cbrt(m), math.Cbrt(s)

	return Oklab{
		L: 0.2104542553*l + 0.7936177850*m - 0.0040720468*s,
		A: 1.9779984951*l - 2.4285922050*m + 0.4505937099*s,
		B: 0.0259040371*l + 0.7827717662*m - 0.8086757660*s,
	}
}

// Distance is the perceptual difference between two colours (ΔE in Oklab).
// For reference: ~0.02 is the threshold of the perceptible side by side; a
// theme's accents need considerably more than that to avoid being confused
// with each other across tokens scattered over a screen.
func Distance(a, b RGB) float64 {
	x, y := a.Oklab(), b.Oklab()
	dl, da, db := x.L-y.L, x.A-y.A, x.B-y.B
	return math.Sqrt(dl*dl + da*da + db*db)
}

// HueAngle is the Oklab hue angle in degrees (0-360). Rough reference:
// 30 red, 70 orange, 100 yellow, 145 green, 195 cyan, 265 blue, 320 magenta.
func HueAngle(c RGB) float64 {
	o := c.Oklab()
	deg := math.Atan2(o.B, o.A) * 180 / math.Pi
	if deg < 0 {
		deg += 360
	}
	return deg
}

// Chroma is perceived saturation in Oklab.
func Chroma(c RGB) float64 {
	o := c.Oklab()
	return math.Hypot(o.A, o.B)
}

// --- Oklch -> sRGB ------------------------------------------------------
//
// Needed to reason in (hue, chroma, lightness) and get back to hex. Picking a
// colour by nudging an RGB channel is guessing; nudging Oklch is saying "this
// blue, a little deeper" and getting exactly that.

func oklabToLinear(o Oklab) (r, g, b float64) {
	l_ := o.L + 0.3963377774*o.A + 0.2158037573*o.B
	m_ := o.L - 0.1055613458*o.A - 0.0638541728*o.B
	s_ := o.L - 0.0894841775*o.A - 1.2914855480*o.B
	l, m, s := l_*l_*l_, m_*m_*m_, s_*s_*s_
	r = +4.0767416621*l - 3.3077115913*m + 0.2309699292*s
	g = -1.2684380046*l + 2.6097574011*m - 0.3413193965*s
	b = -0.0041960863*l - 0.7034186147*m + 1.7076147010*s
	return
}

func inGamut(o Oklab) bool {
	r, g, b := oklabToLinear(o)
	const eps = 1e-4
	return r >= -eps && r <= 1+eps && g >= -eps && g <= 1+eps && b >= -eps && b <= 1+eps
}

// OklchToRGB converts hue (degrees), chroma and lightness to sRGB, reducing
// chroma by binary search when the colour falls outside the gamut.
func OklchToRGB(hueDeg, chroma, lightness float64) RGB {
	rad := hueDeg * math.Pi / 180
	mk := func(c float64) Oklab {
		return Oklab{L: lightness, A: c * math.Cos(rad), B: c * math.Sin(rad)}
	}
	if !inGamut(mk(chroma)) {
		lo, hi := 0.0, chroma
		for range 24 {
			mid := (lo + hi) / 2
			if inGamut(mk(mid)) {
				lo = mid
			} else {
				hi = mid
			}
		}
		chroma = lo
	}
	r, g, b := oklabToLinear(mk(chroma))
	return RGB{delinear(r), delinear(g), delinear(b)}
}

// --- Colour vision deficiency ------------------------------------------
//
// Around 8% of men have some form of colour blindness, almost always on the
// red-green axis. A theme that separates tokens by hue alone on that axis —
// green for string, red for tag — collapses for those people. The blue-yellow
// axis is preserved in protanopia and deuteranopia, and it is also where
// luminance separates most naturally.
//
// Matrices from Viénot, Brettel & Mollon (1999), applied in linear RGB.

// CVDKind is one simulated colour vision deficiency.
type CVDKind struct {
	Name   string
	Matrix [9]float64
}

// CVDKinds are the three deficiencies the audit reports on.
var CVDKinds = []CVDKind{
	{"protanopia", [9]float64{ // ~1% of men: no L (red) cones
		0.11238, 0.88762, 0.00000,
		0.11238, 0.88762, 0.00000,
		0.00401, -0.00401, 1.00000,
	}},
	{"deuteranopia", [9]float64{ // ~6% of men: no M (green) cones
		0.29275, 0.70725, 0.00000,
		0.29275, 0.70725, 0.00000,
		-0.02234, 0.02234, 1.00000,
	}},
	{"tritanopia", [9]float64{ // rare: no S (blue) cones
		1.00000, 0.14461, -0.14461,
		0.00000, 0.85653, 0.14347,
		0.00000, 0.85653, 0.14347,
	}},
}

// Simulate returns the colour as perceived with the given deficiency.
func Simulate(c RGB, k CVDKind) RGB {
	r, g, b := linear(c.R), linear(c.G), linear(c.B)
	m := k.Matrix
	return RGB{
		delinear(m[0]*r + m[1]*g + m[2]*b),
		delinear(m[3]*r + m[4]*g + m[5]*b),
		delinear(m[6]*r + m[7]*g + m[8]*b),
	}
}
