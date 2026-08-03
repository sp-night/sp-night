package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"

	"github.com/sp-night/sp-night/internal/color"
	"github.com/sp-night/sp-night/internal/theme"
)

// paletteFlag registers the --palette flag shared by every command. Empty
// means "use the copy embedded in this binary", which is what a port always
// wants; a path is how this repository retunes the palette and measures the
// result before it ships.
func paletteFlag(fs *flag.FlagSet) *string {
	return fs.String("palette", "", "read the contract from this directory instead of the embedded copy")
}

func loadContract(dir string) (*theme.Palette, theme.Roles, error) {
	if dir == "" {
		return theme.Load()
	}
	return theme.LoadDir(dir)
}

func runPalette(args []string) error {
	fs := flag.NewFlagSet("palette", flag.ContinueOnError)
	dir := paletteFlag(fs)
	asJSON := fs.Bool("json", false, "print the contract as JSON instead of a table")
	roles := fs.Bool("roles", false, "print the resolved role layer instead of the colours")
	if err := parseArgs(fs, args); err != nil {
		return err
	}

	pal, rol, err := loadContract(*dir)
	if err != nil {
		return err
	}

	if *asJSON {
		return printJSON(pal, rol, *roles)
	}
	if *roles {
		printRoles(pal, rol)
		return nil
	}
	printColors(pal)
	return nil
}

func printColors(pal *theme.Palette) {
	fmt.Printf("%s · %d flavours · %d colours\n", pal.Label, len(pal.Flavors), len(pal.Order))

	// One column per flavour, in declaration order.
	fmt.Printf("\n%-15s", "")
	for _, f := range pal.Flavors {
		fmt.Printf("  %-9s", f.ID)
	}
	fmt.Println()

	for _, g := range theme.Groups {
		fmt.Printf("\n  %s\n", g.Title)
		for _, k := range g.Keys {
			fmt.Printf("  %-15s", k)
			for _, f := range pal.Flavors {
				fmt.Printf("  %-9s", f.Colors[k])
			}
			if m := pal.Meaning[k]; m != "" {
				fmt.Printf("  %s", m)
			}
			fmt.Println()
		}
	}
}

func printRoles(pal *theme.Palette, rol theme.Roles) {
	groups := make([]string, 0, len(rol))
	for g := range rol {
		groups = append(groups, g)
	}
	sort.Strings(groups)

	for _, g := range groups {
		fmt.Printf("\n  %s\n", g)
		names := make([]string, 0, len(rol[g]))
		for n := range rol[g] {
			names = append(names, n)
		}
		sort.Strings(names)
		for _, n := range names {
			key := rol[g][n]
			fmt.Printf("  %-16s %-14s", n, key)
			for _, f := range pal.Flavors {
				fmt.Printf("  %-9s", f.Colors[key])
			}
			fmt.Println()
		}
	}
}

// printJSON emits the contract as data. Useful for wiring the palette into
// something that is not a template — and it is how the website's vendored copy
// gets checked against this binary.
func printJSON(pal *theme.Palette, rol theme.Roles, wantRoles bool) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)

	if wantRoles {
		out := map[string]any{}
		for _, f := range pal.Flavors {
			resolved, err := rol.Resolve(f)
			if err != nil {
				return err
			}
			out[f.ID] = resolved
		}
		return enc.Encode(out)
	}

	type flavorOut struct {
		Label       string            `json:"label"`
		Appearance  string            `json:"appearance"`
		Description string            `json:"description"`
		Colors      map[string]string `json:"colors"`
		Contrast    map[string]string `json:"contrast_on_laje"`
	}
	out := struct {
		Name    string               `json:"name"`
		Label   string               `json:"label"`
		URL     string               `json:"url"`
		Order   []string             `json:"order"`
		Flavors map[string]flavorOut `json:"flavors"`
	}{
		Name: pal.Name, Label: pal.Label, URL: pal.URL,
		Order: pal.Order, Flavors: map[string]flavorOut{},
	}
	for _, f := range pal.Flavors {
		ratios := map[string]string{}
		laje := color.MustParseHex(f.Colors["laje"])
		for _, k := range pal.Order {
			ratios[k] = fmt.Sprintf("%.2f", color.Contrast(color.MustParseHex(f.Colors[k]), laje))
		}
		out.Flavors[f.ID] = flavorOut{
			Label: f.Label, Appearance: f.Appearance, Description: f.Description,
			Colors: f.Colors, Contrast: ratios,
		}
	}
	return enc.Encode(out)
}
