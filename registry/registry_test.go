package registry

import (
	"strings"
	"testing"
)

func TestEmbeddedCatalogueLoads(t *testing.T) {
	r, err := Embedded()
	if err != nil {
		t.Fatalf("Embedded: %v", err)
	}
	if len(r.Ports) == 0 {
		t.Fatal("the catalogue is empty")
	}
	for _, slug := range []string{"ghostty", "eza"} {
		if _, ok := r.Port(slug); !ok {
			t.Errorf("%q is not in the catalogue", slug)
		}
	}
	// The catalogue lists only shipped ports: there are no placeholders and no
	// planned entries. That cannot be proved offline, so what is asserted is what
	// a placeholder would look like — an entry with nothing to install.
	for _, p := range r.Ports {
		if p.Repo == "" || p.Install == "" {
			t.Errorf("%q looks like a placeholder; a port is listed only once it ships", p.Slug)
		}
	}
	if _, ok := r.Port("nvim"); ok {
		t.Error("nvim is listed but has no repository yet — the catalogue lists only shipped ports")
	}
}

// The website asserts the same thing from the other side. A mismatch would send
// a README's reader to a repository that does not exist.
func TestRepoFollowsFromSlug(t *testing.T) {
	r, err := Embedded()
	if err != nil {
		t.Fatalf("Embedded: %v", err)
	}
	for _, p := range r.Ports {
		if want := "https://github.com/sp-night/" + p.Slug; p.Repo != want {
			t.Errorf("%s: repo = %q, want %q", p.Slug, p.Repo, want)
		}
	}
}

func TestEveryPortCarriesWhatTheReadmeNeeds(t *testing.T) {
	r, err := Embedded()
	if err != nil {
		t.Fatalf("Embedded: %v", err)
	}
	for _, p := range r.Ports {
		if p.InstallGuide == "" {
			t.Errorf("%s: install_guide is empty; the README's Install section would be blank", p.Slug)
		}
		if len(p.Mapping) < 2 {
			t.Errorf("%s: mapping has %d row(s)", p.Slug, len(p.Mapping))
		}
		if len(p.Preview.Body) == 0 {
			t.Errorf("%s: preview has no body", p.Slug)
		}
		// A role reference has to be group.role, or the preview cannot resolve it.
		for i, line := range p.Preview.Body {
			for j, s := range line {
				if s.Role != "" && !strings.Contains(s.Role, ".") {
					t.Errorf("%s: preview line %d span %d role %q is not group.role", p.Slug, i, j, s.Role)
				}
			}
		}
		for _, ref := range p.Preview.Swatches.Roles {
			if !strings.Contains(ref, ".") {
				t.Errorf("%s: swatch role %q is not group.role", p.Slug, ref)
			}
		}
	}
}

func TestGroupsAreDeclared(t *testing.T) {
	r, err := Embedded()
	if err != nil {
		t.Fatalf("Embedded: %v", err)
	}
	for _, p := range r.Ports {
		if _, ok := r.Groups[p.Group]; !ok {
			t.Errorf("%s: group %q is not declared", p.Slug, p.Group)
		}
	}
	if got := r.GroupOrder(); len(got) == 0 {
		t.Error("GroupOrder is empty")
	}
	if n := len(r.ByGroup("terminal")); n == 0 {
		t.Error("no ports in the terminal group")
	}
	if n := len(r.ByGroup("nonexistent")); n != 0 {
		t.Errorf("ByGroup returned %d ports for a group that does not exist", n)
	}
}

func TestValidateRejectsBrokenEntries(t *testing.T) {
	base := `
groups:
  terminal: Terminals
ports:
  - slug: kitty
    name: kitty
    group: terminal
    blurb: A blurb.
    homepage: https://sw.kovidgoyal.net/kitty/
    repo: https://github.com/sp-night/kitty
    install: ~/.config/kitty/sp_night_{flavor}.conf
    template: kitty.conf.tmpl
    mapping:
      - {key: "a", role: "ui.bg", meaning: "m"}
    preview:
      title: t
      swatches: {label: l, keys: [laje]}
      body:
        - - {t: "x", r: ui.fg}
`
	if _, err := Load([]byte(base)); err != nil {
		t.Fatalf("the baseline entry should be valid: %v", err)
	}

	for name, mutate := range map[string]func(string) string{
		"no slug":  func(s string) string { return strings.Replace(s, "slug: kitty", `slug: ""`, 1) },
		"no blurb": func(s string) string { return strings.Replace(s, "    blurb: A blurb.\n", "", 1) },
		"no install": func(s string) string {
			return strings.Replace(s, "    install: ~/.config/kitty/sp_night_{flavor}.conf\n", "", 1)
		},
		"undeclared group": func(s string) string {
			return strings.Replace(s, "    group: terminal", "    group: editor", 1)
		},
		"repo not from slug": func(s string) string {
			return strings.Replace(s, "sp-night/kitty", "rogeradas/kitty", 1)
		},
		"template not tmpl": func(s string) string {
			return strings.Replace(s, "kitty.conf.tmpl", "kitty.conf", 1)
		},
		"no mapping": func(s string) string {
			return strings.Replace(s, "      - {key: \"a\", role: \"ui.bg\", meaning: \"m\"}\n", "", 1)
		},
		"mapping row without role": func(s string) string {
			return strings.Replace(s, `role: "ui.bg"`, `role: ""`, 1)
		},
		"no preview body": func(s string) string {
			return strings.Replace(s, "        - - {t: \"x\", r: ui.fg}\n", "", 1)
		},
		"span with neither r nor c": func(s string) string {
			return strings.Replace(s, `{t: "x", r: ui.fg}`, `{t: "x"}`, 1)
		},
		"span with both r and c": func(s string) string {
			return strings.Replace(s, `{t: "x", r: ui.fg}`, `{t: "x", r: ui.fg, c: laje}`, 1)
		},
		"swatches with both": func(s string) string {
			return strings.Replace(s, `swatches: {label: l, keys: [laje]}`, `swatches: {label: l, keys: [laje], roles: [ui.bg]}`, 1)
		},
		"swatches without label": func(s string) string {
			return strings.Replace(s, `swatches: {label: l, keys: [laje]}`, `swatches: {keys: [laje]}`, 1)
		},
		"no groups": func(s string) string {
			return strings.Replace(s, "groups:\n  terminal: Terminals\n", "groups:\n", 1)
		},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := Load([]byte(mutate(base))); err == nil {
				t.Error("accepted a broken catalogue entry")
			}
		})
	}
}

func TestValidateRejectsDuplicateSlugs(t *testing.T) {
	src := `
groups:
  terminal: Terminals
ports:
  - slug: kitty
    name: kitty
    group: terminal
    blurb: b
    homepage: https://example.com
    repo: https://github.com/sp-night/kitty
    install: p
    template: kitty.tmpl
    mapping: [{key: a, role: ui.bg, meaning: m}]
    preview:
      title: t
      swatches: {label: l, keys: [laje]}
      body: [[{t: x, r: ui.fg}]]
  - slug: kitty
    name: kitty again
    group: terminal
    blurb: b
    homepage: https://example.com
    repo: https://github.com/sp-night/kitty
    install: p
    template: kitty.tmpl
    mapping: [{key: a, role: ui.bg, meaning: m}]
    preview:
      title: t
      swatches: {label: l, keys: [laje]}
      body: [[{t: x, r: ui.fg}]]
`
	if _, err := Load([]byte(src)); err == nil {
		t.Error("accepted two ports with the same slug")
	}
}

func TestEmbeddedCopyLoads(t *testing.T) {
	c, err := EmbeddedCopy()
	if err != nil {
		t.Fatalf("EmbeddedCopy: %v", err)
	}
	// One blurb per flavour, or a README would have a hole where a flavour goes.
	for _, id := range []string{"noite", "garoa", "jaragua"} {
		blurb, ok := c.Blurb(id)
		if !ok || strings.TrimSpace(blurb) == "" {
			t.Errorf("no English blurb for flavour %q", id)
		}
	}
	if _, ok := c.Blurb("alucard"); ok {
		t.Error("Blurb invented a flavour")
	}
	// The claim the whole project rests on has to be in the shared prose, so
	// every port states it identically.
	if !strings.Contains(c.Provenance, "picked by hand") {
		t.Error("the provenance paragraph no longer makes the central claim")
	}
}

func TestLoadCopyRejectsMissingFields(t *testing.T) {
	if _, err := LoadCopy([]byte("logo_alt: a\n")); err == nil {
		t.Error("accepted prose with no tagline")
	}
	if _, err := LoadCopy([]byte("[")); err == nil {
		t.Error("accepted malformed YAML")
	}
}
