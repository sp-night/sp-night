<p align="center">
  <a href="https://sp-night.github.io">
    <img src="https://raw.githubusercontent.com/sp-night/sp-night.github.io/main/public/logo-noite.svg" width="120" alt="SP Night — the Pico do Jaraguá at dusk, aviation beacon lit, the city's lights at the foot of the range">
  </a>
</p>

<h1 align="center">SP Night</h1>

<p align="center">
  <strong>The sodium lamp turns the whole city this colour.</strong><br>
  A dark colour scheme with São Paulo as its reference — the sodium street lamp,<br>
  exposed concrete, the free span of the MASP, the drizzle before the rain.
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

This repository is the **contract** and the **tool**.

The contract is two files: 23 named colours per flavour, and a role layer that
says which colour means what. The tool turns a port's mapping into finished
files, measures the palette against contrast floors, and refuses to build when it
falls short.

If you are here to *use* a theme, you want [the ports](https://sp-night.github.io/ports)
— each one ships plain text files with no build step. This repository is for
adding a port, or retuning the palette.

## Three layers, one direction

```
palette/sp_night.json     23 colours × 3 flavours — the only place a hex is written
        ↓
palette/roles.json        every role names a palette key, never a colour
        ↓
sp-night/<app>            a mapping asks for a role, never a colour
```

A worked example, from a Ghostty config key to a pixel:

```
palette = 4  →  ansi.blue  →  marginal  →  #6e92de
```

The middle step is the point. Move `syntax.keyword` from `marginal` to
`temporal` in one file, and Neovim, bat, fish, eza and the website all change
together. A mapping that wrote `marginal` directly would stay behind, and nobody
would notice for a year — which is exactly how a family of ports drifts apart.

`spn lint` fails on a mapping that reaches past the roles, so the rule is a gate
rather than a note in a style guide.

## Adding a port

Three steps. The mapping is the only one that needs a human.

```sh
# 1. list the target in registry/ports.yml — name, install path, role table
# 2. scaffold the repository
spn new kitty

# 3. write the mapping, then generate everything else
cd kitty
$EDITOR kitty.conf.tmpl
spn gen && spn readme && spn preview
```

The theme files, the README and the previews are all derived. What you write is
which key of the app means which role — and that mapping stays in the port
repository, because it is the only complete record of that decision.

Full walkthrough: [docs/port-creation.md](docs/port-creation.md).
What each role means: [docs/SPEC.md](docs/SPEC.md).

## What a port repository looks like

```
sp-night/ghostty
  ghostty.tmpl                  the mapping — the only hand-written file
  themes/                       generated, committed
  assets/preview-*.svg          generated, committed
  README.md                     generated, committed
  .github/workflows/theme.yml   six lines
```

Its entire CI:

```yaml
jobs:
  check:
    uses: sp-night/sp-night/.github/workflows/spn-check.yml@v1
    with:
      port: ghostty
```

Nothing there names an engine version. The mapping declares the range it is
written against in its own frontmatter:

```yaml
---
spn:
  version: "^1.0"
  matrix: [flavor]
  filename: "themes/sp_night_{{ .Flavor }}"
---
```

The workflow reads that and installs the right binary, and Renovate bumps the
constraint in the template — one pull request per port, none of them touching a
workflow file.

## The guarantees

**No hex is chosen by hand.** Not in a theme file, not in a README table, not in
a preview. The previews are synthetic SVGs drawn from the palette, so they cannot
show a colour the user will not get.

**Contrast is a gate, not a promise.** 70 pairs per flavour, measured every
build. The floor depends on what the surface is: `laje`, `vao` and `concreto`
demand AA 4.5:1 because that is where you read code for hours; `vidro` — the
selection — demands 3:1, because there you are looking at shape. Ornament and
borders are reported and allowed. An unreadable comment is the number-one issue
of every popular dark theme, and always because nobody measured.

**Accents are separated from each other.** Contrast against the background
cannot see this: two accents can clear AA comfortably and still be confused.
A perceptually neighbouring pair (ΔE < 0.10 in Oklab) has to separate by
lightness (ΔL ≥ 0.04).

**Colour vision is measured, not gated.** Protanopia, deuteranopia and
tritanopia are simulated and reported. Separating eight accents under colour
blindness, keeping the palette's character and clearing AA on a dark background
is an overdetermined system — no popular theme solves all three. The number
exists so the choice is deliberate.

**A published file is reproducible.** The test suite renders the mappings of the
already-shipped ports and compares against the files users installed, byte for
byte. If the engine stops being faithful, the suite says so before a release.

## The flavours

All three are dark, by decision — three ways of looking at the same city.

| id | label | idea |
|---|---|---|
| `noite` | Noite Paulista | the city at 3am; blue-violet dark, the sodium lamp burning warm on top |
| `garoa` | Garoa | the same window through the drizzle; flat grey, chroma near zero |
| `jaragua` | Pico do Jaraguá | the same night from the highest point; the dark turned towards the forest |

They are not one palette at three brightnesses. `noite` is cold and saturated,
`garoa` washes the city out rather than cooling it, `jaragua` rotates the hue
towards green at the same lightness and chroma. A new flavour has to differ by
**hue**, not only by brightness.

## Commands

```
spn gen        render a port's mapping into finished theme files
spn readme     render a port's canonical README from the catalogue
spn preview    draw the synthetic preview for each flavour
spn new        scaffold a new port repository
spn lint       check that mappings ask for roles, not raw colours
spn check      audit contrast, accent separation and colour vision
spn palette    print the palette and the resolved role layer
spn registry   list and validate the port catalogue
```

`--check` on `gen`, `readme` and `preview` writes nothing and fails when what is
committed is out of date. That is what a port's CI runs.

The palette travels inside the binary, so a port carries no copy of it. Pass
`--palette palette` to read this repository's copy instead — that is how the
palette gets retuned and measured before it ships.

## Layout

```
palette/          the contract: colours, roles, schema
registry/         the port catalogue, and the English prose every README shares
cmd/spn/          the command line
internal/color/   sRGB, WCAG, Oklab, Oklch, colour vision simulation
internal/theme/   loading, validation, role resolution
internal/audit/   contrast, accent separation, colour vision, KDE schemes
internal/render/  frontmatter, template helpers, the roles-only lint
internal/port/    README and preview generation, port scaffolding
docs/             the spec, and how to add a port
```

## Building it yourself

```sh
go build ./cmd/spn        # or: go install github.com/sp-night/sp-night/cmd/spn@latest
go test ./...
go run ./cmd/spn check -v --palette palette
```

One dependency, for YAML.

## References

The starting point conceptually was
[tokyodark.nvim](https://github.com/tiagovla/tokyodark.nvim) and
[Tokyo Night](https://github.com/folke/tokyonight.nvim). Palette-as-data with
templates comes from [Catppuccin](https://github.com/catppuccin/catppuccin),
whose [whiskers](https://github.com/catppuccin/whiskers) is the direct model for
keeping a port's mapping in the port's own repository. The separation between
surface and accent comes from [Rosé Pine](https://rosepinetheme.com). That ports
drift apart without a generator is the lesson of
[Dracula](https://github.com/dracula/dracula-theme), which has no engine and
several hundred repositories.

Palette, roles, skyline and names are original.

## License

[MIT](LICENSE)
