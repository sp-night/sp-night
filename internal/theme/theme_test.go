package theme

import (
	"slices"
	"strings"
	"testing"

	"github.com/sp-night/sp-night/internal/color"
)

func mustLoad(t *testing.T) (*Palette, Roles) {
	t.Helper()
	p, r, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	return p, r
}

// The embedded contract has to load and validate. Every other test in the
// project rests on this one.
func TestEmbeddedContractLoads(t *testing.T) {
	p, r := mustLoad(t)
	if p.Name != "sp_night" {
		t.Errorf("name = %q, want sp_night", p.Name)
	}
	if p.URL != "https://sp-night.github.io" {
		t.Errorf("url = %q, want the public contract URL", p.URL)
	}
	if len(r) == 0 {
		t.Error("no role groups loaded")
	}
}

func TestFlavorsAreDeclarationOrdered(t *testing.T) {
	p, _ := mustLoad(t)
	want := []string{"noite", "garoa", "jaragua"}
	if got := p.FlavorIDs(); !slices.Equal(got, want) {
		t.Errorf("flavor order = %v, want %v", got, want)
	}
	for _, f := range p.Flavors {
		if !f.IsDark() {
			t.Errorf("flavor %q is %q; the theme is dark-only by decision", f.ID, f.Appearance)
		}
	}
}

func TestFlavorLookup(t *testing.T) {
	p, _ := mustLoad(t)
	f, ok := p.Flavor("jaragua")
	if !ok {
		t.Fatal("Flavor(jaragua) not found")
	}
	if f.Label != "Pico do Jaraguá" {
		t.Errorf("label = %q, want Pico do Jaraguá", f.Label)
	}
	if _, ok := p.Flavor("alucard"); ok {
		t.Error("Flavor(alucard) should not be found")
	}
}

func TestEveryFlavorDeclaresTheSame22Colours(t *testing.T) {
	p, _ := mustLoad(t)
	if len(p.Order) != 22 {
		t.Fatalf("palette declares %d colours, want 22", len(p.Order))
	}
	for _, f := range p.Flavors {
		if len(f.Colors) != len(p.Order) {
			t.Errorf("flavor %q declares %d colours, want %d", f.ID, len(f.Colors), len(p.Order))
		}
		for _, k := range p.Order {
			hex, ok := f.Colors[k]
			if !ok {
				t.Errorf("flavor %q is missing %q", f.ID, k)
				continue
			}
			if _, err := color.ParseHex(hex); err != nil {
				t.Errorf("flavor %q colour %q: %v", f.ID, k, err)
			}
		}
	}
}

// Order is grouped: surfaces, then text, then accents, then bright ANSI. The
// generated files, the docs and the previews all walk it, so a reshuffle in
// the JSON would silently reorder every output.
func TestOrderMatchesGroupOrder(t *testing.T) {
	p, _ := mustLoad(t)
	var want []string
	for _, g := range Groups {
		want = append(want, g.Keys...)
	}
	if !slices.Equal(p.Order, want) {
		t.Errorf("palette order = %v\nwant grouped order %v", p.Order, want)
	}
}

func TestEveryRoleResolvesToItsPaletteColour(t *testing.T) {
	p, r := mustLoad(t)
	for _, f := range p.Flavors {
		resolved, err := r.Resolve(f)
		if err != nil {
			t.Fatalf("flavor %q: %v", f.ID, err)
		}
		for group, roles := range r {
			for role, key := range roles {
				want := f.Colors[key]
				if got := resolved[group][role]; got != want {
					t.Errorf("flavor %q: %s.%s = %s, want %s (colour %q)",
						f.ID, group, role, got, want, key)
				}
			}
		}
	}
}

// $comment and $_ are prose in the JSON. If they leaked through as roles, a
// template ranging over .R would emit them as colours.
func TestDocKeysAreNotRoles(t *testing.T) {
	_, r := mustLoad(t)
	for group, roles := range r {
		if strings.HasPrefix(group, "$") {
			t.Errorf("group %q is documentation and should have been dropped", group)
		}
		for role := range roles {
			if strings.HasPrefix(role, "$") {
				t.Errorf("role %s.%s is documentation and should have been dropped", group, role)
			}
		}
	}
}

// The 16 ANSI slots have owners, so sodio and estaiada have no bright twin;
// every other accent does, and a bright is the same colour lifted in
// lightness — not a different hue.
func TestBrightsLiftTheirAccent(t *testing.T) {
	p, _ := mustLoad(t)
	accents := AccentKeys()
	for _, f := range p.Flavors {
		for _, bright := range BrightKeys() {
			base, ok := strings.CutSuffix(bright, "_vivo")
			if !ok {
				t.Errorf("%q is in Bright ANSI but is not named <accent>_vivo", bright)
				continue
			}
			if !slices.Contains(accents, base) {
				t.Errorf("%q lifts %q, which is not an accent", bright, base)
				continue
			}
			lb := color.MustParseHex(f.Colors[base]).Oklab().L
			lv := color.MustParseHex(f.Colors[bright]).Oklab().L
			if lv <= lb {
				t.Errorf("flavor %q: %s (L %.3f) is not lighter than %s (L %.3f)",
					f.ID, bright, lv, base, lb)
			}
		}
	}
}

func TestAnsiCoversAllSixteenSlots(t *testing.T) {
	_, r := mustLoad(t)
	ansi := r["ansi"]
	for _, name := range []string{
		"black", "red", "green", "yellow", "blue", "magenta", "cyan", "white",
		"bright_black", "bright_red", "bright_green", "bright_yellow",
		"bright_blue", "bright_magenta", "bright_cyan", "bright_white",
	} {
		if _, ok := ansi[name]; !ok {
			t.Errorf("ansi.%s is not defined", name)
		}
	}
	if len(ansi) != 16 {
		t.Errorf("ansi has %d roles, want exactly 16", len(ansi))
	}
}

// The two-accent rule: ui.accent is identity (terminals, bars, cursor),
// ui.accent_alt is the system accent for app widgets. An orange system accent
// makes a whole file manager look like a warning.
func TestTheTwoAccentsAreDistinct(t *testing.T) {
	_, r := mustLoad(t)
	ui := r["ui"]
	if ui["accent"] != "sodio" {
		t.Errorf("ui.accent = %q, want sodio", ui["accent"])
	}
	if ui["accent_alt"] != "marginal" {
		t.Errorf("ui.accent_alt = %q, want marginal", ui["accent_alt"])
	}
}

func TestValidateRejectsABrokenPalette(t *testing.T) {
	p, _ := mustLoad(t)

	t.Run("missing colour", func(t *testing.T) {
		broken := *p
		broken.Flavors = slices.Clone(p.Flavors)
		colours := make(map[string]string, len(broken.Flavors[0].Colors))
		for k, v := range broken.Flavors[0].Colors {
			colours[k] = v
		}
		delete(colours, "sodio")
		broken.Flavors[0].Colors = colours
		if err := broken.Validate(); err == nil {
			t.Error("Validate accepted a flavour missing sodio")
		}
	})

	t.Run("role pointing nowhere", func(t *testing.T) {
		broken := Roles{"ui": {"bg": "asfalto"}}
		if err := broken.Validate(p); err == nil {
			t.Error("Validate accepted a role naming a colour that does not exist")
		}
	})

	t.Run("bad hex", func(t *testing.T) {
		f := p.Flavors[0]
		colours := map[string]string{}
		for k, v := range f.Colors {
			colours[k] = v
		}
		colours["laje"] = "not-a-colour"
		broken := *p
		broken.Flavors = []Flavor{{ID: f.ID, Label: f.Label, Appearance: f.Appearance, Colors: colours}}
		if err := broken.Validate(); err == nil {
			t.Error("Validate accepted an invalid hex")
		}
	})
}
