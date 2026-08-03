package port

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/sp-night/sp-night/internal/render"
	"github.com/sp-night/sp-night/internal/theme"
	"github.com/sp-night/sp-night/registry"
)

// The canonical port README.
//
// Every section that a reader uses to decide whether to trust a port — the role
// table, the install path, the flavour list — is derived rather than written, so
// it cannot disagree with the template beside it. Only the app-specific install
// prose is free text, because Ghostty's extensionless-file explanation and
// eza's EZA_COLORS warning are genuinely not derivable from anything.
const readmeTemplate = `<p align="center">
  <a href="https://sp-night.github.io">
    <img src="https://raw.githubusercontent.com/sp-night/sp-night.github.io/main/public/logo-noite.svg" width="120" alt="{{ .Copy.LogoAlt }}">
  </a>
</p>

<h1 align="center">SP Night for <a href="{{ .Port.Homepage }}">{{ .Port.Name }}</a></h1>

<p align="center">
  <strong>{{ .Copy.TaglineStrong }}</strong><br>
{{ .TaglineHTML }}
</p>

<p align="center">
  <a href="https://sp-night.github.io"><strong>sp-night.github.io</strong></a>
  &nbsp;·&nbsp;
  <a href="https://sp-night.github.io/palette">palette</a>
  &nbsp;·&nbsp;
  <a href="https://sp-night.github.io/spec">spec</a>
  &nbsp;·&nbsp;
  <a href="https://sp-night.github.io/ports">ports</a>
</p>

---

## The flavours

{{ .Copy.FlavoursIntro }}
{{ range .Flavours }}
### {{ .Label }} — ` + "`{{ .File }}`" + `

{{ .Blurb }}

![{{ $.Port.Name }} themed with SP Night {{ .Label }}](assets/preview-{{ .ID }}.svg)
{{ end }}
## Install

{{ .Port.InstallGuide }}

## What gets themed

| {{ .KeyLabel }} | Role | Meaning |
|---|---|---|
{{- range .Port.Mapping }}
| {{ .Key }} | {{ .Role }} | {{ .Meaning }} |
{{- end }}

{{ .Copy.Provenance }}

## The mapping

{{ .MappingSection }}

## License

[MIT](LICENSE)
`

// flavourView is one flavour as the README shows it.
type flavourView struct {
	ID    string
	Label string
	Blurb string
	File  string // the file name this flavour ships as, e.g. sp_night_noite.yml
}

type readmeData struct {
	Port           registry.Port
	Copy           *registry.Copy
	Flavours       []flavourView
	TaglineHTML    string
	MappingSection string
	KeyLabel       string
}

// README renders the canonical README for a port.
//
// tpl is the port's own mapping: the flavour file names come from its
// spn.filename, so the README can never name a file the template does not
// produce.
func README(p registry.Port, c *registry.Copy, pal *theme.Palette, roles theme.Roles, tpl *render.Template) ([]byte, error) {
	files, err := tpl.Render(pal, roles, MetaOf(p))
	if err != nil {
		return nil, fmt.Errorf("%s: rendering the mapping to name the flavour files: %w", p.Slug, err)
	}

	byFlavor := map[string]string{}
	for _, f := range files {
		byFlavor[f.Flavor] = baseName(f.Path)
	}

	var views []flavourView
	for _, fl := range pal.Flavors {
		blurb, ok := c.Blurb(fl.ID)
		if !ok {
			return nil, fmt.Errorf("copy: no English blurb for flavour %q", fl.ID)
		}
		name, ok := byFlavor[fl.ID]
		if !ok {
			// A single-file target renders once; every flavour still ships as a
			// file in themes/, named after the flavour.
			name = fmt.Sprintf("sp_night_%s", fl.ID)
			if len(files) == 1 {
				name = baseName(files[0].Path)
			}
		}
		views = append(views, flavourView{ID: fl.ID, Label: fl.Label, Blurb: blurb, File: name})
	}

	// The tagline is a centred HTML block, so its line breaks are <br>.
	var tagline strings.Builder
	lines := strings.Split(strings.TrimRight(c.Tagline, "\n"), "\n")
	for i, line := range lines {
		tagline.WriteString("  " + line)
		if i < len(lines)-1 {
			tagline.WriteString("<br>")
		}
		if i < len(lines)-1 {
			tagline.WriteString("\n")
		}
	}

	mapping := strings.ReplaceAll(c.MappingSection, "{template}", p.Template)
	mapping = strings.ReplaceAll(mapping, "{app}", p.Name)
	mapping = strings.ReplaceAll(mapping, "{keylabel}", keyLabel(p))
	scope := ""
	if p.MappingScope != "" {
		scope = ", " + p.MappingScope
	}
	mapping = strings.ReplaceAll(mapping, "{scope}", scope)
	closer := p.MappingCloser
	if closer == "" {
		closer = "what anything means"
	}
	mapping = strings.ReplaceAll(mapping, "{closer}", closer)

	data := readmeData{
		Port: p, Copy: c, Flavours: views,
		KeyLabel:       keyLabel(p),
		TaglineHTML:    tagline.String(),
		MappingSection: strings.TrimRight(mapping, "\n"),
	}

	t, err := template.New("readme").Option("missingkey=error").Parse(readmeTemplate)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// MetaOf is the registry's view of a port, as a template sees it.
func MetaOf(p registry.Port) render.Meta {
	return render.Meta{
		App: p.Name, Slug: p.Slug, Repo: p.Repo, Homepage: p.Homepage,
		Install: p.Install, Activate: p.Activate,
	}
}

func baseName(path string) string {
	if i := strings.LastIndexAny(path, `/\`); i >= 0 {
		return path[i+1:]
	}
	return path
}

// keyLabel is the mapping table's first-column header.
func keyLabel(p registry.Port) string {
	if p.KeyLabel != "" {
		return p.KeyLabel
	}
	return p.Name + " key"
}
