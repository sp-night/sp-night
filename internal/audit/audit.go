// Package audit measures the palette and fails the build when it falls short.
//
// The rule exists because every popular dark theme accumulates an "unreadable
// comment" issue, and always because nobody measured. Here the CI measures.
//
// Three measurements, deliberately different in kind:
//
//	contrast     WCAG 2.1 text/surface ratios — a gate
//	separation   accents against each other, in Oklab — a warning
//	cvd          the palette under colour vision deficiency — a diagnostic
//
// Only the first stops a build. The reasoning for that is in Separation and
// CVD below; it is not an oversight.
package audit

import (
	"fmt"
	"io"
	"math"
	"sort"
	"strings"

	"github.com/sp-night/sp-night/internal/color"
	"github.com/sp-night/sp-night/internal/theme"
)

// The level a pair must clear depends on what the surface *is*, not on an
// average:
//
//   - laje, vao, concreto — where you read code for hours (default background,
//     popup/float, cursorline). Accents and fg need AA 4.5:1.
//   - vidro — selection and visual mode. A transient state, and in it you are
//     looking at shape, not reading. 3:1 is enough, but it is required.
//   - fg_muted (line numbers, indent guides) is ornament: 3:1, warning only.
//   - fiacao is a border, not text: 1.5:1, warning only, and only against the
//     surfaces a border is actually drawn on.
const (
	LevelAA      = 4.5
	LevelLargeAA = 3.0
	LevelNonText = 1.5

	surfaceDeep   = "vao"
	surfaceMain   = "laje"
	surfacePanel  = "concreto"
	surfaceRaised = "vidro"
)

// textLevel is the floor for text on each surface.
var textLevel = map[string]float64{
	surfaceDeep:   LevelAA,
	surfaceMain:   LevelAA,
	surfacePanel:  LevelAA,
	surfaceRaised: LevelLargeAA,
}

// Surfaces are measured as backgrounds, deepest first.
var Surfaces = []string{surfaceDeep, surfaceMain, surfacePanel, surfaceRaised}

// borderSurfaces are the surfaces a border is really drawn against. Measuring
// fiacao against vidro — the selection colour — would not mean anything.
var borderSurfaces = []string{surfaceDeep, surfaceMain}

// Check is one measured text/surface pair.
type Check struct {
	FG, BG  string
	Ratio   float64
	Want    float64
	Fatal   bool
	Passing bool
}

// Pair is two accents that sit too close to be told apart.
type Pair struct {
	A, B      string
	Distance  float64
	Lightness float64
}

// CVDSummary is how the palette behaves for one colour vision deficiency.
type CVDSummary struct {
	Kind  string
	Tight int
	Total int
	Worst float64
}

// Report is everything measured for one flavour.
type Report struct {
	Flavor string
	Checks []Check
	Pairs  []Pair
	CVD    []CVDSummary
}

// Failures are the pairs that stop a build; Warnings are reported and allowed.
func (r Report) Failures() []Check { return r.filter(false, true) }
func (r Report) Warnings() []Check { return r.filter(false, false) }

func (r Report) filter(passing, fatal bool) []Check {
	var out []Check
	for _, c := range r.Checks {
		if c.Passing == passing && c.Fatal == fatal {
			out = append(out, c)
		}
	}
	return out
}

// Flavor measures every text/surface pair of one flavour.
func Flavor(pal *theme.Palette, f theme.Flavor) Report {
	rep := Report{Flavor: f.ID}

	add := func(fg, bg string, want float64, fatal bool) {
		ratio := color.Contrast(
			color.MustParseHex(f.Colors[fg]),
			color.MustParseHex(f.Colors[bg]),
		)
		rep.Checks = append(rep.Checks, Check{
			FG: fg, BG: bg, Ratio: ratio, Want: want,
			Fatal: fatal, Passing: ratio >= want,
		})
	}

	accents, brights := theme.AccentKeys(), theme.BrightKeys()

	for _, bg := range Surfaces {
		want := textLevel[bg]

		add("fg", bg, want, true)
		for _, a := range accents {
			add(a, bg, want, true)
		}
		for _, b := range brights {
			add(b, bg, want, true)
		}

		// Comments have to be comfortable on the default background and at
		// least distinguishable on the rest.
		if bg == surfaceMain {
			add("fg_dim", bg, LevelAA, true)
		} else {
			add("fg_dim", bg, LevelLargeAA, true)
		}

		add("fg_muted", bg, LevelLargeAA, false)
	}

	for _, bg := range borderSurfaces {
		add("fiacao", bg, LevelNonText, false)
	}

	sort.SliceStable(rep.Checks, func(i, j int) bool { return rep.Checks[i].Ratio < rep.Checks[j].Ratio })
	rep.Pairs = Separation(f)
	rep.CVD = CVD(f)
	return rep
}

// Separation between accents.
//
// Contrast against the background does not capture this: two accents can clear
// AA comfortably and still be confused with each other. That is how estaiada
// and sereno slipped past the first version of this audit — near-identical
// luminance on laje, separable by hue alone.
//
// ΔE alone is not a usable threshold. In an eight-accent palette covering the
// hue circle, neighbouring hues land naturally at ΔE 0.07–0.09; demanding more
// of every pair would be demanding a palette with fewer colours. Measured
// across the flavours, the genuinely confusable pairs are not the ones with
// the smallest ΔE — they are the ones with the smallest ΔL.
//
// So: a perceptually neighbouring pair (ΔE < NearHue) has to separate by
// lightness (ΔL >= MinLightness). That is what keeps the two distinguishable
// in greyscale — and, for the same reason, for someone with a colour vision
// deficiency.
const (
	NearHue      = 0.10
	MinLightness = 0.04
)

// Separation measures every accent against every other accent.
func Separation(f theme.Flavor) []Pair {
	accents := theme.AccentKeys()
	var out []Pair
	for i := range accents {
		for j := i + 1; j < len(accents); j++ {
			x := color.MustParseHex(f.Colors[accents[i]])
			y := color.MustParseHex(f.Colors[accents[j]])
			de := color.Distance(x, y)
			dl := math.Abs(x.Oklab().L - y.Oklab().L)
			if de < NearHue && dl < MinLightness {
				out = append(out, Pair{accents[i], accents[j], de, dl})
			}
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Lightness < out[j].Lightness })
	return out
}

// CVD summarises how the accents hold up under colour vision deficiency.
//
// This is a diagnostic, not a gate, and for a concrete reason: separating
// eight accents under colour blindness, keeping the palette's character, and
// clearing AA on a dark background is an overdetermined system — no popular
// theme solves all three. The number exists so the choice is deliberate, not
// so it blocks the build.
func CVD(f theme.Flavor) []CVDSummary {
	accents := theme.AccentKeys()
	total := len(accents) * (len(accents) - 1) / 2

	var out []CVDSummary
	for _, kind := range color.CVDKinds {
		s := CVDSummary{Kind: kind.Name, Total: total, Worst: math.Inf(1)}
		for i := range accents {
			for j := i + 1; j < len(accents); j++ {
				d := color.Distance(
					color.Simulate(color.MustParseHex(f.Colors[accents[i]]), kind),
					color.Simulate(color.MustParseHex(f.Colors[accents[j]]), kind),
				)
				if d < NearHue {
					s.Tight++
				}
				s.Worst = math.Min(s.Worst, d)
			}
		}
		out = append(out, s)
	}
	return out
}

// All measures every flavour and writes the report. The return value is the
// number of fatal pairs: zero means the palette may ship.
func All(pal *theme.Palette, w io.Writer, verbose bool) int {
	fmt.Fprintln(w, strings.Repeat("─", 66))
	fmt.Fprintln(w, "WCAG 2.1 contrast · accent separation in Oklab · colour vision")

	var reps []Report
	for _, f := range pal.Flavors {
		reps = append(reps, Flavor(pal, f))
	}
	n := WriteReport(w, reps, verbose)
	fmt.Fprintln(w, strings.Repeat("─", 66))
	return n
}

// WriteReport prints the measurements and returns the fatal count.
func WriteReport(w io.Writer, reps []Report, verbose bool) (fatal int) {
	for _, rep := range reps {
		fails, warns := rep.Failures(), rep.Warnings()
		fatal += len(fails)

		status := "ok"
		if len(fails) > 0 {
			status = fmt.Sprintf("%d FAILED", len(fails))
		}
		fmt.Fprintf(w, "\n  %s (%d pairs, %s", rep.Flavor, len(rep.Checks), status)
		if n := len(warns) + len(rep.Pairs); n > 0 {
			fmt.Fprintf(w, ", %d warning(s)", n)
		}
		fmt.Fprintln(w, ")")

		show := func(prefix string, cs []Check) {
			for _, c := range cs {
				fmt.Fprintf(w, "    %s %-9s on %-9s %5.2f:1  (floor %.1f)\n",
					prefix, c.FG, c.BG, c.Ratio, c.Want)
			}
		}
		show("✗", fails)
		show("!", warns)

		for _, p := range rep.Pairs {
			fmt.Fprintf(w, "    ! %-9s and %-9s ΔL %.3f  (floor %.2f) — neighbouring hues, same lightness\n",
				p.A, p.B, p.Lightness, MinLightness)
		}

		if verbose {
			for _, s := range rep.CVD {
				fmt.Fprintf(w, "    · %-13s %2d/%d pairs close, worst %.3f\n",
					s.Kind, s.Tight, s.Total, s.Worst)
			}
			for _, c := range rep.Checks {
				if c.Passing {
					fmt.Fprintf(w, "    ✓ %-9s on %-9s %5.2f:1\n", c.FG, c.BG, c.Ratio)
				}
			}
		}
	}
	return fatal
}
