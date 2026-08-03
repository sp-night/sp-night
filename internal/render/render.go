package render

import (
	"bytes"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/sp-night/sp-night/internal/theme"
)

// Meta is the per-port information a template needs in its header but must not
// hard-code: where the file installs, what line enables it, which repository
// it lives in. It comes from the registry, so the header of a generated file
// cannot drift from what the website says about the same port.
type Meta struct {
	App      string // display name, e.g. "Ghostty"
	Slug     string // repository name under the org, e.g. "ghostty"
	Repo     string // https://github.com/sp-night/<slug>
	Homepage string // the themed app's own site
	Install  string // where the file goes on the user's machine
	Activate string // the line the user adds, if the app needs one
}

// forFlavor substitutes {flavor} and {label} in the catalogue's per-port strings.
//
// The catalogue states an install path once, with a placeholder — a port installs
// to sp_night_{flavor}, not to three separately written paths. Rendering has to
// resolve it, or a generated header would tell the reader to copy a file called
// "sp_night_{flavor}".
func (m Meta) forFlavor(f theme.Flavor) Meta {
	sub := func(s string) string {
		s = strings.ReplaceAll(s, "{flavor}", f.ID)
		return strings.ReplaceAll(s, "{label}", f.Label)
	}
	m.Install = sub(m.Install)
	m.Activate = sub(m.Activate)
	return m
}

// FlavorRef is one flavour as seen by another. It carries no All of its own, so
// a template cannot walk into a cycle.
type FlavorRef struct {
	Flavor      string
	Label       string
	Appearance  string
	Description string
	IsDark      bool
	C           map[string]string
	R           theme.Resolved
}

// Context is what a template sees.
//
// The rule the whole project rests on: a template reads .R — a role — and never
// .C, the raw palette. Changing syntax.keyword in roles.json then repaints
// every port at once. .C exists for the handful of targets that publish the
// palette as variables for the user's own stylesheet, and using it anywhere
// else is a lint failure.
type Context struct {
	// the theme
	Name        string
	Label       string
	Description string
	Author      string
	URL         string

	// this flavour
	Flavor      string
	FlavorLabel string
	FlavorDesc  string
	Appearance  string
	IsDark      bool

	R       theme.Resolved    // roles -> hex — what a template uses
	C       map[string]string // the raw palette — variable lists only
	Order   []string
	Meaning map[string]string

	// All reaches the other flavours, for formats that keep a light block
	// alongside the dark one: {{ (index .All "garoa").C.laje }}. The theme is
	// dark-only, so those blocks mirror the dark side.
	All map[string]FlavorRef

	// Meta is the registry's view of this port.
	Meta
}

// File is one rendered output.
type File struct {
	Path    string
	Content []byte
	Flavor  string // empty when the template does not iterate flavours
}

// Template is a parsed port mapping, ready to render.
type Template struct {
	Source string // path the template was read from
	Spec   Spec
	Body   string

	tpl *template.Template
}

// Parse reads a template's frontmatter and compiles its body.
func Parse(source string, src []byte) (*Template, error) {
	fm, body, err := SplitFrontmatter(src)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", source, err)
	}

	// missingkey=error turns a mistyped role into a build failure. The default
	// would render "<no value>" into a config file and ship it.
	tpl, err := template.New(filepath.Base(source)).
		Funcs(Funcs).
		Option("missingkey=error").
		Parse(body)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", source, err)
	}

	return &Template{Source: source, Spec: fm.SPN, Body: body, tpl: tpl}, nil
}

// Render produces every file this template describes.
func (t *Template) Render(pal *theme.Palette, roles theme.Roles, meta Meta) ([]File, error) {
	refs, err := flavorRefs(pal, roles)
	if err != nil {
		return nil, err
	}

	build := func(f theme.Flavor) (Context, error) {
		resolved, err := roles.Resolve(f)
		if err != nil {
			return Context{}, err
		}
		return Context{
			Name: pal.Name, Label: pal.Label, Description: pal.Description,
			Author: pal.Author, URL: pal.URL,
			Flavor: f.ID, FlavorLabel: f.Label, FlavorDesc: f.Description,
			Appearance: f.Appearance, IsDark: f.IsDark(),
			R: resolved, C: f.Colors, Order: pal.Order, Meaning: pal.Meaning,
			All: refs, Meta: meta.forFlavor(f),
		}, nil
	}

	// Without a flavour axis the template renders once. It still needs a
	// flavour in context — the first one — because a single-file target has to
	// pick a flavour to be. eza is the case: it reads one theme.yml.
	if !t.Spec.PerFlavor() {
		ctx, err := build(pal.Flavors[0])
		if err != nil {
			return nil, err
		}
		f, err := t.renderOne(ctx)
		if err != nil {
			return nil, err
		}
		f.Flavor = ""
		return []File{f}, nil
	}

	var out []File
	for _, fl := range pal.Flavors {
		ctx, err := build(fl)
		if err != nil {
			return nil, err
		}
		f, err := t.renderOne(ctx)
		if err != nil {
			return nil, err
		}
		f.Flavor = fl.ID
		out = append(out, f)
	}
	return out, nil
}

// RenderFlavor renders a single named flavour. Used by a single-file target
// that has to be installed as one specific flavour.
func (t *Template) RenderFlavor(pal *theme.Palette, roles theme.Roles, meta Meta, flavor string) (File, error) {
	fl, ok := pal.Flavor(flavor)
	if !ok {
		return File{}, fmt.Errorf("unknown flavour %q", flavor)
	}
	refs, err := flavorRefs(pal, roles)
	if err != nil {
		return File{}, err
	}
	resolved, err := roles.Resolve(fl)
	if err != nil {
		return File{}, err
	}
	ctx := Context{
		Name: pal.Name, Label: pal.Label, Description: pal.Description,
		Author: pal.Author, URL: pal.URL,
		Flavor: fl.ID, FlavorLabel: fl.Label, FlavorDesc: fl.Description,
		Appearance: fl.Appearance, IsDark: fl.IsDark(),
		R: resolved, C: fl.Colors, Order: pal.Order, Meaning: pal.Meaning,
		All: refs, Meta: meta.forFlavor(fl),
	}
	f, err := t.renderOne(ctx)
	if err != nil {
		return File{}, err
	}
	f.Flavor = fl.ID
	return f, nil
}

func (t *Template) renderOne(ctx Context) (File, error) {
	path, err := t.outputPath(ctx)
	if err != nil {
		return File{}, err
	}

	var buf bytes.Buffer
	if err := t.tpl.Execute(&buf, ctx); err != nil {
		return File{}, fmt.Errorf("%s -> %s: %w", t.Source, path, err)
	}
	return File{Path: path, Content: buf.Bytes()}, nil
}

// outputPath expands spn.filename against the same context the body sees, so
// a path can name the flavour it belongs to.
func (t *Template) outputPath(ctx Context) (string, error) {
	name, err := template.New("filename").
		Funcs(Funcs).
		Option("missingkey=error").
		Parse(t.Spec.Filename)
	if err != nil {
		return "", fmt.Errorf("%s: spn.filename: %w", t.Source, err)
	}
	var buf bytes.Buffer
	if err := name.Execute(&buf, ctx); err != nil {
		return "", fmt.Errorf("%s: spn.filename: %w", t.Source, err)
	}
	out := buf.String()
	if out == "" {
		return "", fmt.Errorf("%s: spn.filename rendered empty", t.Source)
	}
	if filepath.IsAbs(out) {
		return "", fmt.Errorf("%s: spn.filename %q must be relative to the repository", t.Source, out)
	}
	return filepath.Clean(out), nil
}

func flavorRefs(pal *theme.Palette, roles theme.Roles) (map[string]FlavorRef, error) {
	refs := map[string]FlavorRef{}
	for _, f := range pal.Flavors {
		resolved, err := roles.Resolve(f)
		if err != nil {
			return nil, err
		}
		refs[f.ID] = FlavorRef{
			Flavor: f.ID, Label: f.Label, Appearance: f.Appearance,
			Description: f.Description, IsDark: f.IsDark(),
			C: f.Colors, R: resolved,
		}
	}
	return refs, nil
}

// SortFiles orders rendered files by path so reports and diffs are stable.
func SortFiles(files []File) {
	sort.SliceStable(files, func(i, j int) bool { return files[i].Path < files[j].Path })
}
