package port

import (
	"fmt"
	"strings"
)

// Entry returns a complete, paste-ready catalogue entry for a new port.
//
// It exists because of what the catalogue actually costs to write. Measured
// across the four shipped ports, `preview.body` is 54 to 203 lines of an entry
// that runs 131 to 278 — while `mapping`, the one part that is a decision
// somebody makes, is 18 to 34. The blank page was in the wrong place: most of
// the effort of adding a port went into hand-writing a fake terminal session,
// span by span, and none of that is a judgement about the port.
//
// So the session is written here once, in the shape the four existing previews
// already share, and every new port starts from it. What comes out is valid the
// moment it is pasted — `spn registry` passes on the placeholders — so the
// entry can be checked in as it is filled in rather than only at the end.
//
// This is a starting point, not a generated file: the entry lives in the
// catalogue and belongs to the port from then on. The previews that carry a
// port's own voice, eza's especially, are what happens when someone rewrites
// this body, and they should keep rewriting it.
func Entry(slug string) string {
	var b strings.Builder

	fmt.Fprintf(&b, "  - slug: %s\n", slug)
	fmt.Fprintf(&b, "    name: %s                       # TODO display name, as the project spells it\n", slug)
	fmt.Fprintf(&b, "    group: terminal                # TODO a key of `groups` at the top of this file\n")
	fmt.Fprintf(&b, "    blurb: TODO one line on what the port covers.\n")
	fmt.Fprintf(&b, "    homepage: https://example.com  # TODO the themed app's own site\n")
	fmt.Fprintf(&b, "    repo: https://github.com/sp-night/%s\n", slug)
	fmt.Fprintf(&b, "    install: ~/.config/%s/sp_night_{flavor}.conf   # TODO where the file goes\n", slug)
	fmt.Fprintf(&b, "    activate: TODO the line the user adds, or drop this field\n")
	fmt.Fprintf(&b, "    template: %s.conf.tmpl\n", slug)
	fmt.Fprintf(&b, "\n")
	fmt.Fprintf(&b, "    install_guide: |-\n")
	fmt.Fprintf(&b, "      TODO the app-specific prose of the README's Install section.\n")
	fmt.Fprintf(&b, "\n")
	fmt.Fprintf(&b, "      ```sh\n")
	fmt.Fprintf(&b, "      mkdir -p ~/.config/%s\n", slug)
	fmt.Fprintf(&b, "      curl -Lo ~/.config/%s/sp_night_noite.conf \\\n", slug)
	fmt.Fprintf(&b, "        https://raw.githubusercontent.com/sp-night/%s/main/themes/sp_night_noite.conf\n", slug)
	fmt.Fprintf(&b, "      ```\n")
	fmt.Fprintf(&b, "\n")

	// The mapping rows and the preview body below are deliberately consistent:
	// the session shows the same keys the table names. Editing one without the
	// other is the drift this whole project exists to prevent, so they start
	// out agreeing.
	fmt.Fprintf(&b, "    mapping:\n")
	fmt.Fprintf(&b, "      - key: \"`background` / `foreground`\"\n")
	fmt.Fprintf(&b, "        role: \"`ui.bg` / `ui.fg`\"\n")
	fmt.Fprintf(&b, "        meaning: \"*laje* under the main text\"\n")
	fmt.Fprintf(&b, "      - key: \"`cursor`\"\n")
	fmt.Fprintf(&b, "        role: \"`ui.cursor`\"\n")
	fmt.Fprintf(&b, "        meaning: the *sódio* cursor\n")
	fmt.Fprintf(&b, "      # TODO one row per key this port themes. This table is what a reader\n")
	fmt.Fprintf(&b, "      # checks before trusting the port, so it is required.\n")
	fmt.Fprintf(&b, "\n")

	fmt.Fprintf(&b, "    preview:\n")
	fmt.Fprintf(&b, "      title: %s — sp_night_{flavor}\n", slug)
	fmt.Fprintf(&b, "      swatches:\n")
	fmt.Fprintf(&b, "        label: palette 0–15          # TODO name the strip below\n")
	fmt.Fprintf(&b, "        roles:\n")
	for _, r := range ansiStrip {
		fmt.Fprintf(&b, "          - %s\n", r)
	}
	fmt.Fprintf(&b, "        # TODO a non-terminal port wants the roles it actually paints with\n")
	fmt.Fprintf(&b, "        # here instead of the ANSI ladder. `spn palette --roles` lists them.\n")
	fmt.Fprintf(&b, "      body:\n")
	b.WriteString(previewBody(slug))

	return b.String()
}

// ansiStrip is the 16-colour ladder, the default swatch row. Every terminal
// port shows it, and a terminal is what most ports are.
var ansiStrip = []string{
	"ansi.black", "ansi.red", "ansi.green", "ansi.yellow",
	"ansi.blue", "ansi.magenta", "ansi.cyan", "ansi.white",
	"ansi.bright_black", "ansi.bright_red", "ansi.bright_green", "ansi.bright_yellow",
	"ansi.bright_blue", "ansi.bright_magenta", "ansi.bright_cyan", "ansi.bright_white",
}

// previewBody is the fake session every port starts with: the shape the four
// shipped previews already converged on. A prompt showing the port's own
// install, a block that exercises the syntax roles so the preview shows more
// than two colours, and the audit's own output — which is true, and is the
// claim the project most wants a reader to see.
func previewBody(slug string) string {
	var b strings.Builder
	line := func(spans ...string) {
		b.WriteString("        - -")
		for i, s := range spans {
			if i > 0 {
				b.WriteString("         ")
			}
			fmt.Fprintf(&b, " %s\n", s)
		}
	}
	gap := func() { b.WriteString("        - []\n") }

	line(
		`{t: "~/sp-night ", r: ui.accent}`,
		`- {t: "(main) ", r: ui.match}`,
		`- {t: "❯ ", r: ui.accent_alt}`,
		fmt.Sprintf(`- {t: "cat ~/.config/%s/sp_night_{flavor}.conf", r: ui.fg}`, slug),
	)
	line(`{t: "# TODO a line or two of this port's real config format", r: syntax.comment}`)
	line(
		`{t: "background", r: syntax.attribute}`,
		`- {t: " = ", r: syntax.operator}`,
		`- {t: "{r:ui.bg}", r: syntax.string}`,
	)
	line(
		`{t: "foreground", r: syntax.attribute}`,
		`- {t: " = ", r: syntax.operator}`,
		`- {t: "{r:ui.fg}", r: syntax.string}`,
	)
	gap()
	line(
		`{t: "~/sp-night ", r: ui.accent}`,
		`- {t: "❯ ", r: ui.accent_alt}`,
		`- {t: "spn check", r: ui.fg}`,
	)
	line(
		`{t: "✓ ", r: diagnostic.ok}`,
		`- {t: "contrast   ", r: ui.fg}`,
		`- {t: "70 pairs per flavour, AA on every surface", r: ui.fg_dim}`,
	)
	line(
		`{t: "✓ ", r: diagnostic.ok}`,
		`- {t: "accents    ", r: ui.fg}`,
		`- {t: "separation ΔE ≥ 0.10 or ΔL ≥ 0.04", r: ui.fg_dim}`,
	)
	return b.String()
}
