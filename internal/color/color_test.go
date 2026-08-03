package color

import (
	"math"
	"testing"
)

func TestParseHex(t *testing.T) {
	got, err := ParseHex("#f2984a")
	if err != nil {
		t.Fatalf("ParseHex: %v", err)
	}
	if want := (RGB{242, 152, 74}); got != want {
		t.Errorf("ParseHex(#f2984a) = %v, want %v", got, want)
	}
	if bare, err := ParseHex("f2984a"); err != nil || bare != got {
		t.Errorf("ParseHex without # = %v, %v; want %v, nil", bare, err, got)
	}
	for _, bad := range []string{"#fff", "#gggggg", "", "#f2984a0"} {
		if _, err := ParseHex(bad); err == nil {
			t.Errorf("ParseHex(%q) = nil error, want a failure", bad)
		}
	}
}

func TestHexRoundTrip(t *testing.T) {
	for _, s := range []string{"#000000", "#ffffff", "#151723", "#05b89e"} {
		if got := MustParseHex(s).Hex(); got != s {
			t.Errorf("round trip %s = %s", s, got)
		}
	}
}

func TestContrast(t *testing.T) {
	white, black := RGB{255, 255, 255}, RGB{0, 0, 0}
	if got := Contrast(white, black); math.Abs(got-21) > 0.01 {
		t.Errorf("Contrast(white, black) = %.4f, want 21", got)
	}
	if got := Contrast(white, white); math.Abs(got-1) > 1e-9 {
		t.Errorf("Contrast(white, white) = %.4f, want 1", got)
	}
	// The ratio must not depend on argument order.
	a, b := MustParseHex("#d3d7eb"), MustParseHex("#151723")
	if Contrast(a, b) != Contrast(b, a) {
		t.Error("Contrast is not symmetric")
	}
}

func TestMixEndpoints(t *testing.T) {
	a, b := RGB{0, 0, 0}, RGB{255, 255, 255}
	if got := Mix(a, b, 0); got != a {
		t.Errorf("Mix at t=0 = %v, want %v", got, a)
	}
	if got := Mix(a, b, 1); got != b {
		t.Errorf("Mix at t=1 = %v, want %v", got, b)
	}
	// Out-of-range t clamps rather than extrapolating into nonsense.
	if got := Mix(a, b, 2); got != b {
		t.Errorf("Mix at t=2 = %v, want %v (clamped)", got, b)
	}
	if got := Lighten(a, 1); got != b {
		t.Errorf("Lighten(black, 1) = %v, want white", got)
	}
	if got := Darken(b, 1); got != a {
		t.Errorf("Darken(white, 1) = %v, want black", got)
	}
}

func TestOklabLightnessEndpoints(t *testing.T) {
	if l := (RGB{0, 0, 0}).Oklab().L; math.Abs(l) > 1e-6 {
		t.Errorf("Oklab L of black = %.6f, want 0", l)
	}
	if l := (RGB{255, 255, 255}).Oklab().L; math.Abs(l-1) > 1e-3 {
		t.Errorf("Oklab L of white = %.6f, want 1", l)
	}
}

func TestDistanceIsZeroForSameColour(t *testing.T) {
	c := MustParseHex("#6e92de")
	if d := Distance(c, c); d != 0 {
		t.Errorf("Distance(c, c) = %.6f, want 0", d)
	}
}

// Why the accent-separation rule needs Oklab and not contrast: two colours
// built at the same Oklab lightness sit at effectively the same WCAG ratio on
// any background, so contrast alone cannot tell them apart. Distance can.
//
// This is the shape of the bug that got estaiada and sereno past the first
// version of the audit. The pair has since been retuned, so the property is
// asserted directly rather than against those two hexes.
func TestContrastCannotSeparateHuesAtEqualLightness(t *testing.T) {
	const lightness = 0.72
	laje := MustParseHex("#151723")
	teal := OklchToRGB(180, 0.08, lightness)
	cyan := OklchToRGB(215, 0.08, lightness)

	ct, cc := Contrast(teal, laje), Contrast(cyan, laje)
	if math.Abs(ct-cc) > 0.25 {
		t.Fatalf("equal-lightness colours should have near-equal contrast, got %.2f and %.2f", ct, cc)
	}
	if d := Distance(teal, cyan); d < 0.02 {
		t.Errorf("Oklab distance %.4f is below the perceptible threshold; these should be separable", d)
	}
	if dl := math.Abs(teal.Oklab().L - cyan.Oklab().L); dl > 0.01 {
		t.Errorf("built at the same lightness but ΔL = %.4f", dl)
	}
}

func TestOklchToRGBStaysInGamut(t *testing.T) {
	// A wildly out-of-gamut request must come back as a real sRGB colour
	// rather than wrapping around a channel.
	for hue := 0.0; hue < 360; hue += 30 {
		c := OklchToRGB(hue, 0.9, 0.7)
		if _, err := ParseHex(c.Hex()); err != nil {
			t.Fatalf("hue %.0f produced an unparseable colour %v", hue, c)
		}
	}
}

func TestSimulateIsDefinedForEveryKind(t *testing.T) {
	c := MustParseHex("#ea5d5d")
	if len(CVDKinds) != 3 {
		t.Fatalf("want 3 CVD kinds, got %d", len(CVDKinds))
	}
	for _, k := range CVDKinds {
		if k.Name == "" {
			t.Error("a CVD kind has no name")
		}
		// Red under deuteranopia must actually move.
		if k.Name == "deuteranopia" && Simulate(c, k) == c {
			t.Error("deuteranopia left pure red unchanged")
		}
	}
}
