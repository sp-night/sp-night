// Package registry is the catalogue of SP Night ports.
//
// It answers the questions that are not about colour: what the app is called,
// where its theme file goes, what line switches it on, which key of its config
// means which role. Every one of those facts used to be written by hand in
// three places — the template header, the port's README, and the website's port
// list — and they drifted the moment one was edited.
//
// Nothing here describes how a theme is produced. The registry describes what a
// user does with a port, so it stays correct no matter what builds the files.
package registry

import (
	"fmt"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

// Registry is the whole catalogue.
type Registry struct {
	Groups map[string]string `yaml:"groups"`
	Ports  []Port            `yaml:"ports"`
}

// Port is one target.
type Port struct {
	// --- identity, shared with the website's port list ---

	Slug     string `yaml:"slug"`     // repository name under the org
	Name     string `yaml:"name"`     // display name
	Group    string `yaml:"group"`    // a key of Registry.Groups
	Blurb    string `yaml:"blurb"`    // one line on what the port covers
	Homepage string `yaml:"homepage"` // the themed app's own site
	Repo     string `yaml:"repo"`     // https://github.com/sp-night/<slug>

	Install  string `yaml:"install"`  // where the file goes
	Activate string `yaml:"activate"` // the line the user adds, if any
	Note     string `yaml:"note"`     // caveat worth surfacing

	// --- what the engine needs on top ---

	Template string `yaml:"template"` // the mapping file, e.g. ghostty.tmpl

	// InstallGuide is the app-specific prose of the README's Install section.
	// It cannot be derived: Ghostty has to explain that its theme files have no
	// extension, eza has to explain that a set EZA_COLORS overrides the file.
	InstallGuide string `yaml:"install_guide"`

	// Mapping is the "What gets themed" table: the app's own key, the role it
	// gets, and the gloss. This is the table a reader checks before trusting a
	// port, and it is the one thing that must never disagree with the template.
	Mapping []MappingRow `yaml:"mapping"`

	// KeyLabel is the first column's header. Defaults to "<Name> key", but eza's
	// table is about theme.yml keys and says so — the header is the first thing a
	// reader uses to locate their own config in the table.
	KeyLabel string `yaml:"key_label"`

	// MappingScope qualifies how complete the mapping is, e.g. "across all ~80
	// keys eza exposes". Empty for a port whose table is already the whole story.
	MappingScope string `yaml:"mapping_scope"`

	// MappingCloser completes the sentence "…without anyone re-deciding ___".
	// Each port names its own most telling key, which is what stops the closing
	// paragraph from reading like boilerplate.
	MappingCloser string `yaml:"mapping_closer"`

	// Preview describes the synthetic screenshot. Synthetic on purpose: drawn
	// from the palette, so it cannot drift from what the user installs.
	Preview Preview `yaml:"preview"`
}

// MappingRow is one line of the README's role table.
type MappingRow struct {
	Key     string `yaml:"key"`
	Role    string `yaml:"role"`
	Meaning string `yaml:"meaning"`
}

// Preview is the synthetic terminal mockup for one port.
type Preview struct {
	// Title is the window title. {flavor} and {label} are substituted.
	Title string `yaml:"title"`

	// Body is the fake session: a list of lines, each a list of coloured runs.
	// An empty line is a vertical gap rather than a blank row.
	Body [][]Span `yaml:"body"`

	Swatches Swatches `yaml:"swatches"`
}

// Span is one coloured run of text.
type Span struct {
	Text string `yaml:"t"`
	Role string `yaml:"r"` // "group.role" from the role layer
	Key  string `yaml:"c"` // or a raw palette key, for a strip that shows colours
	Bold bool   `yaml:"b"`
}

// Swatches is the colour strip along the bottom.
type Swatches struct {
	Label string   `yaml:"label"`
	Roles []string `yaml:"roles"` // "group.role" each
	Keys  []string `yaml:"keys"`  // or raw palette keys
}

// Load parses a registry and validates it.
func Load(data []byte) (*Registry, error) {
	var r Registry
	if err := yaml.Unmarshal(data, &r); err != nil {
		return nil, fmt.Errorf("registry: %w", err)
	}
	if err := r.Validate(); err != nil {
		return nil, err
	}
	return &r, nil
}

// Embedded loads the registry that shipped with this binary.
func Embedded() (*Registry, error) { return Load(portsYAML) }

// Port looks a port up by slug.
func (r *Registry) Port(slug string) (Port, bool) {
	for _, p := range r.Ports {
		if p.Slug == slug {
			return p, true
		}
	}
	return Port{}, false
}

// Slugs lists every port slug, in registry order.
func (r *Registry) Slugs() []string {
	out := make([]string, 0, len(r.Ports))
	for _, p := range r.Ports {
		out = append(out, p.Slug)
	}
	return out
}

// ByGroup returns the ports of a group, in registry order.
func (r *Registry) ByGroup(group string) []Port {
	var out []Port
	for _, p := range r.Ports {
		if p.Group == group {
			out = append(out, p)
		}
	}
	return out
}

// Validate fails on an entry that would produce a broken port. The website's
// test suite asserts the same shape from the other side; this is the engine's
// half, so a bad entry cannot reach a generated README.
func (r *Registry) Validate() error {
	if len(r.Groups) == 0 {
		return fmt.Errorf("registry: no groups defined")
	}

	seen := map[string]bool{}
	for i, p := range r.Ports {
		where := fmt.Sprintf("registry: port %d", i)
		if p.Slug != "" {
			where = fmt.Sprintf("registry: port %q", p.Slug)
		}

		for _, f := range []struct{ name, value string }{
			{"slug", p.Slug}, {"name", p.Name}, {"group", p.Group},
			{"blurb", p.Blurb}, {"homepage", p.Homepage}, {"repo", p.Repo},
			{"install", p.Install}, {"template", p.Template},
			// The website has always refused an entry without it — the port
			// page's Install section is generated from it, so a missing one
			// means a page telling nobody how to install the thing. Requiring
			// it only there meant `spn registry` passed an entry that then
			// failed the site's suite inside the fan-out, where a failure opens
			// no pull request. The two gates now refuse the same catalogue.
			{"install_guide", p.InstallGuide},
		} {
			if strings.TrimSpace(f.value) == "" {
				return fmt.Errorf("%s: %s is required", where, f.name)
			}
		}

		if seen[p.Slug] {
			return fmt.Errorf("%s: duplicate slug", where)
		}
		seen[p.Slug] = true

		if _, ok := r.Groups[p.Group]; !ok {
			return fmt.Errorf("%s: group %q is not defined in groups", where, p.Group)
		}

		// The repo URL is derived from the slug, not written independently.
		// The website asserts exactly this, and a mismatch would send readers
		// of a README to a repository that does not exist.
		if want := "https://github.com/sp-night/" + p.Slug; p.Repo != want {
			return fmt.Errorf("%s: repo is %q, want %q", where, p.Repo, want)
		}

		if !strings.HasSuffix(p.Template, ".tmpl") {
			return fmt.Errorf("%s: template %q must end in .tmpl", where, p.Template)
		}
		if len(p.Mapping) == 0 {
			return fmt.Errorf("%s: mapping is required — it is the table a reader checks before trusting a port", where)
		}
		for j, row := range p.Mapping {
			if strings.TrimSpace(row.Key) == "" || strings.TrimSpace(row.Role) == "" {
				return fmt.Errorf("%s: mapping row %d needs both a key and a role", where, j)
			}
		}
		if err := p.Preview.validate(where); err != nil {
			return err
		}
	}
	return nil
}

func (p Preview) validate(where string) error {
	if strings.TrimSpace(p.Title) == "" {
		return fmt.Errorf("%s: preview.title is required", where)
	}
	if len(p.Body) == 0 {
		return fmt.Errorf("%s: preview.body is required", where)
	}
	for i, line := range p.Body {
		for j, s := range line {
			if s.Role == "" && s.Key == "" {
				return fmt.Errorf("%s: preview.body line %d span %d has neither a role (r) nor a palette key (c)", where, i, j)
			}
			if s.Role != "" && s.Key != "" {
				return fmt.Errorf("%s: preview.body line %d span %d sets both r and c; pick one", where, i, j)
			}
		}
	}
	if len(p.Swatches.Roles) == 0 && len(p.Swatches.Keys) == 0 {
		return fmt.Errorf("%s: preview.swatches needs roles or keys", where)
	}
	if len(p.Swatches.Roles) > 0 && len(p.Swatches.Keys) > 0 {
		return fmt.Errorf("%s: preview.swatches sets both roles and keys; pick one", where)
	}
	if strings.TrimSpace(p.Swatches.Label) == "" {
		return fmt.Errorf("%s: preview.swatches.label is required", where)
	}
	return nil
}

// GroupOrder returns the group keys in a stable order: the order they are
// referenced by the ports, so the README and the website agree.
func (r *Registry) GroupOrder() []string {
	var out []string
	for _, p := range r.Ports {
		if !slices.Contains(out, p.Group) {
			out = append(out, p.Group)
		}
	}
	return out
}
