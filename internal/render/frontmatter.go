package render

import (
	"fmt"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Frontmatter is the block at the top of every template, delimited by ---.
//
//	---
//	spn:
//	  version: "^1.0"
//	  matrix: [flavor]
//	  filename: "themes/sp_night_{{ .Flavor }}"
//	---
//
// It exists so a port's CI does not have to be told anything: the template
// declares which tool versions its mapping is written against, how many files
// it produces, and where they go. A six-line workflow can then install the
// right tool and check the result.
type Frontmatter struct {
	SPN Spec `yaml:"spn"`
}

// Spec is the spn block of the frontmatter.
type Spec struct {
	// Version is the range of tool versions this mapping is written against,
	// e.g. "^1.0". A port's CI reads it to pick the binary to install.
	Version string `yaml:"version"`

	// Matrix lists the axes to iterate. Only "flavor" is defined today; an
	// empty matrix renders once, which is what a target with a single fixed
	// config file needs (eza reads one theme.yml).
	Matrix []string `yaml:"matrix"`

	// Filename is a Go template for the output path, relative to the template.
	Filename string `yaml:"filename"`

	// RawPalette allows this template to reference .C — the raw palette —
	// which is otherwise a lint failure. It is true only for targets that
	// publish the palette as variables for the user's own CSS, because the
	// user will want to write @sp_sodio themselves.
	RawPalette bool `yaml:"raw_palette"`
}

const delimiter = "---"

// SplitFrontmatter separates the frontmatter from the template body.
//
// The body is returned with no leading newline, so a rendered file starts at
// the template's first real line. Getting this wrong would shift every
// generated file by one byte.
func SplitFrontmatter(src []byte) (Frontmatter, string, error) {
	var fm Frontmatter

	text := string(src)
	// A byte order mark would otherwise hide the opening delimiter.
	text = strings.TrimPrefix(text, "\ufeff")

	if !strings.HasPrefix(text, delimiter) {
		return fm, "", fmt.Errorf("no frontmatter: a template must open with a %q line declaring its spn block", delimiter)
	}

	rest := text[len(delimiter):]
	if !strings.HasPrefix(rest, "\n") {
		return fm, "", fmt.Errorf("malformed frontmatter: %q must be alone on the first line", delimiter)
	}
	rest = rest[1:]

	end := strings.Index(rest, "\n"+delimiter)
	if end < 0 {
		return fm, "", fmt.Errorf("unterminated frontmatter: no closing %q line", delimiter)
	}
	block := rest[:end]

	body := rest[end+1+len(delimiter):]
	body = strings.TrimPrefix(body, "\n")

	if err := yaml.Unmarshal([]byte(block), &fm); err != nil {
		return fm, "", fmt.Errorf("frontmatter: %w", err)
	}
	if err := fm.SPN.validate(); err != nil {
		return fm, "", err
	}
	return fm, body, nil
}

func (s Spec) validate() error {
	if s.Filename == "" {
		return fmt.Errorf("frontmatter: spn.filename is required — it is where the rendered file goes")
	}
	if s.Version == "" {
		return fmt.Errorf("frontmatter: spn.version is required — it is how a port's CI picks the right tool")
	}
	for _, axis := range s.Matrix {
		if axis != "flavor" {
			return fmt.Errorf("frontmatter: unknown matrix axis %q (only \"flavor\" is defined)", axis)
		}
	}
	return nil
}

// PerFlavor reports whether the template renders once per flavour.
func (s Spec) PerFlavor() bool {
	for _, axis := range s.Matrix {
		if axis == "flavor" {
			return true
		}
	}
	return false
}

// CheckVersion reports whether the running tool satisfies the template's
// declared range. "dev" builds satisfy everything, so working in this
// repository never fights the constraint.
func CheckVersion(constraint, toolVersion string) error {
	if toolVersion == "dev" || toolVersion == "" {
		return nil
	}
	ok, err := satisfies(constraint, toolVersion)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("this template is written for spn %s, but spn %s is running", constraint, toolVersion)
	}
	return nil
}

// satisfies implements the small slice of semver ranges the frontmatter needs:
// "^X.Y[.Z]" (compatible within the major), ">=X.Y[.Z]", and an exact version.
// A port that needs more than this is a sign the tool broke compatibility too
// often, not that the parser needs features.
func satisfies(constraint, version string) (bool, error) {
	v, err := parseVersion(version)
	if err != nil {
		return false, fmt.Errorf("tool version %q: %w", version, err)
	}

	spec := strings.TrimSpace(constraint)
	switch {
	case strings.HasPrefix(spec, "^"):
		want, err := parseVersion(strings.TrimPrefix(spec, "^"))
		if err != nil {
			return false, fmt.Errorf("constraint %q: %w", constraint, err)
		}
		return v[0] == want[0] && compare(v, want) >= 0, nil

	case strings.HasPrefix(spec, ">="):
		want, err := parseVersion(strings.TrimSpace(strings.TrimPrefix(spec, ">=")))
		if err != nil {
			return false, fmt.Errorf("constraint %q: %w", constraint, err)
		}
		return compare(v, want) >= 0, nil

	default:
		want, err := parseVersion(spec)
		if err != nil {
			return false, fmt.Errorf("constraint %q: %w", constraint, err)
		}
		return compare(v, want) == 0, nil
	}
}

// parseVersion reads "1", "1.2" or "1.2.3", with an optional leading v and any
// pre-release suffix ignored.
func parseVersion(s string) ([3]int, error) {
	var out [3]int
	s = strings.TrimPrefix(strings.TrimSpace(s), "v")
	if i := strings.IndexAny(s, "-+"); i >= 0 {
		s = s[:i]
	}
	if s == "" {
		return out, fmt.Errorf("empty version")
	}
	parts := strings.Split(s, ".")
	if len(parts) > 3 {
		return out, fmt.Errorf("too many components in %q", s)
	}
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil {
			return out, fmt.Errorf("%q is not a number in %q", p, s)
		}
		if n < 0 {
			return out, fmt.Errorf("negative component in %q", s)
		}
		out[i] = n
	}
	return out, nil
}

func compare(a, b [3]int) int {
	for i := range a {
		switch {
		case a[i] > b[i]:
			return 1
		case a[i] < b[i]:
			return -1
		}
	}
	return 0
}
