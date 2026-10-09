# SP Night

**The sodium lamp turns the whole city this colour.** SP Night is a dark colour
scheme with São Paulo as its reference — the sodium street lamp, exposed
concrete, the free span of the MASP, the drizzle before the rain.

This repository is the **contract** and the **tool**. The contract is a palette
and a role layer that says which colour means what. The tool, `spn`, turns a
port's mapping into finished theme files, measures the palette against contrast
floors, and refuses to build when it falls short.

## Where to start

- [Getting started](getting-started.md): pick a flavour, install a port, or
  build `spn` yourself.
- [The spec](SPEC.md): the palette, the roles, and the rule for which colour goes
  where. Read it before touching a mapping.
- [Adding a port](port-creation.md): from a catalogue entry to a published
  repository, step by step.
- [Command reference](reference/cli.md): every `spn` command and flag.

## The three flavours

All three are dark, by decision — three ways of looking at the same city.

| id | label | idea |
|---|---|---|
| `noite` | Noite Paulista | the city at 3am; blue-violet dark, the sodium lamp burning warm on top |
| `garoa` | Garoa | the same window through the drizzle; flat grey, chroma near zero |
| `jaragua` | Pico do Jaraguá | the same night from the highest point; the dark turned towards the forest |

## Links

- Website, with the palette and a page per port: [sp-night.github.io](https://sp-night.github.io)
- Install guides for every port: [sp-night.github.io/ports](https://sp-night.github.io/ports)
- Source: [github.com/sp-night/sp-night](https://github.com/sp-night/sp-night)
- License: [MIT](../LICENSE)
