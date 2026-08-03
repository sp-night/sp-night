package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/sp-night/sp-night/internal/audit"
	"github.com/sp-night/sp-night/internal/render"
)

func runGen(args []string) error {
	fs := flag.NewFlagSet("gen", flag.ContinueOnError)
	dir := paletteFlag(fs)
	check := fs.Bool("check", false, "do not write; fail if any file on disk is out of date")
	stdout := fs.Bool("stdout", false, "write to stdout instead of to files")
	app := fs.String("app", "", "registry slug for this port (default: inferred from the template name)")
	regPath := fs.String("registry", "", "read the port catalogue from this file instead of the embedded copy")
	quiet := fs.Bool("q", false, "print only failures")
	if err := parseArgs(fs, args); err != nil {
		return err
	}

	pal, roles, err := loadContract(*dir)
	if err != nil {
		return err
	}

	templates, err := findTemplates(fs.Args())
	if err != nil {
		return err
	}

	var rendered []render.File
	for _, path := range templates {
		t, err := parseTemplate(path)
		if err != nil {
			return err
		}
		if err := render.CheckVersion(t.Spec.Version, version); err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}

		slug := *app
		if slug == "" {
			slug = slugFromTemplate(path)
		}
		meta, err := metaFor(slug, *regPath)
		if err != nil {
			return err
		}

		files, err := t.Render(pal, roles, meta)
		if err != nil {
			return err
		}
		// Output paths are relative to the directory holding the template,
		// which is the port repository root.
		base := filepath.Dir(path)
		for i := range files {
			files[i].Path = filepath.Join(base, files[i].Path)
		}
		rendered = append(rendered, files...)
	}
	render.SortFiles(rendered)

	// KDE schemes are audited on their OUTPUT: every fg/bg pair of every
	// section of the file that was just built.
	built := map[string][]byte{}
	for _, f := range rendered {
		built[f.Path] = f.Content
	}
	if n := audit.KDESchemes(built, os.Stdout); n > 0 {
		return fmt.Errorf("%d KDE scheme pair(s) below %.1f:1", n, audit.LevelAA)
	}

	switch {
	case *stdout:
		for _, f := range rendered {
			if len(rendered) > 1 {
				fmt.Printf("── %s\n", f.Path)
			}
			os.Stdout.Write(f.Content)
		}
		return nil

	case *check:
		return checkStale(rendered)

	default:
		return writeFiles(rendered, *quiet)
	}
}

// checkStale is what a port's CI runs. It never writes, and its failure message
// says what to do about it.
func checkStale(files []render.File) error {
	var stale, missing []string
	for _, f := range files {
		cur, err := os.ReadFile(f.Path)
		switch {
		case errors.Is(err, os.ErrNotExist):
			missing = append(missing, f.Path)
		case err != nil:
			return err
		case !bytes.Equal(cur, f.Content):
			stale = append(stale, f.Path)
		}
	}

	if len(stale) == 0 && len(missing) == 0 {
		fmt.Printf("%d file(s) match their template\n", len(files))
		return nil
	}

	var b strings.Builder
	fmt.Fprintf(&b, "%d file(s) out of date — run `spn gen`:", len(stale)+len(missing))
	for _, p := range missing {
		fmt.Fprintf(&b, "\n  %s (missing)", p)
	}
	for _, p := range stale {
		fmt.Fprintf(&b, "\n  %s", p)
	}
	return errors.New(b.String())
}

func writeFiles(files []render.File, quiet bool) error {
	written := 0
	for _, f := range files {
		if err := os.MkdirAll(filepath.Dir(f.Path), 0o755); err != nil {
			return err
		}
		// Skip a write when the bytes already match, so regenerating does not
		// churn mtimes and make a watcher loop.
		if cur, err := os.ReadFile(f.Path); err == nil && bytes.Equal(cur, f.Content) {
			continue
		}
		if err := os.WriteFile(f.Path, f.Content, 0o644); err != nil {
			return err
		}
		written++
		if !quiet {
			fmt.Printf("  %s\n", f.Path)
		}
	}
	if !quiet {
		fmt.Printf("%d file(s), %d changed\n", len(files), written)
	}
	return nil
}

func runLint(args []string) error {
	fs := flag.NewFlagSet("lint", flag.ContinueOnError)
	if err := parseArgs(fs, args); err != nil {
		return err
	}

	templates, err := findTemplates(fs.Args())
	if err != nil {
		return err
	}

	total := 0
	for _, path := range templates {
		t, err := parseTemplate(path)
		if err != nil {
			return err
		}
		findings := t.Lint()
		total += len(findings)
		for _, f := range findings {
			fmt.Printf("  %s:%s\n", path, f)
		}
	}

	if total > 0 {
		return fmt.Errorf("%d template rule violation(s)", total)
	}
	fmt.Printf("%d template(s) ask for roles, not colours\n", len(templates))
	return nil
}

func parseTemplate(path string) (*render.Template, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return render.Parse(path, src)
}

// findTemplates uses the arguments when given, and otherwise every *.tmpl in
// the current directory — so a port repository runs bare `spn gen`.
func findTemplates(args []string) ([]string, error) {
	if len(args) > 0 {
		return args, nil
	}
	matches, err := filepath.Glob("*.tmpl")
	if err != nil {
		return nil, err
	}
	if len(matches) == 0 {
		return nil, errors.New("no *.tmpl in this directory — name one, or run from a port repository")
	}
	sort.Strings(matches)
	return matches, nil
}

// slugFromTemplate turns "eza.yml.tmpl" into "eza": the registry slug is the
// first segment, which is also the repository name under the org.
func slugFromTemplate(path string) string {
	name := filepath.Base(path)
	name = strings.TrimSuffix(name, ".tmpl")
	if i := strings.Index(name, "."); i > 0 {
		name = name[:i]
	}
	return name
}

// metaFor is filled in by the registry.
//
// A slug that is not in the catalogue is an error rather than an empty Meta. The
// header of every generated file reads .Repo and .Install from here, so a missing
// entry used to render a file with a blank install path and ship it — silence is
// the wrong answer to "this port is not listed yet".
var metaFor = func(slug, registryPath string) (render.Meta, error) { return render.Meta{}, nil }
