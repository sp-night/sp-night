# Command reference

```
spn <command> [flags]
```

`spn` with no arguments, `spn help` or `spn -h` prints the list of commands.
Flags may come before or after the positional arguments: `spn new eza --out x`
and `spn new --out x eza` are the same command.

The palette and the port catalogue are embedded in the binary. The flags below
that take a file or directory read that copy from disk instead, which is how the
contract gets changed and measured before it ships.

## Commands that write files

`--check` on `gen`, `readme` and `preview` writes nothing and fails when what is
committed is out of date. That is what a port's CI runs.

### spn gen

Render a port's mapping into finished theme files.

```sh
spn gen [templates...]
```

With no arguments it renders every `*.tmpl` in the current directory, so a port
repository runs bare `spn gen`.

| flag | does |
|---|---|
| `--check` | do not write; fail if any file on disk is out of date |
| `--stdout` | write to stdout instead of to files |
| `--app <slug>` | registry slug for this port (default: inferred from the template name) |
| `--registry <file>` | read the port catalogue from this file |
| `--palette <dir>` | read the contract from this directory |
| `-q` | print only failures |

### spn readme

Render a port's canonical README from the catalogue.

```sh
spn readme [slug]
```

| flag | does |
|---|---|
| `--out <file>` | file to write (default `README.md`) |
| `--check` | do not write; fail if the README is out of date |
| `--stdout` | write to stdout instead of to a file |
| `--registry <file>` | read the port catalogue from this file |
| `--copy <file>` | read the shared README prose from this file |
| `--palette <dir>` | read the contract from this directory |

### spn preview

Draw the synthetic preview for each flavour. The previews are SVGs drawn from the
palette, so they cannot show a colour the user will not get.

```sh
spn preview [slug]
```

| flag | does |
|---|---|
| `--out <dir>` | directory to write the previews into (default `assets`) |
| `--check` | do not write; fail if any preview is out of date |
| `--registry <file>` | read the port catalogue from this file |
| `--copy <file>` | read the shared README prose from this file |
| `--palette <dir>` | read the contract from this directory |

### spn new

Scaffold a new port repository: the mapping stub with the canonical header, the
six-line workflow, the Renovate config and the editor settings. The port has to
be listed in the catalogue first.

```sh
spn new <slug>
```

| flag | does |
|---|---|
| `--out <dir>` | directory to scaffold into (default: the slug) |
| `--force` | overwrite files the scaffold would replace |
| `--registry <file>` | read the port catalogue from this file |
| `--copy <file>` | read the shared README prose from this file |

### spn pin

Rewrite the engine version a mapping declares in its frontmatter. The release
fan-out runs it in the same pull request that regenerates a port's files.

```sh
spn pin <version> [templates...]
```

| flag | does |
|---|---|
| `--check` | report whether the pin already matches, write nothing |

## Commands that check

### spn lint

Check that mappings ask for roles, never raw colours. It walks the template's
parse tree, so a counter-example inside a comment is not a finding.

```sh
spn lint [templates...]
```

### spn check

Audit contrast, accent separation and colour vision for every flavour. Fails when
a pair falls below its floor.

| flag | does |
|---|---|
| `-v` | also list the passing pairs and the colour vision summary |
| `--json` | print what the audit measures and the policy it measures against, as data |
| `--palette <dir>` | read the contract from this directory |

### spn registry

List and validate the port catalogue.

| flag | does |
|---|---|
| `--json` | print the catalogue as JSON — used to build the fan-out matrix |
| `--registry <file>` | read the port catalogue from this file |
| `--copy <file>` | read the shared README prose from this file |

## Commands that print

### spn entry

Print a starter catalogue entry for a new port, with every required field, a
mapping table to extend and a complete preview. It validates as printed.

```sh
spn entry <slug> >> registry/ports.yml
```

### spn palette

Print the palette and the resolved role layer.

| flag | does |
|---|---|
| `--roles` | print the resolved role layer instead of the colours |
| `--json` | print the contract as JSON instead of a table |
| `--palette <dir>` | read the contract from this directory |

### spn version

Print the tool version. A build from source prints `dev`.
