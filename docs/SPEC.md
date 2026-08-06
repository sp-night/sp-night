# The SP Night spec

This document exists to answer one question: **when I port the theme to a new
platform, which colour goes where?** Without a written answer every port becomes
a different interpretation of the theme — that is how the Dracula ports drifted
apart from each other.

The same rules are published at [sp-night.github.io/spec](https://sp-night.github.io/spec).
This copy is the one the tool enforces.

## The three layers

```
palette/sp_night.json     23 colours per flavour. São Paulo names.
        ↓
palette/roles.json        semantic roles → a palette key
        ↓
<app>.tmpl                asks for a role, never the raw palette
```

**Hard rule: a mapping never references a palette colour directly.**

```
✗  palette = 4={{ .C.marginal }}
✓  palette = 4={{ .R.ansi.blue }}
```

Why: moving `syntax.keyword` from `marginal` to `temporal` has to repaint Neovim,
bat, fish and the website at once. A mapping with `marginal` written into it
stays behind and nobody notices.

`spn lint` fails on the first form. It walks the parse tree, not the text, so a
counter-example inside a comment is not a finding.

The exception is the variable lists — `waybar.css`, `gtk.css`, `hyprland.conf` —
which publish the raw palette as `@define-color` / `$var` because the end user
will want to write `@sp_sodio` in their own stylesheet. Those declare
`raw_palette: true` in their frontmatter. Even there, everything the template
itself styles still uses a role.

## The palette

### Surfaces — names neutral about lightness

These names describe **function** and depth, not a specific hex. That is what
lets one mapping serve every flavour.

| key | function | reference |
|---|---|---|
| `vao` | the deepest recess: float, popup, tab bar | the free span of the MASP |
| `laje` | **the default background** — the plane everything sits on | a concrete slab |
| `concreto` | panels, cards, cursorline, statusline | exposed concrete |
| `vidro` | selection, visual mode, active item | glass reflecting the street |
| `fiacao` | borders, dividers, indent guides | overhead wiring |

### Text

| key | use |
|---|---|
| `fg_vivo` | bold default text, `ansi.bright_white` |
| `fg` | main text, identifiers, variables |
| `fg_dim` | comments, punctuation, secondary text |
| `fg_muted` | line numbers, guides, disabled text |

### Accents

| key | ANSI | reference |
|---|---|---|
| `brasa` | red | the MASP, brake lights |
| `sodio` | — | the sodium street lamp |
| `taxi` | yellow | traffic light, taxi |
| `ibira` | green | Ibirapuera, a green light |
| `estaiada` | — | the Estaiada bridge lit up |
| `sereno` | cyan | the dew before dawn |
| `marginal` | blue | the Marginal, the metro |
| `temporal` | magenta | the sky just before the rain |

### Bright ANSI

The six terminal brights (`brasa_vivo`, `taxi_vivo`, `ibira_vivo`,
`sereno_vivo`, `marginal_vivo`, `temporal_vivo`) are the same colour with +0.06
Oklch lightness — hue and chroma untouched. Bold text in a terminal uses bright;
if bright equals normal, that information disappears. They live in the palette
rather than being derived in a template, so they pass the same contrast audit as
everything else.

`fg_vivo` is that same lift applied to the text ramp, and it exists for the same
reason. A target that separates default text from **bold** default text —
Alacritty's `colors.primary.bright_foreground`, kitty's bold font colour — had
nothing above `ui.fg` to point at, so a mapping had to write `fg` into both keys
and claim a lift that was not there. `ui.fg_bright` names it, and
`ansi.bright_white` resolves to it. `ansi.white` stays `fg_dim`: the ANSI table
keeps the two ends of the ramp, and `fg` remains the terminal's plain
`foreground`.

For the text ramp +0.06 is the ceiling, not an inherited constant. `fg` in
`noite` already sits at Oklch L 0.88; at +0.07 the blue channel clips and chroma
starts to fall (0.028 → 0.023). At +0.06 all three flavours land the lift exactly
with chroma intact — measured, which is what decided the value.

`sodio` is the signature colour: cursor, active border, active workspace, clock,
title. If a port needs "the theme's colour", that is it — **in a terminal or a
bar**. In an application widget (selection and focus in GTK/Qt/KDE) the accent is
`ui.accent_alt`, blue: an orange system accent makes a whole file manager look
like a warning. Tokyo Dark makes the same split, and its primary was always blue.

`sodio` and `estaiada` have no ANSI slot — the 16 slots already have owners.
They exist for the interface, not the terminal.

## Assignment rules

**Syntax** — the logic is: *what the code does* gets blue-cyan, *what the code
is* gets purple-orange, *literal data* gets green-orange.

| role | colour |
|---|---|
| keyword, conditional, repeat | `marginal` |
| function, method, operator | `sereno` |
| type, namespace | `temporal` |
| constant, number, boolean | `sodio` |
| string, character | `ibira` |
| parameter, macro, escape | `estaiada` |
| builtin, attribute | `taxi` |
| tag | `brasa` |
| variable, property, field | `fg` |
| comment, punctuation | `fg_dim` |

Variables stay `fg` on purpose. A theme where everything is coloured has no
hierarchy — the eye needs a rest, and the ordinary identifier is the largest
volume of text on screen.

**Diagnostics and git** follow the universal convention and must not be
reinterpreted: error = red, warning = yellow, info = blue, hint = cyan,
ok/added = green, modified = yellow, removed = red.

**A diff has one pattern**, in every target: added = `ibira`, modified and
renamed = `taxi`, removed and conflict = `brasa`. All the structure — hunk,
header, index, untracked — is neutral (`fg_dim` / `fg_muted`). A diff that also
uses purple, blue and orange becomes a rainbow where only three states matter.

## Surface temperature

Blue-yellow opposition is the strongest axis a theme has, and in a dark theme it
does **not** get resolved in the accents: blue has intrinsically low luminance,
so a blue that clears AA on `concreto` has to be light and desaturated — the
opposite of a pole. Measured, repainting the accents moves the axis from 1.03 to
1.35 and costs the character of the whole palette.

It gets resolved in the **background**, which is not text and therefore has no
contrast constraint. `noite` has a blue-violet `laje` at chroma 0.037; the accents
stay warm. The tension between field and accent mass goes from 0.020 to 0.051
without changing a single accent.

Hence the split between the two dark flavours: `noite` is cold and **saturated**,
`garoa` is **flat** grey (chroma 0.005). The garoa does not cool the city down,
it washes it out — the distinction is saturation, not temperature.

## The lesson of the robust themes (KDE)

What makes Breeze Dark and Catppuccin work in any app is not the palette — it is
the discipline of the scheme:

1. **One single set of foregrounds, repeated literally** across View, Window,
   Button, Tooltip, Complementary and Header. Nothing is derived per section;
   consistency is the absence of derivation.
2. **The Selection section is self-contained.** Every foreground in it is chosen
   to read *on the accent*, and its `BackgroundAlternate` is a neighbour of the
   accent — never the general alternate. That substitution is what produced
   1.11:1 in Dolphin.
3. **`[WM]` is dark and calm.** A title bar never takes the accent.

And one none of them has: the tool parses the `.colors` file it has just built
and fails the build if any fg/bg pair of any section falls below 4.5:1
(`internal/audit/kde.go`). The real Breeze has pairs at 2–3:1 in Selection;
SP Night does not compile with one.

## Contrast

`spn check` measures every text/surface pair of every flavour and **fails the
build** when one falls short. The floor depends on what the surface is:

| surface | floor | why |
|---|---|---|
| `laje`, `vao`, `concreto` | 4.5:1 (AA) | where you read code for hours |
| `vidro` | 3.0:1 | selection: a transient state, you look at shape |
| `fg_muted` on any | 3.0:1, warning | ornament, not reading text |
| `fiacao` on `laje`/`vao` | 1.5:1, warning | a border, not text |

74 pairs per flavour. A gate that fails stops the build; a warning is reported.

`fg_dim` is the delicate case: comments need full AA on `laje`. An unreadable
comment is the number-one issue of every popular dark theme, and always because
nobody measured.

### Separation between accents

Contrast against the background is not enough. Two accents can clear AA
comfortably and still be confused **with each other** — that is how `estaiada`
and `sereno` slipped past the first version of this audit: near-identical
luminance on `laje`, separable by hue alone.

ΔE alone is not a usable threshold. In an eight-accent palette covering the hue
circle, neighbouring hues land naturally at ΔE 0.07–0.09 in Oklab; demanding more
of every pair would be demanding a palette with fewer colours. Measured across
the flavours, the genuinely confusable pairs are not the ones with the smallest
ΔE — they are the ones with the smallest **ΔL**.

The rule: a perceptually neighbouring pair (ΔE < 0.10) has to separate by
lightness (ΔL ≥ 0.04). That is what keeps the two distinguishable in greyscale —
and, for the same reason, for someone with a colour vision deficiency.

### Colour vision deficiency

`spn check -v` also simulates protanopia, deuteranopia and tritanopia (Viénot
1999 matrices) and reports how many accent pairs sit too close for each.

This is a **diagnostic, not a gate**, for a concrete reason: separating eight
accents under colour blindness, keeping the palette's smoky character, and
clearing AA on a dark background is an overdetermined system. A search over hue,
chroma and lightness finds solutions with much better separation, but all of them
get there by moving the theme somewhere else — yellow drops to olive, red becomes
terracotta, the cyans hit the edge of the gamut and turn neon. No popular theme
solves all three; Catppuccin, Tokyo Night and Dracula do not either.

The number exists so the choice is deliberate. If the priority ever changes, the
tool already measures it.

### About the ΔL rule

It has a blind spot worth knowing: it only examines pairs of neighbouring hue in
normal vision. Under deuteranopia the pairs that collapse are the *distant* ones —
red and green land in the same place. One does not cover the other, which is why
both measurements coexist.

It is a warning, not a failure: these are the theme's identity colours, and the
decision to repaint them is the maintainer's. The accents' lightness ladder keeps
the list empty today.

Run `spn check` before touching a colour.

## Adding a flavour

1. Add the block to `palette/sp_night.json` → `flavors`.
2. `spn check`.

That is all — order comes from the JSON itself, so there is no second list in Go
to keep in step.

A new flavour has to distinguish itself from the existing ones by **hue**, not
only by lightness. `garoa` and `noite` are the example: the first version of
garoa was noite with the blacks raised, and of ΔE 0.034 between the surfaces,
0.033 was ΔL — the same colour, lighter. Pulling the grey towards blue (Oklab B
from −0.005 to −0.017) raised Δchroma 7× and dropped ΔL: they became two
atmospheres rather than two brightnesses.

No mapping changes. If a mapping has to change to accommodate a new flavour, that
is a sign it is using the raw palette where it should use a role.

## Adding a target

1. List it in `registry/ports.yml`.
2. `spn new <slug>`.
3. Write the mapping using `.R`.

The frontmatter accepts an empty `matrix` to render once — for an app that reads a
single fixed config file. Formats that need a light block alongside the dark one
mirror the dark side, since the theme is dark-only; `.All` reaches the other
flavours: `{{ (index .All "garoa").R.ui.bg }}`.

For "text that reads on this accent", use `readable`, which picks between two
options by measured contrast instead of guessing:

```
"mOnPrimary": "{{ readable .C.vao .C.fg .R.ui.accent }}"
```

Helpers available in a mapping: `nohash`, `rgb`, `rgbn`, `rgba`, `hexa`, `argb`,
`sgrfg`, `sgrbg`, `mix`, `lighten`, `darken`, `contrast`, `readable`, `kebab`,
`pad`, `repeat`, `tojson`, `upper`, `lower`, `r`, `g`, `b`.
