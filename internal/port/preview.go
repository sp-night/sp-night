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
	"unicode/utf8"

	"github.com/sp-night/sp-night/internal/theme"
	"github.com/sp-night/sp-night/registry"
)

// The mockup geometry. A window, not a screenshot of one: the colours come from
// the palette, so the preview cannot show something the user will not get.
//
// Four frames share the canvas and the colour strip, and differ in the chrome
// above and below the session. Every coordinate here is an integer so that
// only the half-pixel rules and the swatch stride ever pass through f1: the
// published previews are checked byte for byte, and a rounding change would
// shift every one of them.
const (
	width  = 840
	height = 462
	radius = 12

	// terminal: the titlebar with its three dots and a centred title
	titlebarHeight = 42
	dotY           = 21
	dotRadius      = 5.5
	bodyTop        = 82
	bodyLeft       = 26
	lineHeight     = 26

	// editor and app: a shorter strip, text at the left edge
	stripHeight  = 36
	stripTextY   = 23
	stripPad     = 14
	innerBodyTop = 66

	// editor: the gutter, and the tab whose width follows its label
	gutterWidth    = 56
	gutterNumX     = 44
	editorBodyLeft = 66
	tabCharWidth   = 9 // 0.6em at 15px, the advance of every font in monoStyle
	tabPad         = 14
	tabMarker      = 2
	cursorLineRise = 19

	// editor and app: the bottom bar, above the colour strip
	barY     = 342
	barTextY = 360

	// pane: nothing above the session
	paneBodyTop = 44

	swatchLabelY = 384
	swatchY      = 396
	swatchHeight = 26
	swatchRadius = 4
	swatchGap    = 6

	monoStyle = `.mono { font: 15px ui-monospace, 'JetBrains Mono', 'Fira Code', Menlo, monospace; }`
)

// layout is where a frame puts the session.
type layout struct {
	top, left int
}

// SVG renders the preview for one flavour.
func SVG(p registry.Port, pal *theme.Palette, roles theme.Roles, flavor theme.Flavor) ([]byte, error) {
	resolved, err := roles.Resolve(flavor)
	if err != nil {
		return nil, err
	}
	pt := &painter{slug: p.Slug, flavor: flavor, resolved: resolved}
	pv := p.Preview

	title, err := pt.subst(pv.Title)
	if err != nil {
		return nil, fmt.Errorf("%s preview: %w", p.Slug, err)
	}

	fmt.Fprintf(&pt.b, `<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d" role="img" aria-label="%s themed with SP Night %s">`+"\n",
		width, height, width, height, escAttr(p.Name), escAttr(flavor.Label))
	fmt.Fprintf(&pt.b, "  <style>%s</style>\n", monoStyle)
	pt.drawWindow()

	var at layout
	switch pv.Kind() {
	case registry.FrameTerminal:
		pt.drawTitlebar(title)
		at = layout{bodyTop, bodyLeft}
	case registry.FrameEditor:
		pt.drawBufferline(title)
		at = layout{innerBodyTop, editorBodyLeft}
		if pv.CursorLine > 0 {
			pt.drawCursorLine(at.top + (pv.CursorLine-1)*lineHeight)
		}
		pt.drawGutter(len(pv.Body), pv.CursorLine)
	case registry.FrameApp:
		pt.drawHeader(title)
		at = layout{innerBodyTop, bodyLeft}
	case registry.FramePane:
		at = layout{paneBodyTop, bodyLeft}
	default:
		return nil, fmt.Errorf("%s preview: frame %q is not one of %s", p.Slug, pv.Frame, strings.Join(registry.Frames, ", "))
	}

	if err := pt.drawBody(pv.Body, at); err != nil {
		return nil, err
	}
	if pv.Kind() == registry.FrameEditor || pv.Kind() == registry.FrameApp {
		if err := pt.drawBar(pv.Bar); err != nil {
			return nil, err
		}
	}
	if err := pt.drawSwatches(pv.Swatches); err != nil {
		return nil, err
	}

	pt.b.WriteString("</svg>\n")
	return []byte(pt.b.String()), nil
}

// painter draws one preview into b, resolving roles against one flavour.
type painter struct {
	b        strings.Builder
	slug     string
	flavor   theme.Flavor
	resolved map[string]map[string]string
}

// look resolves a span or swatch reference: a role, or a raw palette key.
func (pt *painter) look(role, key string) (string, error) {
	switch {
	case role != "":
		group, name, ok := strings.Cut(role, ".")
		if !ok {
			return "", fmt.Errorf("role %q is not in group.role form", role)
		}
		hex, ok := pt.resolved[group][name]
		if !ok {
			return "", fmt.Errorf("role %q does not exist", role)
		}
		return hex, nil
	case key != "":
		hex, ok := pt.flavor.Colors[key]
		if !ok {
			return "", fmt.Errorf("palette colour %q does not exist", key)
		}
		return hex, nil
	default:
		return "", fmt.Errorf("neither a role nor a palette key")
	}
}

// role is look for a role the frame itself paints with, which cannot be
// mistyped by a catalogue entry.
func (pt *painter) role(name string) string {
	group, r, _ := strings.Cut(name, ".")
	return pt.resolved[group][r]
}

// subst resolves the placeholders a preview may use in its text.
//
// {r:ui.bg} and {c:laje} put the hex itself into the fake session, which is how
// a preview can show the value it is painting with — a mockup of
// `kitty @ get-colors` printing the real background is worth more than one
// printing an invented number.
func (pt *painter) subst(s string) (string, error) {
	s = strings.ReplaceAll(s, "{flavor}", pt.flavor.ID)
	s = strings.ReplaceAll(s, "{label}", pt.flavor.Label)

	var out strings.Builder
	for {
		open := strings.Index(s, "{")
		if open < 0 {
			out.WriteString(s)
			return out.String(), nil
		}
		close := strings.Index(s[open:], "}")
		if close < 0 {
			out.WriteString(s)
			return out.String(), nil
		}
		close += open

		kind, name, ok := strings.Cut(s[open+1:close], ":")
		if !ok || (kind != "r" && kind != "c") {
			// Not a reference. A literal brace in a shell snippet is
			// legitimate, so it is copied through.
			out.WriteString(s[:open+1])
			s = s[open+1:]
			continue
		}

		var hex string
		var err error
		if kind == "r" {
			hex, err = pt.look(name, "")
		} else {
			hex, err = pt.look("", name)
		}
		if err != nil {
			return "", fmt.Errorf("preview text {%s}: %w", s[open+1:close], err)
		}
		out.WriteString(s[:open])
		out.WriteString(hex)
		s = s[close+1:]
	}
}

// drawWindow is the window itself: the background and its border. Every frame
// starts here — a pane is nothing more than this.
func (pt *painter) drawWindow() {
	fmt.Fprintf(&pt.b, "  <rect width=\"%d\" height=\"%d\" rx=\"%d\" fill=\"%s\"/>\n",
		width, height, radius, pt.role("ui.bg"))
	fmt.Fprintf(&pt.b, "  <rect x=\"0.5\" y=\"0.5\" width=\"%d\" height=\"%d\" rx=\"%s\" fill=\"none\" stroke=\"%s\"/>\n",
		width-1, height-1, f1(radius-0.5), pt.role("ui.border"))
}

// drawTopStrip is a panel across the top, rounded to match the window, with a
// rule under it. The titlebar, the bufferline and the header are all this
// strip with different things on it.
func (pt *painter) drawTopStrip(h int) {
	fmt.Fprintf(&pt.b, "  <path d=\"M0 %da%d %d 0 0 1 %d-%dh%da%d %d 0 0 1 %d %dv%dH0z\" fill=\"%s\"/>\n",
		radius, radius, radius, radius, radius, width-2*radius, radius, radius, radius, radius,
		h-radius, pt.role("ui.panel"))
	fmt.Fprintf(&pt.b, "  <line x1=\"0\" y1=\"%s\" x2=\"%d\" y2=\"%s\" stroke=\"%s\"/>\n",
		f1(float64(h)+0.5), width, f1(float64(h)+0.5), pt.role("ui.border"))
}

// drawTitlebar is the terminal frame's top: three dots and a centred title.
func (pt *painter) drawTitlebar(title string) {
	pt.drawTopStrip(titlebarHeight)
	// The three window dots, drawn from the palette rather than a screenshot's
	// idea of what a window looks like.
	for i, key := range []string{"diagnostic.error", "diagnostic.warn", "diagnostic.ok"} {
		fmt.Fprintf(&pt.b, "  <circle cx=\"%d\" cy=\"%d\" r=\"%s\" fill=\"%s\"/>\n",
			24+i*18, dotY, f1(dotRadius), pt.role(key))
	}
	fmt.Fprintf(&pt.b, "  <text x=\"%d\" y=\"26\" text-anchor=\"middle\" class=\"mono\" fill=\"%s\">%s</text>\n",
		width/2, pt.role("ui.fg_muted"), escText(title))
}

// drawBufferline is the editor frame's top: one open buffer, its tab lifted to
// the background colour with an accent marker under it, the way an active tab
// joins the text below it.
func (pt *painter) drawBufferline(title string) {
	pt.drawTopStrip(stripHeight)
	tabW := 2*tabPad + tabCharWidth*utf8.RuneCountInString(title)
	fmt.Fprintf(&pt.b, "  <path d=\"M0 %da%d %d 0 0 1 %d-%dh%dv%dH0z\" fill=\"%s\"/>\n",
		radius, radius, radius, radius, radius, tabW-radius, stripHeight, pt.role("ui.bg"))
	fmt.Fprintf(&pt.b, "  <rect x=\"0\" y=\"%d\" width=\"%d\" height=\"%d\" fill=\"%s\"/>\n",
		stripHeight-tabMarker, tabW, tabMarker, pt.role("ui.accent"))
	fmt.Fprintf(&pt.b, "  <text x=\"%d\" y=\"%d\" class=\"mono\" fill=\"%s\">%s</text>\n",
		tabPad, stripTextY, pt.role("ui.fg"), escText(title))
}

// drawHeader is the app frame's top: the app's name, in the accent, where a
// full-screen program puts its own.
func (pt *painter) drawHeader(title string) {
	pt.drawTopStrip(stripHeight)
	fmt.Fprintf(&pt.b, "  <text x=\"%d\" y=\"%d\" class=\"mono\" fill=\"%s\" font-weight=\"bold\">%s</text>\n",
		stripPad, stripTextY, pt.role("ui.accent"), escText(title))
}

// drawCursorLine highlights the line whose baseline is at y. Drawn before the
// gutter and the text so both sit on top of it.
func (pt *painter) drawCursorLine(y int) {
	fmt.Fprintf(&pt.b, "  <rect x=\"1\" y=\"%d\" width=\"%d\" height=\"%d\" fill=\"%s\"/>\n",
		y-cursorLineRise, width-2, lineHeight, pt.role("ui.line"))
}

// drawGutter numbers n lines, brightening the cursor's, and rules the gutter
// off from the text. No block cursor: its column would need a guess at the
// font's advance, and the font is whatever the viewer has.
func (pt *painter) drawGutter(n, cursor int) {
	for i := 1; i <= n; i++ {
		fill := pt.role("ui.fg_muted")
		if i == cursor {
			fill = pt.role("ui.fg")
		}
		fmt.Fprintf(&pt.b, "  <text x=\"%d\" y=\"%d\" text-anchor=\"end\" class=\"mono\" fill=\"%s\">%d</text>\n",
			gutterNumX, innerBodyTop+(i-1)*lineHeight, fill, i)
	}
	fmt.Fprintf(&pt.b, "  <line x1=\"%s\" y1=\"%d\" x2=\"%s\" y2=\"%d\" stroke=\"%s\"/>\n",
		f1(gutterWidth+0.5), stripHeight+1, f1(gutterWidth+0.5), barY, pt.role("ui.border"))
}

// drawBody is the fake session. An empty line is a vertical gap, not a blank
// row: nothing is written for it, but it still takes its line's height.
func (pt *painter) drawBody(lines [][]registry.Span, at layout) error {
	for i, line := range lines {
		if len(line) == 0 {
			continue
		}
		if err := pt.writeSpans(at.left, at.top+i*lineHeight, "", line); err != nil {
			return err
		}
	}
	return nil
}

// drawBar is the strip along the bottom of the editor and app frames. A nil
// bar still draws the panel, so an editor without a statusline declared reads
// as an editor with an empty one rather than as a pane.
func (pt *painter) drawBar(bar *registry.Bar) error {
	fmt.Fprintf(&pt.b, "  <rect x=\"0\" y=\"%d\" width=\"%d\" height=\"%d\" fill=\"%s\"/>\n",
		barY, width, lineHeight, pt.role("ui.panel"))
	fmt.Fprintf(&pt.b, "  <line x1=\"0\" y1=\"%s\" x2=\"%d\" y2=\"%s\" stroke=\"%s\"/>\n",
		f1(barY+0.5), width, f1(barY+0.5), pt.role("ui.border"))
	if bar == nil {
		return nil
	}
	if len(bar.Left) > 0 {
		if err := pt.writeSpans(stripPad, barTextY, "", bar.Left); err != nil {
			return err
		}
	}
	if len(bar.Right) > 0 {
		if err := pt.writeSpans(width-stripPad, barTextY, "end", bar.Right); err != nil {
			return err
		}
	}
	return nil
}

// writeSpans is one line of coloured runs. anchor is "" for a line that starts
// at x, or an SVG text-anchor for one that ends there.
func (pt *painter) writeSpans(x, y int, anchor string, spans []registry.Span) error {
	if anchor == "" {
		fmt.Fprintf(&pt.b, "  <text x=\"%d\" y=\"%d\" xml:space=\"preserve\" class=\"mono\">", x, y)
	} else {
		fmt.Fprintf(&pt.b, "  <text x=\"%d\" y=\"%d\" text-anchor=\"%s\" xml:space=\"preserve\" class=\"mono\">", x, y, anchor)
	}
	for _, s := range spans {
		hex, err := pt.look(s.Role, s.Key)
		if err != nil {
			return fmt.Errorf("%s preview: %w", pt.slug, err)
		}
		bold := ""
		if s.Bold {
			bold = ` font-weight="bold"`
		}
		text, err := pt.subst(s.Text)
		if err != nil {
			return fmt.Errorf("%s preview: %w", pt.slug, err)
		}
		fmt.Fprintf(&pt.b, `<tspan fill="%s"%s>%s</tspan>`, hex, bold, escText(text))
	}
	pt.b.WriteString("</text>\n")
	return nil
}

// drawSwatches is the colour strip along the bottom, the same in every frame.
func (pt *painter) drawSwatches(sw registry.Swatches) error {
	label, err := pt.subst(sw.Label)
	if err != nil {
		return fmt.Errorf("%s preview: %w", pt.slug, err)
	}
	fmt.Fprintf(&pt.b, "  <text x=\"%d\" y=\"%d\" class=\"mono\" fill=\"%s\" font-size=\"12\">%s</text>\n",
		bodyLeft, swatchLabelY, pt.role("ui.fg_muted"), escText(label))

	refs := sw.Roles
	raw := len(refs) == 0
	if raw {
		refs = sw.Keys
	}
	n := len(refs)
	span := float64(width - 2*bodyLeft)
	w := (span - swatchGap*float64(n-1)) / float64(n)
	for i, ref := range refs {
		var hex string
		var err error
		if raw {
			hex, err = pt.look("", ref)
		} else {
			hex, err = pt.look(ref, "")
		}
		if err != nil {
			return fmt.Errorf("%s preview swatches: %w", pt.slug, err)
		}
		x := float64(bodyLeft) + float64(i)*(w+swatchGap)
		fmt.Fprintf(&pt.b, "  <rect x=\"%s\" y=\"%d\" width=\"%s\" height=\"%d\" rx=\"%d\" fill=\"%s\"/>\n",
			f1(x), swatchY, f1(w), swatchHeight, swatchRadius, hex)
	}
	return nil
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
