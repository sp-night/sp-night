// Package theme loads and validates the SP Night contract: the palette, the
// role layer, and the resolution of one against the other.
//
// Three layers, one direction:
//
//	palette   23 named colours per flavour — the only place a hex is written
//	roles     every role names a palette key, never a colour
//	ports     a template asks for a role, never a colour
//
// Nothing here knows what a port looks like. It answers one question:
// "for this flavour, what colour is this role?"
package theme

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/sp-night/sp-night/internal/color"
	"github.com/sp-night/sp-night/palette"
)

// Group is a semantic band of the palette. The bands are not decoration: the
// audits need to know which keys are accents (they get measured against each
// other) and which are bright ANSI (they get measured against the surfaces).
//
// They live in sp_night.json rather than in this file because the website
// vendors the contract and needs the same partition to publish it. Declared
// here, the site had to re-type it, and a colour the contract declared but the
// site's copy did not know about vanished from the published palette without a
// single test failing.
type Group struct {
	// ID is the stable handle the code and the website address a band by;
	// Label is what a reader sees.
	ID    string
	Label string
	Keys  []string
}

// The two bands the audit addresses by name: accents get measured against each
// other, brights against the surfaces. Everything else about a group is data.
const (
	GroupAccents = "accents"
	GroupVivo    = "vivo"
)

// Group looks a band up by id.
func (p *Palette) Group(id string) (Group, bool) {
	for _, g := range p.Groups {
		if g.ID == id {
			return g, true
		}
	}
	return Group{}, false
}

// GroupKeys returns the palette keys of a band, panicking on an id that does
// not exist — Validate has already proved the canonical ones are present, so
// this is a typo in a caller rather than a runtime condition.
func (p *Palette) GroupKeys(id string) []string {
	g, ok := p.Group(id)
	if !ok {
		panic("theme: no palette group with id " + id)
	}
	return g.Keys
}

// AccentKeys and BrightKeys read the partition, so moving a colour between
// bands never requires editing a second list in the audit.
func (p *Palette) AccentKeys() []string { return p.GroupKeys(GroupAccents) }
func (p *Palette) BrightKeys() []string { return p.GroupKeys(GroupVivo) }

// Flavor is one way of looking at the city.
type Flavor struct {
	ID          string
	Label       string
	Appearance  string
	Description string
	Colors      map[string]string
}

// IsDark reports whether the flavour is dark. All three are, by decision.
func (f Flavor) IsDark() bool { return f.Appearance == "dark" }

// Palette is sp_night.json. Flavors is ordered as declared in the file, which
// is why adding a flavour is a JSON edit and nothing else.
type Palette struct {
	Name        string
	Label       string
	Description string
	Author      string
	URL         string
	Meaning     map[string]string

	// Order is the palette keys in declaration order. Go maps have none, and
	// every output — files, docs, previews — has to be stable.
	Order []string

	// Groups partitions Order into semantic bands, in declaration order.
	Groups  []Group
	Flavors []Flavor
}

// Flavor looks a flavour up by id.
func (p *Palette) Flavor(id string) (Flavor, bool) {
	for _, f := range p.Flavors {
		if f.ID == id {
			return f, true
		}
	}
	return Flavor{}, false
}

// FlavorIDs returns the flavour ids in declaration order.
func (p *Palette) FlavorIDs() []string {
	ids := make([]string, 0, len(p.Flavors))
	for _, f := range p.Flavors {
		ids = append(ids, f.ID)
	}
	return ids
}

// Roles is the semantic layer: group -> role -> palette key.
type Roles map[string]map[string]string

// Resolved is a role layer with the keys already turned into hex for one
// flavour: group -> role -> "#rrggbb". This is what a template sees as .R.
type Resolved map[string]map[string]string

// wire types mirror the JSON exactly; the exported types above are what the
// rest of the tool works with.
type wireFlavor struct {
	Label       string            `json:"label"`
	Appearance  string            `json:"appearance"`
	Description string            `json:"description"`
	Colors      map[string]string `json:"colors"`
}

type wireGroup struct {
	Label string   `json:"label"`
	Keys  []string `json:"keys"`
}

type wirePalette struct {
	Name        string                `json:"name"`
	Label       string                `json:"label"`
	Description string                `json:"description"`
	Author      string                `json:"author"`
	URL         string                `json:"url"`
	Meaning     map[string]string     `json:"meaning"`
	Groups      map[string]wireGroup  `json:"groups"`
	Flavors     map[string]wireFlavor `json:"flavors"`
	rawFlavors  json.RawMessage
}

// Load reads the palette and roles embedded in the binary. This is the normal
// path: a port never carries palette data, the tool brings it.
func Load() (*Palette, Roles, error) {
	return load(palette.SpNight, palette.Roles)
}

// LoadDir reads palette/sp_night.json and palette/roles.json from a directory
// instead of the embedded copies. Used while retuning the palette in this
// repo, so `spn check` measures what is on disk rather than what shipped.
func LoadDir(dir string) (*Palette, Roles, error) {
	pb, err := os.ReadFile(filepath.Join(dir, "sp_night.json"))
	if err != nil {
		return nil, nil, err
	}
	rb, err := os.ReadFile(filepath.Join(dir, "roles.json"))
	if err != nil {
		return nil, nil, err
	}
	return load(pb, rb)
}

func load(paletteJSON, rolesJSON []byte) (*Palette, Roles, error) {
	p, err := parsePalette(paletteJSON)
	if err != nil {
		return nil, nil, err
	}
	r, err := parseRoles(rolesJSON)
	if err != nil {
		return nil, nil, err
	}
	if err := p.Validate(); err != nil {
		return nil, nil, err
	}
	if err := r.Validate(p); err != nil {
		return nil, nil, err
	}
	return p, r, nil
}

func parsePalette(data []byte) (*Palette, error) {
	// Decoded twice on purpose: once into maps for the values, once as raw
	// JSON for the key order. A JSON object is ordered on disk and unordered
	// in a Go map, and the order is the contract for every generated file.
	var w wirePalette
	if err := json.Unmarshal(data, &w); err != nil {
		return nil, fmt.Errorf("sp_night.json: %w", err)
	}
	var probe struct {
		Groups  json.RawMessage `json:"groups"`
		Flavors json.RawMessage `json:"flavors"`
	}
	if err := json.Unmarshal(data, &probe); err != nil {
		return nil, fmt.Errorf("sp_night.json: %w", err)
	}

	flavorIDs, err := objectKeys(probe.Flavors)
	if err != nil {
		return nil, fmt.Errorf("sp_night.json: flavors: %w", err)
	}

	p := &Palette{
		Name: w.Name, Label: w.Label, Description: w.Description,
		Author: w.Author, URL: w.URL, Meaning: w.Meaning,
	}

	// Band order is the declaration order too: it is how the palette reads in
	// the file, in `spn palette`, and on the website.
	groupIDs, err := objectKeys(probe.Groups)
	if err != nil {
		return nil, fmt.Errorf("sp_night.json: groups: %w", err)
	}
	for _, id := range groupIDs {
		wg := w.Groups[id]
		p.Groups = append(p.Groups, Group{ID: id, Label: wg.Label, Keys: wg.Keys})
	}
	for _, id := range flavorIDs {
		wf := w.Flavors[id]
		p.Flavors = append(p.Flavors, Flavor{
			ID: id, Label: wf.Label, Appearance: wf.Appearance,
			Description: wf.Description, Colors: wf.Colors,
		})
	}

	// Colour order comes from the first flavour; Validate proves the rest
	// declare exactly the same keys.
	if len(flavorIDs) > 0 {
		var byID map[string]json.RawMessage
		if err := json.Unmarshal(probe.Flavors, &byID); err != nil {
			return nil, err
		}
		var colorsProbe struct {
			Colors json.RawMessage `json:"colors"`
		}
		if err := json.Unmarshal(byID[flavorIDs[0]], &colorsProbe); err != nil {
			return nil, err
		}
		if p.Order, err = objectKeys(colorsProbe.Colors); err != nil {
			return nil, fmt.Errorf("sp_night.json: flavor %q: colors: %w", flavorIDs[0], err)
		}
	}
	return p, nil
}

func parseRoles(data []byte) (Roles, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("roles.json: %w", err)
	}
	r := Roles{}
	for group, msg := range raw {
		if isDoc(group) {
			continue // $comment and friends are prose, not groups
		}
		var m map[string]string
		if err := json.Unmarshal(msg, &m); err != nil {
			return nil, fmt.Errorf("roles.json: group %q: %w", group, err)
		}
		for k := range m {
			if isDoc(k) {
				delete(m, k) // $_ inside a group is prose too
			}
		}
		r[group] = m
	}
	return r, nil
}

// isDoc reports whether a JSON key is inline documentation rather than data.
// The convention is shared with the website's resolver: a leading $.
func isDoc(key string) bool { return len(key) > 0 && key[0] == '$' }

// objectKeys returns the keys of a JSON object in the order they appear.
func objectKeys(obj json.RawMessage) ([]string, error) {
	dec := json.NewDecoder(bytes.NewReader(obj))
	t, err := dec.Token()
	if err != nil {
		return nil, err
	}
	if d, ok := t.(json.Delim); !ok || d != '{' {
		return nil, fmt.Errorf("want an object, got %v", t)
	}
	var keys []string
	for dec.More() {
		kt, err := dec.Token()
		if err != nil {
			return nil, err
		}
		k, ok := kt.(string)
		if !ok {
			return nil, fmt.Errorf("want a string key, got %v", kt)
		}
		var skip json.RawMessage
		if err := dec.Decode(&skip); err != nil {
			return nil, err
		}
		keys = append(keys, k)
	}
	return keys, nil
}

// Validate proves every flavour declares exactly the canonical keys with a
// valid hex, and that Groups partitions those keys. An incomplete flavour
// otherwise breaks silently in one obscure target; here it breaks now.
func (p *Palette) Validate() error {
	if len(p.Flavors) == 0 {
		return fmt.Errorf("sp_night.json: no flavors defined")
	}
	if len(p.Order) == 0 {
		return fmt.Errorf("sp_night.json: no colours defined")
	}

	// The bands the audit addresses by name have to be there, or it would
	// measure an empty list and report a clean palette it never looked at.
	for _, id := range []string{GroupAccents, GroupVivo} {
		if _, ok := p.Group(id); !ok {
			return fmt.Errorf("sp_night.json: groups declares no %q band, which the audit measures", id)
		}
	}
	for _, g := range p.Groups {
		if g.Label == "" {
			return fmt.Errorf("sp_night.json: group %q has no label", g.ID)
		}
		if len(g.Keys) == 0 {
			return fmt.Errorf("sp_night.json: group %q lists no colours", g.ID)
		}
	}

	// The bands must cover the declared colours exactly — no colour outside a
	// band, no band naming a colour that no longer exists, nothing listed
	// twice. This is the check that would have caught a colour added to the
	// palette and forgotten in the partition, which is how it disappears from
	// everything that walks the groups.
	var grouped []string
	for _, g := range p.Groups {
		for _, k := range g.Keys {
			if slices.Contains(grouped, k) {
				return fmt.Errorf("sp_night.json: colour %q is in more than one group", k)
			}
			grouped = append(grouped, k)
		}
	}
	for _, k := range p.Order {
		if !slices.Contains(grouped, k) {
			return fmt.Errorf("sp_night.json: colour %q is in no group", k)
		}
	}
	for _, k := range grouped {
		if !slices.Contains(p.Order, k) {
			return fmt.Errorf("sp_night.json: group names %q, which no flavour declares", k)
		}
	}

	for _, f := range p.Flavors {
		if f.Appearance != "dark" && f.Appearance != "light" {
			return fmt.Errorf("flavor %q: appearance must be dark or light, got %q", f.ID, f.Appearance)
		}
		if f.Label == "" {
			return fmt.Errorf("flavor %q: label is required", f.ID)
		}
		for _, k := range p.Order {
			hex, ok := f.Colors[k]
			if !ok {
				return fmt.Errorf("flavor %q: missing colour %q", f.ID, k)
			}
			if _, err := color.ParseHex(hex); err != nil {
				return fmt.Errorf("flavor %q, colour %q: %w", f.ID, k, err)
			}
		}
		for k := range f.Colors {
			if !slices.Contains(p.Order, k) {
				return fmt.Errorf("flavor %q: unknown colour %q", f.ID, k)
			}
		}
	}
	return nil
}

// Validate proves every role points at a colour that exists. A role naming a
// deleted colour is the one failure that would reach every port at once.
func (r Roles) Validate(p *Palette) error {
	if len(r) == 0 {
		return fmt.Errorf("roles.json: no role groups defined")
	}
	for _, f := range p.Flavors {
		if _, err := r.Resolve(f); err != nil {
			return err
		}
	}
	return nil
}

// Resolve turns roles into hex for one flavour.
func (r Roles) Resolve(f Flavor) (Resolved, error) {
	out := Resolved{}
	for group, roles := range r {
		g := map[string]string{}
		for role, key := range roles {
			hex, ok := f.Colors[key]
			if !ok {
				return nil, fmt.Errorf("role %s.%s points at colour %q, which flavor %q does not define",
					group, role, key, f.ID)
			}
			g[role] = hex
		}
		out[group] = g
	}
	return out, nil
}
