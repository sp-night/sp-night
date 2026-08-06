package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/sp-night/sp-night/internal/audit"
)

func runCheck(args []string) error {
	fs := flag.NewFlagSet("check", flag.ContinueOnError)
	dir := paletteFlag(fs)
	verbose := fs.Bool("v", false, "also list the passing pairs and the colour vision summary")
	asJSON := fs.Bool("json", false, "print what the audit measures and the policy it measures against, as data")
	if err := parseArgs(fs, args); err != nil {
		return err
	}

	pal, _, err := loadContract(*dir)
	if err != nil {
		return err
	}

	// The summary is for whoever publishes the policy rather than enforces it —
	// the website prints it as a table, and typing those numbers by hand is how
	// it came to contradict the tool. It reports the policy, not a verdict, so it
	// does not gate: `spn check` without --json is what fails a build.
	if *asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(audit.Summarise(pal))
	}

	fmt.Printf("%s · %d flavours · %d colours\n", pal.Label, len(pal.Flavors), len(pal.Order))

	if failures := audit.All(pal, os.Stdout, *verbose); failures > 0 {
		return fmt.Errorf("%d pair(s) below the contrast floor", failures)
	}
	fmt.Println("\ncontrast ok")
	return nil
}
