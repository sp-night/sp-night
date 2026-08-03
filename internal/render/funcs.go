package render

import (
	"encoding/json"
	"fmt"
	"strings"
	"text/template"

	"github.com/sp-night/sp-night/internal/color"
)

// Funcs are the helpers available in a template.
//
// They exist so a mapping never has to compute a colour by hand. Every app
// wants the same colours in a different notation — bare hex, "r, g, b", an SGR
// escape, alpha at the front or the back — and a port should be spelling out
// which role a key gets, not converting between formats.
//
// The hex arguments come from .R, which Palette.Validate has already proved
// parseable, so a failure here means a template passed something that is not a
// colour at all. Failing loudly beats emitting a silently wrong file.
var Funcs = template.FuncMap{
	// --- notation ---
	"nohash": func(s string) string { return strings.TrimPrefix(s, "#") },
	"upper":  strings.ToUpper,
	"lower":  strings.ToLower,

	"r": func(s string) int { return int(color.MustParseHex(s).R) },
	"g": func(s string) int { return int(color.MustParseHex(s).G) },
	"b": func(s string) int { return int(color.MustParseHex(s).B) },

	// "18, 16, 17"
	"rgb": func(s string) string {
		c := color.MustParseHex(s)
		return fmt.Sprintf("%d, %d, %d", c.R, c.G, c.B)
	},
	// "18,16,17" — no spaces (kdeglobals, kcolorscheme)
	"rgbn": func(s string) string {
		c := color.MustParseHex(s)
		return fmt.Sprintf("%d,%d,%d", c.R, c.G, c.B)
	},
	// "rgba(18, 16, 17, 0.50)"
	"rgba": func(a float64, s string) string {
		c := color.MustParseHex(s)
		return fmt.Sprintf("rgba(%d, %d, %d, %.2f)", c.R, c.G, c.B, color.Clamp01(a))
	},
	// "#121011cc" — alpha last (CSS, Ghostty)
	"hexa": func(a float64, s string) string {
		return fmt.Sprintf("%s%02x", s, int(color.Clamp01(a)*255+0.5))
	},
	// "cc121011" — alpha first (Hyprland, Android)
	"argb": func(a float64, s string) string {
		return fmt.Sprintf("%02x%s", int(color.Clamp01(a)*255+0.5), strings.TrimPrefix(s, "#"))
	},
	// "38;2;18;16;17" — truecolor SGR foreground (LS_COLORS, EZA_COLORS)
	"sgrfg": func(s string) string {
		c := color.MustParseHex(s)
		return fmt.Sprintf("38;2;%d;%d;%d", c.R, c.G, c.B)
	},
	// "48;2;18;16;17" — truecolor SGR background
	"sgrbg": func(s string) string {
		c := color.MustParseHex(s)
		return fmt.Sprintf("48;2;%d;%d;%d", c.R, c.G, c.B)
	},

	// --- manipulation ---
	"mix": func(t float64, a, b string) string {
		return color.Mix(color.MustParseHex(a), color.MustParseHex(b), t).Hex()
	},
	"lighten": func(t float64, s string) string { return color.Lighten(color.MustParseHex(s), t).Hex() },
	"darken":  func(t float64, s string) string { return color.Darken(color.MustParseHex(s), t).Hex() },

	// WCAG contrast, for a template that wants to state the ratio it relies on
	"contrast": func(a, b string) string {
		return fmt.Sprintf("%.2f", color.Contrast(color.MustParseHex(a), color.MustParseHex(b)))
	},
	// Returns whichever of dark/light reads better on s — so "text on this
	// accent" is measured rather than guessed.
	"readable": func(dark, light, s string) string {
		c := color.MustParseHex(s)
		if color.Contrast(c, color.MustParseHex(dark)) >= color.Contrast(c, color.MustParseHex(light)) {
			return dark
		}
		return light
	},

	// --- text ---
	"tojson": func(v any) (string, error) {
		b, err := json.MarshalIndent(v, "", "  ")
		return string(b), err
	},
	"repeat": strings.Repeat,
	"pad": func(n int, s string) string {
		if len(s) >= n {
			return s
		}
		return s + strings.Repeat(" ", n-len(s))
	},
	// fg_dim -> fg-dim: palette keys are snake_case, CSS is kebab
	"kebab": func(s string) string { return strings.ReplaceAll(s, "_", "-") },
}
