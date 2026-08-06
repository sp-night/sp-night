package port

import (
	"fmt"
	"strings"

	"github.com/sp-night/sp-night/internal/render"
	"github.com/sp-night/sp-night/registry"
)

// FallbackVersion is what a mapping declares when the scaffolding binary does
// not know its own version — a `go run` from a checkout rather than a release.
// A range, because a dev build cannot honestly name a released version.
const FallbackVersion = "^1.0"

// Scaffold returns the files a new port repository starts with.
//
// Everything here is either boilerplate identical across ports, or derived from
// the catalogue entry. The only thing left for a human is the mapping itself —
// which key of the app means which role — and that is the one thing a human
// should be deciding.
func Scaffold(p registry.Port, spnVersion string) ([]render.File, error) {
	if p.Template == "" {
		return nil, fmt.Errorf("%s: the catalogue entry has no template name", p.Slug)
	}

	files := []render.File{
		{Path: p.Template, Content: []byte(mappingStub(p, spnVersion))},
		{Path: ".github/workflows/theme.yml", Content: []byte(workflow(p))},
		{Path: ".gitattributes", Content: []byte(gitAttributes(p))},
		{Path: "renovate.json", Content: []byte(renovateConfig)},
		{Path: ".editorconfig", Content: []byte(editorConfig)},
		{Path: ".gitignore", Content: []byte(gitignore)},
	}
	render.SortFiles(files)
	return files, nil
}

// mappingStub is the template a port starts from: the canonical header wired to
// the catalogue, and one commented example of the rule that matters.
func mappingStub(p registry.Port, spnVersion string) string {
	filename := "themes/sp_night_{{ .Flavor }}"
	if ext := configExtension(p.Install); ext != "" {
		filename += ext
	}

	var b strings.Builder
	fmt.Fprintf(&b, "---\nspn:\n  version: %q\n  matrix: [flavor]\n  filename: %q\n---\n", spnVersion, filename)
	fmt.Fprintf(&b, "# SP Night — {{ .FlavorLabel }}\n")
	fmt.Fprintf(&b, "# {{ .FlavorDesc }}\n")
	fmt.Fprintf(&b, "#\n")
	fmt.Fprintf(&b, "# Generated from the SP Night palette — do not edit by hand.\n")
	fmt.Fprintf(&b, "# {{ .Repo }}\n")
	fmt.Fprintf(&b, "#\n")
	fmt.Fprintf(&b, "# install: {{ .Install }}\n")
	if p.Activate != "" {
		fmt.Fprintf(&b, "# enable:  {{ .Activate }}\n")
	}
	fmt.Fprintf(&b, "\n")
	fmt.Fprintf(&b, "# Write the mapping below. Ask for a role, never a colour:\n")
	fmt.Fprintf(&b, "#\n")
	fmt.Fprintf(&b, "#   ✗  background = {{ \"{{\" }} .C.laje }}        the palette directly\n")
	fmt.Fprintf(&b, "#   ✓  background = {{ \"{{\" }} .R.ui.bg }}      the role\n")
	fmt.Fprintf(&b, "#\n")
	fmt.Fprintf(&b, "# `spn lint` fails on the first form. Roles are listed by `spn palette --roles`,\n")
	fmt.Fprintf(&b, "# and what each one means is in https://sp-night.github.io/spec.\n")
	return b.String()
}

// configExtension guesses the output extension from the install path, so a
// scaffolded mapping writes themes/sp_night_noite.yml rather than a bare name
// when the app expects a suffix.
func configExtension(install string) string {
	base := baseName(install)
	i := strings.LastIndex(base, ".")
	if i <= 0 {
		return "" // Ghostty resolves by exact name and wants no extension
	}
	return base[i:]
}

// workflow is the port's entire CI: install the tool the mapping asks for, and
// fail if anything committed is out of date.
func workflow(p registry.Port) string {
	return fmt.Sprintf(`name: theme

# The mapping in %s pins the exact spn version it was generated with, so this
# file never needs editing when the engine moves. The engine's release fan-out
# bumps that pin and regenerates the theme files in one pull request.

on:
  workflow_dispatch:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  check:
    uses: sp-night/sp-night/.github/workflows/spn-check.yml@v1
    with:
      port: %s
`, p.Template, p.Slug)
}

// gitAttributes marks the derived files as such. It matters for one reason: a
// regeneration pull request should lead with the change to the mapping, which is
// the decision someone made, instead of burying it under a few thousand lines of
// hex. The files are still verified either way, because `spn gen --check` fails
// if any of them drifts.
//
// It lives in the scaffold rather than being copied between ports because it was
// copied between ports, and the copies carried terminal escape codes in from the
// shell they were pasted out of.
func gitAttributes(p registry.Port) string {
	return fmt.Sprintf(`# Everything here except the mapping is generated.
#
# Marking it as such collapses it in a diff, so a regeneration pull request leads
# with the change to %s, the decision someone made, instead of burying it under a
# few thousand lines of hex. The files are still verified: `+"`spn gen --check`"+`
# fails if any of them drifts from the mapping.
themes/** linguist-generated=true
assets/*.svg linguist-generated=true
README.md linguist-generated=true

# The mapping is the one hand-written file, but a Go template is not this
# repository's language. A colourscheme port does not have one.
*.tmpl linguist-detectable=false
`, p.Template)
}

const renovateConfig = `{
  "$schema": "https://docs.renovatebot.com/renovate-schema.json",
  "extends": [
    "local>sp-night/sp-night//renovate-config"
  ]
}
`

const editorConfig = `root = true

[*]
charset = utf-8
end_of_line = lf
insert_final_newline = true
trim_trailing_whitespace = true
`

const gitignore = `.DS_Store
*.swp
`
