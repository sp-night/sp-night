package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/sp-night/sp-night/internal/audit"
)

func runCheck(args []string) error {
	fs := flag.NewFlagSet("check", flag.ContinueOnError)
	dir := paletteFlag(fs)
	verbose := fs.Bool("v", false, "also list the passing pairs and the colour vision summary")
	if err := fs.Parse(args); err != nil {
		return err
	}

	pal, _, err := loadContract(*dir)
	if err != nil {
		return err
	}

	fmt.Printf("%s · %d flavours · %d colours\n", pal.Label, len(pal.Flavors), len(pal.Order))

	if failures := audit.All(pal, os.Stdout, *verbose); failures > 0 {
		return fmt.Errorf("%d pair(s) below the contrast floor", failures)
	}
	fmt.Println("\ncontrast ok")
	return nil
}
