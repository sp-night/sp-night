package audit

import (
	"io"
	"maps"
	"strings"
	"testing"

	"github.com/sp-night/sp-night/internal/color"
	"github.com/sp-night/sp-night/internal/theme"
)

func mustLoad(t *testing.T) *theme.Palette {
	t.Helper()
	p, _, err := theme.Load()
	if err != nil {
		t.Fatalf("theme.Load: %v", err)
	}
	return p
}

// The whole point of the package: the shipped palette clears every gate. If
// this fails, the palette may not ship.
func TestShippedPaletteClearsEveryGate(t *testing.T) {
	if n := All(mustLoad(t), io.Discard, true); n != 0 {
		t.Errorf("%d fatal pair(s); the shipped palette must have none", n)
	}
}

// 70 is the number the website's /spec page publishes. It is 4 surfaces ×
// (fg + 8 accents + 6 brights + fg_dim + fg_muted) + fiacao on 2 surfaces.
func TestSeventyPairsPerFlavour(t *testing.T) {
	pal := mustLoad(t)
	for _, f := range pal.Flavors {
		if got := len(Flavor(pal, f).Checks); got != 70 {
			t.Errorf("flavor %q measured %d pairs, want 70", f.ID, got)
		}
	}
}

// A gate that never fails is decoration. Drop a comment colour onto the
// background it has to be read against and the audit must catch it.
func TestGateFailsOnAnUnreadableComment(t *testing.T) {
	pal := mustLoad(t)
	f := pal.Flavors[0]

	broken := theme.Flavor{
		ID: f.ID, Label: f.Label, Appearance: f.Appearance,
		Colors: maps.Clone(f.Colors),
	}
	// A grey barely off the laje — the classic unreadable comment.
	broken.Colors["fg_dim"] = "#1e2030"

	rep := Flavor(pal, broken)
	fails := rep.Failures()
	if len(fails) == 0 {
		t.Fatal("audit passed an unreadable comment colour")
	}
	var sawFgDimOnLaje bool
	for _, c := range fails {
		if c.FG == "fg_dim" && c.BG == "laje" {
			sawFgDimOnLaje = true
			if c.Want != LevelAA {
				t.Errorf("fg_dim on laje demanded %.1f, want AA %.1f", c.Want, LevelAA)
			}
		}
	}
	if !sawFgDimOnLaje {
		t.Error("fg_dim on laje was not among the failures")
	}
}

// fg_muted and fiacao are ornament and border. They are reported, never fatal —
// otherwise the palette would have to give up line numbers and indent guides.
func TestOrnamentAndBorderAreWarningsNotGates(t *testing.T) {
	pal := mustLoad(t)
	for _, f := range pal.Flavors {
		for _, c := range Flavor(pal, f).Checks {
			if (c.FG == "fg_muted" || c.FG == "fiacao") && c.Fatal {
				t.Errorf("flavor %q: %s on %s is fatal; it should be a warning", f.ID, c.FG, c.BG)
			}
		}
	}
}

// vidro is selection: a transient state where you look at shape, not text. It
// is held to 3:1 rather than AA, but it is still a gate.
func TestSelectionIsHeldToLargeTextLevel(t *testing.T) {
	pal := mustLoad(t)
	for _, c := range Flavor(pal, pal.Flavors[0]).Checks {
		if c.BG != "vidro" || c.FG == "fg_muted" {
			continue
		}
		if c.Want != LevelLargeAA {
			t.Errorf("%s on vidro demanded %.1f, want %.1f", c.FG, c.Want, LevelLargeAA)
		}
		if !c.Fatal {
			t.Errorf("%s on vidro should still be a gate", c.FG)
		}
	}
}

// The accent ladder keeps this list empty today, and the spec says so. If a
// retune ever fills it, that is a real finding and this test is the messenger.
func TestShippedAccentsAreSeparated(t *testing.T) {
	for _, f := range mustLoad(t).Flavors {
		if pairs := Separation(f); len(pairs) > 0 {
			for _, p := range pairs {
				t.Errorf("flavor %q: %s and %s are neighbouring hues at ΔL %.3f (floor %.2f)",
					f.ID, p.A, p.B, p.Lightness, MinLightness)
			}
		}
	}
}

// The shape of the bug the rule exists for: two accents at the same lightness
// with neighbouring hues. Contrast cannot see it; Separation must.
func TestSeparationCatchesNeighbouringHuesAtEqualLightness(t *testing.T) {
	pal := mustLoad(t)
	f := pal.Flavors[0]
	broken := theme.Flavor{
		ID: f.ID, Label: f.Label, Appearance: f.Appearance,
		Colors: maps.Clone(f.Colors),
	}
	// Rebuild estaiada and sereno at one lightness, a few degrees apart.
	broken.Colors["estaiada"] = color.OklchToRGB(185, 0.07, 0.75).Hex()
	broken.Colors["sereno"] = color.OklchToRGB(200, 0.07, 0.75).Hex()

	pairs := Separation(broken)
	if len(pairs) == 0 {
		t.Fatal("Separation missed two accents at identical lightness")
	}
	var found bool
	for _, p := range pairs {
		if (p.A == "estaiada" && p.B == "sereno") || (p.A == "sereno" && p.B == "estaiada") {
			found = true
			if p.Distance >= NearHue {
				t.Errorf("ΔE %.3f should be below the near-hue threshold %.2f", p.Distance, NearHue)
			}
		}
	}
	if !found {
		t.Errorf("the estaiada/sereno pair was not reported; got %v", pairs)
	}
}

// CVD is a diagnostic. It must produce numbers for all three deficiencies and
// must never contribute to the fatal count.
func TestCVDIsReportedForAllThreeAndNeverFatal(t *testing.T) {
	pal := mustLoad(t)
	f := pal.Flavors[0]
	summaries := CVD(f)
	if len(summaries) != 3 {
		t.Fatalf("got %d CVD summaries, want 3", len(summaries))
	}
	want := map[string]bool{"protanopia": true, "deuteranopia": true, "tritanopia": true}
	for _, s := range summaries {
		if !want[s.Kind] {
			t.Errorf("unexpected CVD kind %q", s.Kind)
		}
		if s.Total != 28 {
			t.Errorf("%s measured %d pairs, want 28 (8 accents choose 2)", s.Kind, s.Total)
		}
	}

	// A palette with eight identical accents is maximally bad under CVD and
	// still must not fail the build.
	flat := theme.Flavor{ID: "flat", Label: "Flat", Appearance: "dark", Colors: maps.Clone(f.Colors)}
	for _, k := range theme.AccentKeys() {
		flat.Colors[k] = f.Colors["sodio"]
	}
	rep := Flavor(pal, flat)
	for _, s := range rep.CVD {
		if s.Tight != 28 {
			t.Errorf("%s: expected all 28 pairs tight for identical accents, got %d", s.Kind, s.Tight)
		}
	}
	var buf strings.Builder
	if n := WriteReport(&buf, []Report{rep}, true); n != 0 {
		// The identical accents still clear contrast, so nothing is fatal —
		// which is precisely the point: CVD does not gate.
		t.Errorf("CVD contributed %d fatal pair(s); it must be diagnostic only", n)
	}
}

func TestReportNamesFlavourAndCounts(t *testing.T) {
	pal := mustLoad(t)
	var buf strings.Builder
	All(pal, &buf, false)
	out := buf.String()
	for _, f := range pal.Flavors {
		if !strings.Contains(out, f.ID) {
			t.Errorf("report does not mention flavour %q", f.ID)
		}
	}
	if !strings.Contains(out, "70 pairs") {
		t.Error("report does not state the pair count")
	}
}
