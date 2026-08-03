package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sp-night/sp-night/internal/port"
	"github.com/sp-night/sp-night/internal/render"
	"github.com/sp-night/sp-night/registry"
)

// The registry travels in the binary alongside the palette, so a port
// repository holds no copy of it. --registry reads it from disk instead, which
// is how this repository edits the catalogue and sees the result.
func loadRegistry(path string) (*registry.Registry, error) {
	if path == "" {
		return registry.Embedded()
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return registry.Load(data)
}

func registryFlags(fs *flag.FlagSet) (*string, *string) {
	return fs.String("registry", "", "read the port catalogue from this file instead of the embedded copy"),
		fs.String("copy", "", "read the shared README prose from this file instead of the embedded copy")
}

func loadCopy(path string) (*registry.Copy, error) {
	if path == "" {
		return registry.EmbeddedCopy()
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return registry.LoadCopy(data)
}

// resolvePort finds the registry entry a command is about: the argument when
// given, and otherwise the single *.tmpl in the current directory — so a port
// repository runs bare `spn readme`.
func resolvePort(reg *registry.Registry, args []string) (registry.Port, error) {
	slug := ""
	if len(args) > 0 {
		slug = args[0]
	}
	if slug == "" {
		templates, err := findTemplates(nil)
		if err != nil {
			return registry.Port{}, err
		}
		if len(templates) != 1 {
			return registry.Port{}, fmt.Errorf("%d templates here — name the port explicitly", len(templates))
		}
		slug = slugFromTemplate(templates[0])
	}
	p, ok := reg.Port(slug)
	if !ok {
		return registry.Port{}, fmt.Errorf("%q is not in the port catalogue (have: %v)", slug, reg.Slugs())
	}
	return p, nil
}

func runPreview(args []string) error {
	fs := flag.NewFlagSet("preview", flag.ContinueOnError)
	dir := paletteFlag(fs)
	regPath, _ := registryFlags(fs)
	out := fs.String("out", "assets", "directory to write the previews into")
	check := fs.Bool("check", false, "do not write; fail if any preview is out of date")
	if err := parseArgs(fs, args); err != nil {
		return err
	}

	pal, roles, err := loadContract(*dir)
	if err != nil {
		return err
	}
	reg, err := loadRegistry(*regPath)
	if err != nil {
		return err
	}
	p, err := resolvePort(reg, fs.Args())
	if err != nil {
		return err
	}

	var files []render.File
	for _, fl := range pal.Flavors {
		svg, err := port.SVG(p, pal, roles, fl)
		if err != nil {
			return err
		}
		files = append(files, render.File{
			Path:    filepath.Join(*out, fmt.Sprintf("preview-%s.svg", fl.ID)),
			Content: svg,
			Flavor:  fl.ID,
		})
	}

	if *check {
		return checkStale(files)
	}
	return writeFiles(files, false)
}

func runReadme(args []string) error {
	fs := flag.NewFlagSet("readme", flag.ContinueOnError)
	dir := paletteFlag(fs)
	regPath, copyPath := registryFlags(fs)
	out := fs.String("out", "README.md", "file to write")
	check := fs.Bool("check", false, "do not write; fail if the README is out of date")
	stdout := fs.Bool("stdout", false, "write to stdout instead of to a file")
	if err := parseArgs(fs, args); err != nil {
		return err
	}

	pal, roles, err := loadContract(*dir)
	if err != nil {
		return err
	}
	reg, err := loadRegistry(*regPath)
	if err != nil {
		return err
	}
	cp, err := loadCopy(*copyPath)
	if err != nil {
		return err
	}
	p, err := resolvePort(reg, fs.Args())
	if err != nil {
		return err
	}

	tpl, err := parseTemplate(p.Template)
	if err != nil {
		return fmt.Errorf("the README names the files the mapping produces, so %s has to be readable: %w", p.Template, err)
	}

	content, err := port.README(p, cp, pal, roles, tpl)
	if err != nil {
		return err
	}

	if *stdout {
		_, err := os.Stdout.Write(content)
		return err
	}
	files := []render.File{{Path: *out, Content: content}}
	if *check {
		return checkStale(files)
	}
	return writeFiles(files, false)
}

// runNew scaffolds a port repository: the mapping stub, the workflow, and the
// files a repository needs before anything can be generated into it.
func runNew(args []string) error {
	fs := flag.NewFlagSet("new", flag.ContinueOnError)
	regPath, _ := registryFlags(fs)
	out := fs.String("out", "", "directory to scaffold into (default: the slug)")
	force := fs.Bool("force", false, "overwrite files the scaffold would replace")
	if err := parseArgs(fs, args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return errors.New("usage: spn new <slug>")
	}
	slug := fs.Arg(0)

	reg, err := loadRegistry(*regPath)
	if err != nil {
		return err
	}
	p, ok := reg.Port(slug)
	if !ok {
		return fmt.Errorf("%q is not in the port catalogue yet.\n"+
			"Add it to registry/ports.yml first — the catalogue is what the README,\n"+
			"the preview and the generated header all read from.", slug)
	}

	root := *out
	if root == "" {
		root = slug
	}

	scaffold, err := port.Scaffold(p)
	if err != nil {
		return err
	}

	// Scaffolding into an existing directory is the normal case, not an error:
	// the documented flow is to create the repository, clone it — so the
	// directory already holds a LICENSE — and scaffold into the clone. What must
	// not happen is silently replacing a file someone wrote, so a collision stops
	// the whole thing before anything is written.
	var clash []string
	for _, f := range scaffold {
		if _, err := os.Stat(filepath.Join(root, f.Path)); err == nil {
			clash = append(clash, f.Path)
		}
	}
	if len(clash) > 0 && !*force {
		return fmt.Errorf("%s already has %s — pass --force to overwrite",
			root, strings.Join(clash, ", "))
	}

	for _, f := range scaffold {
		full := filepath.Join(root, f.Path)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(full, f.Content, 0o644); err != nil {
			return err
		}
		fmt.Printf("  %s\n", full)
	}

	fmt.Printf("\n%s scaffolded. Next:\n", root)
	fmt.Printf("  1. write the mapping in %s/%s — ask for roles, never colours\n", root, p.Template)
	fmt.Printf("  2. cd %s && spn gen && spn preview && spn readme\n", root)
	fmt.Printf("  3. spn lint && spn gen --check\n")
	return nil
}

// verifyRegistry is `spn registry`: it proves the catalogue is well formed and
// that every listed port's repository URL follows from its slug.
func runRegistry(args []string) error {
	fs := flag.NewFlagSet("registry", flag.ContinueOnError)
	regPath, copyPath := registryFlags(fs)
	asJSON := fs.Bool("json", false, "print the catalogue as JSON — used to build the fan-out matrix")
	if err := parseArgs(fs, args); err != nil {
		return err
	}

	reg, err := loadRegistry(*regPath)
	if err != nil {
		return err
	}
	cp, err := loadCopy(*copyPath)
	if err != nil {
		return err
	}
	if _, ok := cp.Blurb("noite"); !ok {
		return errors.New("copy: the noite blurb is missing")
	}

	if *asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetEscapeHTML(false)
		return enc.Encode(struct {
			Ports []registry.Port `json:"ports"`
			Slugs []string        `json:"slugs"`
		}{reg.Ports, reg.Slugs()})
	}

	for _, group := range reg.GroupOrder() {
		fmt.Printf("\n  %s\n", reg.Groups[group])
		for _, p := range reg.ByGroup(group) {
			fmt.Printf("    %-10s %-28s %s\n", p.Slug, p.Template, p.Repo)
		}
	}
	fmt.Printf("\n%d port(s), %d group(s), catalogue ok\n", len(reg.Ports), len(reg.Groups))
	return nil
}

// metaFromRegistry replaces the placeholder in gen.go, so a generated file's
// header comes from the catalogue instead of being typed into the template.
func init() {
	metaFor = func(slug, registryPath string) (render.Meta, error) {
		reg, err := loadRegistry(registryPath)
		if err != nil {
			return render.Meta{}, err
		}
		p, ok := reg.Port(slug)
		if !ok {
			return render.Meta{}, fmt.Errorf(
				"%q is not in the port catalogue, so its install path and repository are unknown.\n"+
					"List it in registry/ports.yml first — that is step 1 of adding a port.\n"+
					"Working on an entry that has not shipped yet? Pass --registry with your local copy.",
				slug)
		}
		return port.MetaOf(p), nil
	}
}
