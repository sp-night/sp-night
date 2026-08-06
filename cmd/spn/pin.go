package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/sp-night/sp-night/internal/port"
	"github.com/sp-night/sp-night/internal/render"
)

// runPin rewrites the engine version a mapping declares.
//
// It exists so the fan-out can bump the pin in the same pull request that
// regenerates the files, which is the only combination that is ever true: the
// committed theme files and the version that produced them are one fact, and a
// port whose pin and files disagree has a red CI through no fault of its own.
//
// A dedicated command rather than a sed in the workflow, because the frontmatter
// is the contract's own format and this repository is where it is parsed.
func runPin(args []string) error {
	fs := flag.NewFlagSet("pin", flag.ContinueOnError)
	check := fs.Bool("check", false, "report whether the pin already matches, write nothing")
	if err := parseArgs(fs, args); err != nil {
		return err
	}

	rest := fs.Args()
	if len(rest) == 0 {
		return fmt.Errorf("name the version to pin, e.g. `spn pin 1.3.0`")
	}
	want := strings.TrimPrefix(rest[0], "v")

	templates, err := findTemplates(rest[1:])
	if err != nil {
		return err
	}

	stale := 0
	for _, path := range templates {
		src, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		fm, _, err := render.SplitFrontmatter(src)
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		if fm.SPN.Version == want {
			fmt.Printf("  %s already declares %s\n", path, want)
			continue
		}
		stale++
		if *check {
			fmt.Printf("  %s declares %s, want %s\n", path, fm.SPN.Version, want)
			continue
		}

		out, err := replaceVersion(src, fm.SPN.Version, want)
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		if err := os.WriteFile(path, out, 0o644); err != nil {
			return err
		}
		fmt.Printf("  %s %s -> %s\n", path, fm.SPN.Version, want)
	}

	if *check && stale > 0 {
		return fmt.Errorf("%d template(s) declare a different version — run `spn pin %s`", stale, want)
	}
	return nil
}

// scaffoldVersion is what a freshly scaffolded mapping pins itself to: the
// version of the binary doing the scaffolding, so the mapping and the tool that
// wrote it are the same fact from the first commit. A build that does not know
// its own version — `go run` from a checkout — falls back to a range, because
// naming a release it is not would be a lie the ports would inherit.
func scaffoldVersion() string {
	if version == "" || version == "dev" {
		return port.FallbackVersion
	}
	return strings.TrimPrefix(version, "v")
}

// replaceVersion swaps the declared version inside the frontmatter, leaving the
// rest of the file byte for byte alone. Rewriting the YAML by re-marshalling it
// would reflow comments and quoting across a file a human maintains.
func replaceVersion(src []byte, from, to string) ([]byte, error) {
	text := string(src)

	// Only the frontmatter, so a body that happens to mention a version string
	// is never touched. The frontmatter is the first --- delimited block.
	const delim = "---"
	if !strings.HasPrefix(text, delim) {
		return nil, fmt.Errorf("no frontmatter to pin")
	}
	end := strings.Index(text[len(delim):], "\n"+delim)
	if end < 0 {
		return nil, fmt.Errorf("frontmatter is not closed")
	}
	end += len(delim) + 1

	head, tail := text[:end], text[end:]
	for _, quoted := range []string{`"` + from + `"`, `'` + from + `'`, from} {
		if strings.Contains(head, quoted) {
			return []byte(strings.Replace(head, quoted, `"`+to+`"`, 1) + tail), nil
		}
	}
	return nil, fmt.Errorf("the frontmatter does not spell the version %q", from)
}
