# Adding a port

A port is a **mapping**, not a copy of colours: it decides which key of an app's
config gets which role, and the tool resolves the rest. Read
[SPEC.md](SPEC.md) first — in particular the rule that a mapping never
references the raw palette.

The canonical example is [`sp-night/ghostty`](https://github.com/sp-night/ghostty).

## Non-negotiables

1. **No hex is chosen by hand.** Not in a theme file, not in a README, not in a
   preview. If something needs to change, the mapping or the catalogue changes and
   everything is regenerated.
2. **Conventional Commits**, in English, lower case, no trailing period:
   `feat(themes):`, `docs(readme):`, `feat(ports):`, `ci:`, `test:`, `chore:`.
3. **Never commit to `main` directly.** Branch → pull request → merge.
4. **Public docs in English.** Colour and flavour names stay Portuguese — they
   are the identity.
5. **MIT** in every repository.

## 1 — List the target in the catalogue

`registry/ports.yml` is what the README, the preview and the generated header all
read from. Nothing can be generated for a port that is not listed.

Start from a printed entry rather than a blank page:

```sh
spn entry <slug> >> registry/ports.yml
```

That writes every required field, a mapping table to extend, and a complete
preview — the fake session, span by span, in the shape the shipped previews
share. It validates as printed, so the entry can be checked in while the TODOs
are still being replaced. The preview is a starting point and nothing more:
rewrite the session once the port has a voice of its own, the way eza's did.

The fields it leaves you:

```yaml
  - slug: kitty                 # also the repository name under the org
    name: kitty
    group: terminal
    blurb: One line on what the port covers.
    homepage: https://sw.kovidgoyal.net/kitty/
    repo: https://github.com/sp-night/kitty     # derived from slug, and checked
    install: ~/.config/kitty/sp_night_{flavor}.conf
    activate: include sp_night_{flavor}.conf
    template: kitty.conf.tmpl
    mapping:
      - key: "`background`"
        role: "`ui.bg`"
        meaning: "*laje* under the main text"
    preview:
      title: kitty — sp_night_{flavor}
      swatches: {label: palette 0–15, roles: [ansi.black, ...]}
      body:
        - - {t: "~/sp-night ", r: ui.accent}
          - {t: "❯ ", r: ui.accent_alt}
          - {t: "kitty +kitten themes --dump-theme", r: ui.fg}
```

Then check it:

```sh
spn registry --registry registry/ports.yml --copy registry/copy.yml
```

`mapping` is not decoration — it is the table a reader checks before trusting a
port, and validation refuses an entry without one.

## 2 — Create the repository

```sh
gh repo create sp-night/<slug> --public --license mit \
  --description "🌃 SP Night for <App> — a dark colour scheme with São Paulo as its reference"
```

The initial commit carries only `LICENSE`. Clone it next to the others, then
scaffold:

```sh
spn new <slug>
```

That writes the mapping stub with the canonical header already wired to the
catalogue, the six-line workflow, the Renovate config, and the editor settings.

## 3 — Write the mapping

This is the only file a human writes. Ask for a role, never a colour:

```
✗  background {{ .C.laje }}
✓  background {{ .R.ui.bg }}
```

`spn palette --roles` lists every role and what it resolves to in each flavour.
What each one *means* is in [SPEC.md](SPEC.md#assignment-rules).

The frontmatter declares the contract:

```yaml
---
spn:
  version: "1.2.0"                                   # the exact engine this was generated with
  matrix: [flavor]                                   # one file per flavour
  filename: "themes/sp_night_{{ .Flavor }}.conf"     # where it goes
---
```

Leave `matrix` out for an app that reads one fixed config file — eza is the case:
its mapping renders once, and each flavour still ships as a file in `themes/`.

`version` is an **exact** version, not a range, and nobody edits it by hand.
`spn new` writes the version of the binary that scaffolded the port, and from
then on the engine's release fan-out bumps it in the same pull request that
regenerates the theme files.

That pairing is the point. The committed files and the version that produced them
are one fact, so a port's CI installs the pinned engine and checks the files
against it. A range — `^1.0` was the original — makes the CI resolve to whatever
is newest, so a port's pull request goes from green to red because *the engine*
released, with nothing in the port having changed. It also made the Renovate rule
that was supposed to propagate engine bumps inert: every `1.x` satisfies `^1.0`,
so there was never anything to bump.

Then generate everything else:

```sh
spn gen && spn readme && spn preview
spn lint && spn gen --check
```

## 4 — Commit history

On branch `feat/<slug>-port`, mirroring ghostty:

```
feat(themes): add the noite, garoa and jaragua flavours
docs(readme): add install guide and palette-generated previews
ci: check the mapping and the generated files
```

Then `gh pr create`, and merge through the pull request.

Before pushing, confirm the generated header points at this port's repository and
carries the install path and the activation line. If it does not, the catalogue
entry is wrong — do not edit a generated file.

## 5 — The website, which is not a step

There is nothing to do here, and nothing to edit. The website vendors this
catalogue rather than keeping its own list, so merging step 1 to `main` is what
publishes the port: the push triggers `sync-ports.yml`, whose `site` job copies
`ports.yml` and `copy.yml` across, redraws the previews, regenerates the site's
derived assets, runs its suite, and opens a pull request. Merging that deploys
`/ports/<slug>` — a page built from the entry, with the install guide, the
key-to-role table and a screenshot per flavour.

Listing a port is publishing it. If you find yourself editing `src/data/` by
hand, the change will disappear at the next sync without leaving a trace.

## Checklist

- [ ] Listed in `registry/ports.yml` (start with `spn entry <slug>`), `spn registry` clean
- [ ] `spn lint` clean — the mapping asks for roles, never colours
- [ ] `spn gen --check`, `spn readme --check`, `spn preview --check` all clean
- [ ] Three flavours present (noite, garoa, jaragua)
- [ ] Generated header points at this port's repository
- [ ] `.github/workflows/theme.yml` calls the reusable workflow
- [ ] Branch `feat/<slug>-port`, Conventional Commits, merged via pull request
- [ ] Catalogue merged to `main`, and the website's sync pull request merged

## When a port needs something the roles do not offer

That is a real finding, not a reason to reach for the palette. A missing role
means the semantic layer has a hole — open an issue on this repository describing
the key you could not map. Adding a role repaints every port that wants it;
writing a colour into one mapping helps exactly one port and hides the gap.
