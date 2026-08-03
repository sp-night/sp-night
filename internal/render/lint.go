package render

import (
	"fmt"
	"regexp"
	"strings"
)

// Linting the one rule the whole project rests on.
//
// A template must never reference a palette colour directly:
//
//	✗  palette = 4={{ .C.marginal }}
//	✓  palette = 4={{ .R.ansi.blue }}
//
// Why: moving syntax.keyword from marginal to temporal has to repaint Neovim,
// bat, fish and the website at once. A template with marginal written into it
// stays behind and nobody notices — which is exactly how a family of ports
// drifts apart over a couple of years.
//
// Until now this was prose in the style guide. Here it is a gate.
//
// The exception is the variable lists: waybar.css, gtk.css and hyprland.conf
// publish the raw palette as @define-color / $var because the end user will
// want to write @sp_sodio in their own stylesheet. Those templates declare
// raw_palette: true in their frontmatter, and even then everything the template
// itself styles must still use a role.

// rawPaletteRe finds a .C reference: {{ .C.sodio }}, {{ index .C "sodio" }},
// {{ $.C.laje }}, or a range over .C.
var rawPaletteRe = regexp.MustCompile(`\$?\.C\b`)

// Finding is one lint result.
type Finding struct {
	Line    int
	Column  int
	Excerpt string
	Message string
}

func (f Finding) String() string {
	return fmt.Sprintf("%d:%d: %s\n      %s", f.Line, f.Column, f.Message, f.Excerpt)
}

// Lint checks a parsed template against the roles-only rule.
func (t *Template) Lint() []Finding {
	if t.Spec.RawPalette {
		return nil
	}

	var out []Finding
	for i, line := range strings.Split(t.Body, "\n") {
		for _, loc := range rawPaletteRe.FindAllStringIndex(line, -1) {
			out = append(out, Finding{
				Line:    i + 1,
				Column:  loc[0] + 1,
				Excerpt: strings.TrimSpace(line),
				Message: "references the raw palette (.C) instead of a role (.R). " +
					"Ask for a role, not a colour — or set raw_palette: true if this template " +
					"publishes the palette as variables for the user's own stylesheet.",
			})
		}
	}
	return out
}
