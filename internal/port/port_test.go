package port

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sp-night/sp-night/internal/render"
	"github.com/sp-night/sp-night/internal/theme"
	"github.com/sp-night/sp-night/registry"
)

func fixtures(t *testing.T) (*theme.Palette, theme.Roles, *registry.Registry, *registry.Copy) {
	t.Helper()
	pal, roles, err := theme.Load()
	if err != nil {
		t.Fatalf("theme.Load: %v", err)
	}
	reg, err := registry.Embedded()
	if err != nil {
		t.Fatalf("registry.Embedded: %v", err)
	}
	cp, err := registry.EmbeddedCopy()
	if err != nil {
		t.Fatalf("registry.EmbeddedCopy: %v", err)
	}
	return pal, roles, reg, cp
}

// A preview that is not well-formed XML renders as nothing on GitHub, which is
// the one place it has to work.
func TestPreviewIsWellFormedXML(t *testing.T) {
	pal, roles, reg, _ := fixtures(t)
	for _, p := range reg.Ports {
		for _, fl := range pal.Flavors {
			svg, err := SVG(p, pal, roles, fl)
			if err != nil {
				t.Fatalf("%s/%s: %v", p.Slug, fl.ID, err)
			}
			if err := xml.Unmarshal(svg, new(any)); err != nil {
				t.Errorf("%s/%s is not well-formed XML: %v", p.Slug, fl.ID, err)
			}
		}
	}
}

// Synthetic means drawn from the palette. Every colour in the file has to be one
// of this flavour's 22, or the preview is showing something the user will not get.
func TestPreviewOnlyUsesThisFlavoursColours(t *testing.T) {
	pal, roles, reg, _ := fixtures(t)
	for _, p := range reg.Ports {
		for _, fl := range pal.Flavors {
			svg, err := SVG(p, pal, roles, fl)
			if err != nil {
				t.Fatalf("%s/%s: %v", p.Slug, fl.ID, err)
			}
			allowed := map[string]bool{}
			for _, hex := range fl.Colors {
				allowed[hex] = true
			}
			for _, found := range hexPattern(string(svg)) {
				if !allowed[found] {
					t.Errorf("%s/%s uses %s, which is not in the %s palette", p.Slug, fl.ID, found, fl.ID)
				}
			}
		}
	}
}

// {flavor} and {label} must be substituted everywhere, including inside the
// fake session — a preview showing a literal "{flavor}" is worse than no preview.
func TestPreviewSubstitutesPlaceholders(t *testing.T) {
	pal, roles, reg, _ := fixtures(t)
	for _, p := range reg.Ports {
		for _, fl := range pal.Flavors {
			svg, err := SVG(p, pal, roles, fl)
			if err != nil {
				t.Fatalf("%s/%s: %v", p.Slug, fl.ID, err)
			}
			for _, ph := range []string{"{flavor}", "{label}"} {
				if strings.Contains(string(svg), ph) {
					t.Errorf("%s/%s still contains %s", p.Slug, fl.ID, ph)
				}
			}
			if !strings.Contains(string(svg), fl.Label) {
				t.Errorf("%s/%s does not name the flavour in its aria-label", p.Slug, fl.ID)
			}
		}
	}
}

func TestPreviewFailsOnAnUnknownRole(t *testing.T) {
	pal, roles, reg, _ := fixtures(t)
	p, _ := reg.Port("ghostty")
	p.Preview.Body = [][]registry.Span{{{Text: "x", Role: "ui.backgruond"}}}
	if _, err := SVG(p, pal, roles, pal.Flavors[0]); err == nil {
		t.Error("a mistyped role rendered a preview instead of failing")
	}

	p2, _ := reg.Port("ghostty")
	p2.Preview.Swatches = registry.Swatches{Label: "l", Keys: []string{"asfalto"}}
	if _, err := SVG(p2, pal, roles, pal.Flavors[0]); err == nil {
		t.Error("a nonexistent palette key rendered a swatch instead of failing")
	}
}

// f1 has to round halves away from zero: Go's own %.1f would write 125.2 where
// the published previews write 125.3, and every swatch would shift.
func TestF1RoundsHalvesAwayFromZero(t *testing.T) {
	for _, tc := range []struct {
		in   float64
		want string
	}{
		{125.25, "125.3"},
		{75.625, "75.6"},
		{770.375, "770.4"},
		{43.625, "43.6"},
		{26, "26.0"},
		{11.5, "11.5"},
		{42.5, "42.5"},
	} {
		if got := f1(tc.in); got != tc.want {
			t.Errorf("f1(%v) = %s, want %s", tc.in, got, tc.want)
		}
	}
}

// A quote is ordinary text in XML character data. Escaping it would turn the
// Lua string in the Ghostty preview into &#34; noise.
func TestEscapingLeavesQuotesAloneInText(t *testing.T) {
	if got := escText(`a "b" & c < d`); got != `a "b" &amp; c &lt; d` {
		t.Errorf("escText = %q", got)
	}
	if got := escAttr(`a "b" & c`); got != `a &quot;b&quot; &amp; c` {
		t.Errorf("escAttr = %q", got)
	}
}

func TestReadmeIsDerivedFromTheCatalogue(t *testing.T) {
	pal, roles, reg, cp := fixtures(t)

	for _, slug := range []string{"ghostty", "eza"} {
		p, ok := reg.Port(slug)
		if !ok {
			t.Fatalf("%s missing from the catalogue", slug)
		}
		src, err := os.ReadFile(filepath.Join("..", "render", "testdata", slug, p.Template))
		if err != nil {
			t.Fatalf("read %s: %v", p.Template, err)
		}
		tpl, err := render.Parse(p.Template, src)
		if err != nil {
			t.Fatalf("parse %s: %v", p.Template, err)
		}

		out, err := README(p, cp, pal, roles, tpl)
		if err != nil {
			t.Fatalf("%s: %v", slug, err)
		}
		md := string(out)

		// The role table is the thing a reader checks before trusting a port.
		for _, row := range p.Mapping {
			if !strings.Contains(md, row.Key) {
				t.Errorf("%s: README omits mapping key %s", slug, row.Key)
			}
			if !strings.Contains(md, row.Role) {
				t.Errorf("%s: README omits role %s", slug, row.Role)
			}
		}

		// Every flavour, named with the file it actually ships as.
		for _, fl := range pal.Flavors {
			if !strings.Contains(md, fl.Label) {
				t.Errorf("%s: README omits flavour %s", slug, fl.Label)
			}
			if !strings.Contains(md, "assets/preview-"+fl.ID+".svg") {
				t.Errorf("%s: README does not link the %s preview", slug, fl.ID)
			}
		}

		for _, want := range []string{
			p.Homepage, p.Name, p.Template,
			"https://sp-night.github.io/spec",
			"picked by hand",
			"[MIT](LICENSE)",
		} {
			if !strings.Contains(md, want) {
				t.Errorf("%s: README omits %q", slug, want)
			}
		}

		// No placeholder may survive into a published README.
		for _, ph := range []string{"{closer}", "{template}", "{app}", "{keylabel}", "{scope}", "<no value>"} {
			if strings.Contains(md, ph) {
				t.Errorf("%s: README still contains %s", slug, ph)
			}
		}
	}
}

// The generated header must not name the tool: the palette is the public
// contract, and the engine is a replaceable detail.
func TestReadmeDoesNotNameTheTool(t *testing.T) {
	pal, roles, reg, cp := fixtures(t)
	p, _ := reg.Port("ghostty")
	src, err := os.ReadFile(filepath.Join("..", "render", "testdata", "ghostty", p.Template))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	tpl, err := render.Parse(p.Template, src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	out, err := README(p, cp, pal, roles, tpl)
	if err != nil {
		t.Fatalf("README: %v", err)
	}
	if strings.Contains(string(out), "spn gen") {
		t.Error("the README tells the reader to run the engine; the shipped files are final")
	}
}

func TestScaffoldProducesAWorkingStub(t *testing.T) {
	_, _, reg, _ := fixtures(t)
	p, _ := reg.Port("ghostty")

	files, err := Scaffold(p)
	if err != nil {
		t.Fatalf("Scaffold: %v", err)
	}

	byPath := map[string]string{}
	for _, f := range files {
		byPath[f.Path] = string(f.Content)
	}
	for _, want := range []string{p.Template, ".github/workflows/theme.yml", "renovate.json"} {
		if _, ok := byPath[want]; !ok {
			t.Errorf("scaffold is missing %s", want)
		}
	}

	// The stub has to be a valid mapping straight away, or `spn new` hands over
	// something that does not build.
	stub, err := render.Parse(p.Template, []byte(byPath[p.Template]))
	if err != nil {
		t.Fatalf("the scaffolded stub does not parse: %v", err)
	}
	if stub.Spec.Version != ScaffoldVersion {
		t.Errorf("stub version = %q, want %q", stub.Spec.Version, ScaffoldVersion)
	}
	if !stub.Spec.PerFlavor() {
		t.Error("the stub should render one file per flavour")
	}
	// It must pass its own lint: the commented counter-example is inside a
	// template action, so it is not a real .C reference.
	if f := stub.Lint(); len(f) != 0 {
		t.Errorf("the scaffolded stub fails lint: %v", f)
	}

	// The workflow calls the engine's reusable workflow, not a copy of it.
	wf := byPath[".github/workflows/theme.yml"]
	if !strings.Contains(wf, "sp-night/sp-night/.github/workflows/spn-check.yml@v1") {
		t.Errorf("the scaffolded workflow does not use the reusable workflow:\n%s", wf)
	}
	if !strings.Contains(wf, "port: "+p.Slug) {
		t.Error("the scaffolded workflow does not name its port")
	}
}

// Ghostty resolves by exact file name and wants no extension; eza wants .yml.
func TestConfigExtensionFollowsTheInstallPath(t *testing.T) {
	for install, want := range map[string]string{
		"~/.config/ghostty/themes/sp_night_{flavor}": "",
		"~/.config/eza/theme.yml":                    ".yml",
		"~/.config/kitty/sp_night_{flavor}.conf":     ".conf",
		"~/.config/alacritty/sp_night.toml":          ".toml",
	} {
		if got := configExtension(install); got != want {
			t.Errorf("configExtension(%q) = %q, want %q", install, got, want)
		}
	}
}

// hexPattern pulls every #rrggbb out of a string.
func hexPattern(s string) []string {
	var out []string
	for i := 0; i+7 <= len(s); i++ {
		if s[i] != '#' {
			continue
		}
		candidate := s[i : i+7]
		if isHex(candidate[1:]) {
			out = append(out, candidate)
		}
	}
	return out
}

func isHex(s string) bool {
	for i := range len(s) {
		c := s[i]
		switch {
		case c >= '0' && c <= '9', c >= 'a' && c <= 'f', c >= 'A' && c <= 'F':
		default:
			return false
		}
	}
	return true
}
