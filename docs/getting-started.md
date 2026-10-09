# Getting started

There are two ways in. Most people want a theme for an app they already use —
that needs no tool at all. The rest want to add a port or retune the palette,
and that is what `spn` is for.

## Use a theme

Every port is its own repository under [github.com/sp-night](https://github.com/sp-night),
and ships plain theme files with no build step. Installing one is always the
same three moves:

1. **Pick a flavour.** `noite` is the default, `garoa` is the flat grey one,
   `jaragua` turns the dark towards green. All three are dark.
2. **Copy the file** for that flavour from the port's `themes/` folder to the
   path the app reads.
3. **Activate it** with the one line the app expects.

Ghostty, for example:

```sh
mkdir -p ~/.config/ghostty/themes
curl -fsSL -o ~/.config/ghostty/themes/sp_night_noite \
  https://raw.githubusercontent.com/sp-night/ghostty/main/themes/sp_night_noite
echo 'theme = sp_night_noite' >> ~/.config/ghostty/config
```

Each port's README, and its page on [sp-night.github.io/ports](https://sp-night.github.io/ports),
gives the exact install path and activation line, a preview per flavour, and the
table of which key of the app gets which role. Those pages are generated from the
same catalogue as the theme files, so they cannot disagree with what you install.

## Install spn

You only need the engine to add a port, change a mapping, or retune the palette.

```sh
go install github.com/sp-night/sp-night/cmd/spn@latest
spn version
```

Release binaries for Linux and macOS (amd64 and arm64) are on the
[releases page](https://github.com/sp-night/sp-night/releases). The palette and
the port catalogue travel inside the binary, so `spn` works from any directory.

## Build from source

```sh
git clone https://github.com/sp-night/sp-night
cd sp-night
go build ./cmd/spn
go test ./...
go run ./cmd/spn check -v --palette palette
```

One dependency, for YAML. The test suite renders the mappings of ports that have
already shipped and compares the output against the files users installed, byte
for byte.

## Look at the palette

```sh
spn palette            # every colour, per flavour
spn palette --roles    # what each role resolves to
spn check -v           # the contrast audit, pair by pair
```

## Next

- Before writing or changing a mapping, read [the spec](SPEC.md).
- To add an app, follow [Adding a port](port-creation.md).
- Every command and flag is in the [command reference](reference/cli.md).
