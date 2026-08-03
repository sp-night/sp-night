package render

import (
	"strings"
	"testing"

	"github.com/sp-night/sp-night/internal/theme"
)

const validFrontmatter = `---
spn:
  version: "^1.0"
  matrix: [flavor]
  filename: "themes/sp_night_{{ .Flavor }}"
---
background = {{ nohash .R.ui.bg }}
`

func TestSplitFrontmatter(t *testing.T) {
	fm, body, err := SplitFrontmatter([]byte(validFrontmatter))
	if err != nil {
		t.Fatalf("SplitFrontmatter: %v", err)
	}
	if fm.SPN.Version != "^1.0" {
		t.Errorf("version = %q", fm.SPN.Version)
	}
	if !fm.SPN.PerFlavor() {
		t.Error("matrix [flavor] should render per flavour")
	}
	if fm.SPN.Filename != "themes/sp_night_{{ .Flavor }}" {
		t.Errorf("filename = %q", fm.SPN.Filename)
	}
	// The body must start at the template's first real line. A stray leading
	// newline here would shift every generated file by one byte.
	if want := "background = {{ nohash .R.ui.bg }}\n"; body != want {
		t.Errorf("body = %q, want %q", body, want)
	}
}

func TestSplitFrontmatterRejectsMalformedInput(t *testing.T) {
	for name, src := range map[string]string{
		"no frontmatter":   "background = {{ .R.ui.bg }}\n",
		"unterminated":     "---\nspn:\n  version: \"^1.0\"\n",
		"delimiter inline": "--- spn:\n  version: \"^1.0\"\n---\n",
		"no filename":      "---\nspn:\n  version: \"^1.0\"\n---\nx\n",
		"no version":       "---\nspn:\n  filename: \"a\"\n---\nx\n",
		"unknown axis":     "---\nspn:\n  version: \"^1.0\"\n  matrix: [flavour]\n  filename: \"a\"\n---\nx\n",
		"bad yaml":         "---\nspn:\n  version: [\n---\nx\n",
	} {
		t.Run(name, func(t *testing.T) {
			if _, _, err := SplitFrontmatter([]byte(src)); err == nil {
				t.Error("accepted malformed frontmatter")
			}
		})
	}
}

// A byte order mark must not hide the opening delimiter.
func TestSplitFrontmatterToleratesAByteOrderMark(t *testing.T) {
	if _, _, err := SplitFrontmatter([]byte("\ufeff" + validFrontmatter)); err != nil {
		t.Errorf("BOM broke the parse: %v", err)
	}
}

// An empty matrix renders once — what a target with a single fixed config file
// needs.
func TestEmptyMatrixRendersOnce(t *testing.T) {
	src := "---\nspn:\n  version: \"^1.0\"\n  filename: \"theme.yml\"\n---\nfg: {{ .R.ui.fg }}\n"
	fm, _, err := SplitFrontmatter([]byte(src))
	if err != nil {
		t.Fatalf("SplitFrontmatter: %v", err)
	}
	if fm.SPN.PerFlavor() {
		t.Error("an absent matrix should not iterate flavours")
	}

	pal, roles := mustContract(t)
	tpl, err := Parse("t.tmpl", []byte(src))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	files, err := tpl.Render(pal, roles, Meta{})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("rendered %d file(s), want 1", len(files))
	}
	if files[0].Path != "theme.yml" {
		t.Errorf("path = %q", files[0].Path)
	}
}

func TestCheckVersion(t *testing.T) {
	for _, tc := range []struct {
		constraint, tool string
		ok               bool
	}{
		{"^1.0", "1.0.0", true},
		{"^1.0", "1.4.2", true},
		{"^1.0", "2.0.0", false},
		{"^1.0", "0.9.0", false},
		{"^1.2", "1.1.0", false}, // within the major, but older than required
		{">=1.2", "2.0.0", true},
		{">=1.2", "1.1.9", false},
		{"1.2.3", "1.2.3", true},
		{"1.2.3", "1.2.4", false},
		{"^1.0", "v1.5.0", true},     // a leading v is tolerated
		{"^1.0", "1.5.0-rc.1", true}, // pre-release suffix ignored
		{"^2.0", "dev", true},        // working in this repository
	} {
		err := CheckVersion(tc.constraint, tc.tool)
		if tc.ok && err != nil {
			t.Errorf("CheckVersion(%q, %q) = %v, want ok", tc.constraint, tc.tool, err)
		}
		if !tc.ok && err == nil {
			t.Errorf("CheckVersion(%q, %q) = nil, want a failure", tc.constraint, tc.tool)
		}
	}
}

// The rule the project rests on. A template asking for .C is a finding.
func TestLintCatchesRawPaletteReferences(t *testing.T) {
	for name, body := range map[string]string{
		"field":     "color4 = {{ .C.marginal }}\n",
		"index":     "color4 = {{ index .C \"marginal\" }}\n",
		"dollar":    "{{ range .Order }}{{ index $.C . }}{{ end }}\n",
		"in a pipe": "bg = {{ .C.laje | nohash }}\n",
	} {
		t.Run(name, func(t *testing.T) {
			tpl := mustParse(t, "---\nspn:\n  version: \"^1.0\"\n  filename: \"o\"\n---\n"+body)
			findings := tpl.Lint()
			if len(findings) == 0 {
				t.Fatalf("Lint missed a raw palette reference in %q", body)
			}
			if !strings.Contains(findings[0].Message, "role") {
				t.Errorf("message does not point at roles: %s", findings[0].Message)
			}
		})
	}
}

func TestLintAllowsRolesAndDeclaredVariableLists(t *testing.T) {
	roleOnly := mustParse(t, "---\nspn:\n  version: \"^1.0\"\n  filename: \"o\"\n---\nbg = {{ .R.ui.bg }}\n")
	if f := roleOnly.Lint(); len(f) != 0 {
		t.Errorf("Lint flagged a roles-only template: %v", f)
	}

	// waybar, gtk and hyprland publish the raw palette on purpose, so the user
	// can write @sp_sodio in their own stylesheet.
	declared := mustParse(t, "---\nspn:\n  version: \"^1.0\"\n  filename: \"o\"\n  raw_palette: true\n---\n@define-color sp_sodio {{ .C.sodio }};\n")
	if f := declared.Lint(); len(f) != 0 {
		t.Errorf("Lint flagged a template that declared raw_palette: %v", f)
	}
}

func TestLintReportsLineAndColumn(t *testing.T) {
	tpl := mustParse(t, "---\nspn:\n  version: \"^1.0\"\n  filename: \"o\"\n---\nfirst = {{ .R.ui.bg }}\nsecond = {{ .C.sodio }}\n")
	findings := tpl.Lint()
	if len(findings) != 1 {
		t.Fatalf("got %d finding(s), want 1", len(findings))
	}
	if findings[0].Line != 2 {
		t.Errorf("line = %d, want 2", findings[0].Line)
	}
	if !strings.Contains(findings[0].Excerpt, "second") {
		t.Errorf("excerpt = %q", findings[0].Excerpt)
	}
}

// A mistyped role must fail the build. The default template behaviour is to
// render "<no value>" into the config file and ship it.
func TestARoleTypoFailsTheRender(t *testing.T) {
	pal, roles := mustContract(t)
	tpl := mustParse(t, "---\nspn:\n  version: \"^1.0\"\n  matrix: [flavor]\n  filename: \"o_{{ .Flavor }}\"\n---\nbg = {{ .R.ui.backgruond }}\n")
	if _, err := tpl.Render(pal, roles, Meta{}); err == nil {
		t.Fatal("a mistyped role rendered without error")
	}
}

func TestOutputPathRejectsAbsoluteAndEmpty(t *testing.T) {
	pal, roles := mustContract(t)

	abs := mustParse(t, "---\nspn:\n  version: \"^1.0\"\n  filename: \"/etc/theme\"\n---\nx\n")
	if _, err := abs.Render(pal, roles, Meta{}); err == nil {
		t.Error("an absolute spn.filename was accepted")
	}

	empty := mustParse(t, "---\nspn:\n  version: \"^1.0\"\n  filename: \"{{ .Meaning.nothing }}\"\n---\nx\n")
	if _, err := empty.Render(pal, roles, Meta{}); err == nil {
		t.Error("an spn.filename that renders empty was accepted")
	}
}

// The registry supplies the header fields, so a generated file's install path
// cannot drift from what the website publishes for the same port.
func TestMetaReachesTheTemplate(t *testing.T) {
	pal, roles := mustContract(t)
	tpl := mustParse(t, "---\nspn:\n  version: \"^1.0\"\n  filename: \"o\"\n---\n# {{ .App }} — {{ .Repo }}\n# install: {{ .Install }}\n# enable: {{ .Activate }}\n")
	files, err := tpl.Render(pal, roles, Meta{
		App: "Ghostty", Slug: "ghostty",
		Repo:     "https://github.com/sp-night/ghostty",
		Install:  "~/.config/ghostty/themes/sp_night_noite",
		Activate: "theme = sp_night_noite",
	})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	out := string(files[0].Content)
	for _, want := range []string{"Ghostty", "sp-night/ghostty", "~/.config/ghostty", "theme = sp_night_noite"} {
		if !strings.Contains(out, want) {
			t.Errorf("rendered file does not contain %q:\n%s", want, out)
		}
	}
}

// .All is how a format that keeps a light block alongside the dark one reaches
// the other flavours. The theme is dark-only, so those blocks mirror the dark.
func TestAllReachesTheOtherFlavours(t *testing.T) {
	pal, roles := mustContract(t)
	tpl := mustParse(t, "---\nspn:\n  version: \"^1.0\"\n  matrix: [flavor]\n  filename: \"o_{{ .Flavor }}\"\n---\n{{ (index .All \"garoa\").R.ui.bg }}\n")
	files, err := tpl.Render(pal, roles, Meta{})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	garoa, _ := pal.Flavor("garoa")
	want := garoa.Colors["laje"] + "\n"
	for _, f := range files {
		if string(f.Content) != want {
			t.Errorf("%s = %q, want %q", f.Path, f.Content, want)
		}
	}
}

func TestRenderFlavorPicksOne(t *testing.T) {
	pal, roles := mustContract(t)
	tpl := mustParse(t, "---\nspn:\n  version: \"^1.0\"\n  filename: \"theme.yml\"\n---\n{{ .Flavor }} {{ .R.ui.bg }}\n")
	f, err := tpl.RenderFlavor(pal, roles, Meta{}, "jaragua")
	if err != nil {
		t.Fatalf("RenderFlavor: %v", err)
	}
	if !strings.HasPrefix(string(f.Content), "jaragua ") {
		t.Errorf("content = %q", f.Content)
	}
	if _, err := tpl.RenderFlavor(pal, roles, Meta{}, "alucard"); err == nil {
		t.Error("RenderFlavor accepted an unknown flavour")
	}
}

func TestFuncs(t *testing.T) {
	pal, roles := mustContract(t)
	for _, tc := range []struct{ tmpl, want string }{
		{`{{ nohash .R.ui.bg }}`, "151723"},
		{`{{ rgb .R.ui.bg }}`, "21, 23, 35"},
		{`{{ rgbn .R.ui.bg }}`, "21,23,35"},
		{`{{ rgba 0.5 .R.ui.bg }}`, "rgba(21, 23, 35, 0.50)"},
		{`{{ hexa 0.8 .R.ui.bg }}`, "#151723cc"},
		{`{{ argb 0.8 .R.ui.bg }}`, "cc151723"},
		{`{{ sgrfg .R.ui.fg }}`, "38;2;211;215;235"},
		{`{{ sgrbg .R.ui.bg }}`, "48;2;21;23;35"},
		{`{{ r .R.ui.bg }} {{ g .R.ui.bg }} {{ b .R.ui.bg }}`, "21 23 35"},
		{`{{ upper (nohash .R.ui.bg) }}`, "151723"},
		{`{{ kebab "fg_dim" }}`, "fg-dim"},
		{`{{ pad 6 "ab" }}|`, "ab    |"},
		{`{{ repeat "-" 3 }}`, "---"},
		{`{{ mix 0 .R.ui.bg .R.ui.fg }}`, "#151723"},
		{`{{ mix 1 .R.ui.bg .R.ui.fg }}`, "#d3d7eb"},
		{`{{ lighten 0 .R.ui.bg }}`, "#151723"},
		{`{{ darken 1 .R.ui.fg }}`, "#000000"},
		// on_accent is vao, which is what reads on the sodio accent
		{`{{ readable .C.vao .C.fg .R.ui.accent }}`, "#0f101a"},
	} {
		tpl := mustParse(t, "---\nspn:\n  version: \"^1.0\"\n  filename: \"o\"\n  raw_palette: true\n---\n"+tc.tmpl)
		files, err := tpl.Render(pal, roles, Meta{})
		if err != nil {
			t.Errorf("%s: %v", tc.tmpl, err)
			continue
		}
		if got := string(files[0].Content); got != tc.want {
			t.Errorf("%s = %q, want %q", tc.tmpl, got, tc.want)
		}
	}
}

// contrast and tojson are the two helpers whose output is not a colour.
func TestContrastAndTojsonHelpers(t *testing.T) {
	pal, roles := mustContract(t)
	tpl := mustParse(t, "---\nspn:\n  version: \"^1.0\"\n  filename: \"o\"\n---\n{{ contrast .R.ui.fg .R.ui.bg }}\n{{ tojson .R.diagnostic }}\n")
	files, err := tpl.Render(pal, roles, Meta{})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	out := string(files[0].Content)
	if !strings.HasPrefix(out, "1") { // fg on laje is comfortably above 10:1
		t.Errorf("contrast did not render a ratio: %q", out)
	}
	if !strings.Contains(out, `"error"`) {
		t.Errorf("tojson did not emit the diagnostic group: %q", out)
	}
}

func mustContract(t *testing.T) (*theme.Palette, theme.Roles) {
	t.Helper()
	pal, roles, err := theme.Load()
	if err != nil {
		t.Fatalf("theme.Load: %v", err)
	}
	return pal, roles
}

func mustParse(t *testing.T, src string) *Template {
	t.Helper()
	tpl, err := Parse("test.tmpl", []byte(src))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	return tpl
}
