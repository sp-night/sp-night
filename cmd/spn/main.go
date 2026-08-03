// Command spn turns the SP Night contract into finished files.
//
// The palette travels inside the binary, so a port repository carries only its
// mapping — a template — and the result. Run spn with no arguments for the
// list of commands.
package main

import (
	"fmt"
	"os"
)

// version is stamped by the release build (-ldflags "-X main.version=..."). A
// template declares the range of tool versions its mapping is written against,
// which is what lets a port's CI install the right one without being told.
var version = "dev"

type command struct {
	name    string
	summary string
	run     func(args []string) error
}

func commands() []command {
	return []command{
		{"gen", "render a port's templates into finished theme files", runGen},
		{"lint", "check that templates ask for roles, not raw colours", runLint},
		{"check", "audit contrast, accent separation and colour vision", runCheck},
		{"palette", "print the palette and the resolved role layer", runPalette},
		{"version", "print the tool version", runVersion},
	}
}

func main() {
	if len(os.Args) < 2 {
		usage(os.Stdout)
		return
	}
	name := os.Args[1]
	for _, f := range []string{"-h", "--help", "help"} {
		if name == f {
			usage(os.Stdout)
			return
		}
	}
	for _, c := range commands() {
		if c.name == name {
			if err := c.run(os.Args[2:]); err != nil {
				fmt.Fprintf(os.Stderr, "spn %s: %v\n", name, err)
				os.Exit(1)
			}
			return
		}
	}
	fmt.Fprintf(os.Stderr, "spn: unknown command %q\n\n", name)
	usage(os.Stderr)
	os.Exit(2)
}

func usage(w *os.File) {
	fmt.Fprintf(w, "spn — the SP Night engine (%s)\n\n", version)
	fmt.Fprintln(w, "usage: spn <command> [flags]")
	fmt.Fprintln(w)
	for _, c := range commands() {
		fmt.Fprintf(w, "  %-9s %s\n", c.name, c.summary)
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "The palette is embedded in this binary. Pass --palette to any command to")
	fmt.Fprintln(w, "read palette/ from disk instead — that is how the palette gets retuned.")
}

func runVersion([]string) error {
	fmt.Println(version)
	return nil
}
