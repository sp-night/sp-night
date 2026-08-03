// Package port turns a registry entry into the artefacts a port repository
// publishes: the synthetic preview and the README.
//
// Both are generated for the same reason the theme files are. A screenshot
// drifts from what the user installs the moment the palette is retuned, and a
// hand-written README's role table drifts from the template the moment someone
// edits one and not the other.
package port

import (
	"fmt"
	"math"
	"strings"

	"github.com/sp-night/sp-night/internal/theme"
	"github.com/sp-night/sp-night/registry"
)

// The mockup geometry. A terminal window, not a screenshot of one: the colours
// come from the palette, so the preview cannot show something the user will not
// get.
const (
	width  = 840
	height = 462
	radius = 12

	titlebarHeight = 42
	dotY           = 21
	dotRadius      = 5.5

	bodyLeft   = 26
	bodyTop    = 82
	lineHeight = 26

	swatchLabelY = 384
	swatchY      = 396
	swatchHeight = 26
	swatchRadius = 4
	swatchGap    = 6

	monoStyle = `.mono { font: 15px ui-monospace, 'JetBrains Mono', 'Fira Code', Menlo, monospace; }`
)

// SVG renders the preview for one flavour.
func SVG(p registry.Port, pal *theme.Palette, roles theme.Roles, flavor theme.Flavor) ([]byte, error) {
	resolved, err := roles.Resolve(flavor)
	if err != nil {
		return nil, err
	}

	// look resolves a span or swatch reference: a role, or a raw palette key.
	look := func(role, key string) (string, error) {
		switch {
		case role != "":
			group, name, ok := strings.Cut(role, ".")
			if !ok {
				return "", fmt.Errorf("role %q is not in group.role form", role)
			}
			hex, ok := resolved[group][name]
			if !ok {
				return "", fmt.Errorf("role %q does not exist", role)
			}
			return hex, nil
		case key != "":
			hex, ok := flavor.Colors[key]
			if !ok {
				return "", fmt.Errorf("palette colour %q does not exist", key)
			}
			return hex, nil
		default:
			return "", fmt.Errorf("neither a role nor a palette key")
		}
	}

	role := func(name string) string {
		group, r, _ := strings.Cut(name, ".")
		return resolved[group][r]
	}

	subst := func(s string) string {
		s = strings.ReplaceAll(s, "{flavor}", flavor.ID)
		return strings.ReplaceAll(s, "{label}", flavor.Label)
	}

	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d" role="img" aria-label="%s themed with SP Night %s">`+"\n",
		width, height, width, height, escAttr(p.Name), escAttr(flavor.Label))
	fmt.Fprintf(&b, "  <style>%s</style>\n", monoStyle)

	// window, border, titlebar
	fmt.Fprintf(&b, "  <rect width=\"%d\" height=\"%d\" rx=\"%d\" fill=\"%s\"/>\n",
		width, height, radius, role("ui.bg"))
	fmt.Fprintf(&b, "  <rect x=\"0.5\" y=\"0.5\" width=\"%d\" height=\"%d\" rx=\"%s\" fill=\"none\" stroke=\"%s\"/>\n",
		width-1, height-1, f1(radius-0.5), role("ui.border"))
	fmt.Fprintf(&b, "  <path d=\"M0 %da%d %d 0 0 1 %d-%dh%da%d %d 0 0 1 %d %dv%dH0z\" fill=\"%s\"/>\n",
		radius, radius, radius, radius, radius, width-2*radius, radius, radius, radius, radius,
		titlebarHeight-radius, role("ui.panel"))
	fmt.Fprintf(&b, "  <line x1=\"0\" y1=\"%s\" x2=\"%d\" y2=\"%s\" stroke=\"%s\"/>\n",
		f1(titlebarHeight+0.5), width, f1(titlebarHeight+0.5), role("ui.border"))

	// The three window dots, drawn from the palette rather than a screenshot's
	// idea of what a window looks like.
	for i, key := range []string{"diagnostic.error", "diagnostic.warn", "diagnostic.ok"} {
		fmt.Fprintf(&b, "  <circle cx=\"%d\" cy=\"%d\" r=\"%s\" fill=\"%s\"/>\n",
			24+i*18, dotY, f1(dotRadius), role(key))
	}
	fmt.Fprintf(&b, "  <text x=\"%d\" y=\"26\" text-anchor=\"middle\" class=\"mono\" fill=\"%s\">%s</text>\n",
		width/2, role("ui.fg_muted"), escText(subst(p.Preview.Title)))

	// the fake session — an empty line is a vertical gap, not a blank row
	for i, line := range p.Preview.Body {
		if len(line) == 0 {
			continue
		}
		y := bodyTop + i*lineHeight
		fmt.Fprintf(&b, "  <text x=\"%d\" y=\"%d\" xml:space=\"preserve\" class=\"mono\">", bodyLeft, y)
		for _, s := range line {
			hex, err := look(s.Role, s.Key)
			if err != nil {
				return nil, fmt.Errorf("%s preview: %w", p.Slug, err)
			}
			bold := ""
			if s.Bold {
				bold = ` font-weight="bold"`
			}
			fmt.Fprintf(&b, `<tspan fill="%s"%s>%s</tspan>`, hex, bold, escText(subst(s.Text)))
		}
		b.WriteString("</text>\n")
	}

	// the colour strip
	fmt.Fprintf(&b, "  <text x=\"%d\" y=\"%d\" class=\"mono\" fill=\"%s\" font-size=\"12\">%s</text>\n",
		bodyLeft, swatchLabelY, role("ui.fg_muted"), escText(subst(p.Preview.Swatches.Label)))

	refs := p.Preview.Swatches.Roles
	raw := len(refs) == 0
	if raw {
		refs = p.Preview.Swatches.Keys
	}
	n := len(refs)
	span := float64(width - 2*bodyLeft)
	w := (span - swatchGap*float64(n-1)) / float64(n)
	for i, ref := range refs {
		var hex string
		var err error
		if raw {
			hex, err = look("", ref)
		} else {
			hex, err = look(ref, "")
		}
		if err != nil {
			return nil, fmt.Errorf("%s preview swatches: %w", p.Slug, err)
		}
		x := float64(bodyLeft) + float64(i)*(w+swatchGap)
		fmt.Fprintf(&b, "  <rect x=\"%s\" y=\"%d\" width=\"%s\" height=\"%d\" rx=\"%d\" fill=\"%s\"/>\n",
			f1(x), swatchY, f1(w), swatchHeight, swatchRadius, hex)
	}

	b.WriteString("</svg>\n")
	return []byte(b.String()), nil
}

// f1 formats to one decimal, rounding halves away from zero. Go's %.1f rounds
// halves to even, which would render the swatch stride as 125.2 where the
// published previews write 125.3. The trailing .0 is kept for the same reason.
func f1(v float64) string {
	r := math.Floor(math.Abs(v)*10+0.5) / 10
	if v < 0 {
		r = -r
	}
	return fmt.Sprintf("%.1f", r)
}

// escText escapes XML character data. Only &, < and > are special there — a
// double quote is ordinary text, and escaping it to &#34; would turn a Lua
// string in the preview into noise.
func escText(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	return strings.ReplaceAll(s, ">", "&gt;")
}

// escAttr escapes an attribute value, where the quote delimiter does matter.
func escAttr(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	return strings.ReplaceAll(s, `"`, "&quot;")
}
