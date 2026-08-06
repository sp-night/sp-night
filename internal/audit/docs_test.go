package audit

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

// The counts in the prose are derived facts, and derived facts written by hand
// drift. Adding fg_vivo to the palette meant editing "22 colours" in six places
// across two repositories, and one of them — README.md's pair count — was still
// wrong afterwards, with nothing to say so: the number is not wired to anything
// that runs.
//
// This is the wire. It does not generate the docs; it reads the numbers back out
// of them and compares against what the code measures, so a stale count fails
// the build instead of misinforming a reader.
func TestTheDocsStateTheCountsTheCodeMeasures(t *testing.T) {
	pal := mustLoad(t)

	colours := len(pal.Order)
	pairs := len(Flavor(pal, pal.Flavors[0]).Checks)

	// Each pattern captures one number that has to equal one measurement. The
	// wording is deliberately loose about what surrounds the number and strict
	// about the noun, so rephrasing a sentence does not break the test but
	// changing the claim does.
	for _, c := range []struct {
		pattern *regexp.Regexp
		want    int
		what    string
	}{
		{regexp.MustCompile(`(\d+) (?:named )?colours`), colours, "palette colours"},
		{regexp.MustCompile(`(\d+) pairs`), pairs, "measured pairs"},
	} {
		for _, name := range []string{
			filepath.Join("..", "..", "docs", "SPEC.md"),
			filepath.Join("..", "..", "README.md"),
		} {
			body, err := os.ReadFile(name)
			if err != nil {
				t.Fatalf("read %s: %v", name, err)
			}
			for _, m := range c.pattern.FindAllSubmatch(body, -1) {
				got := 0
				if _, err := fmt.Sscanf(string(m[1]), "%d", &got); err != nil {
					t.Fatalf("%s: %q is not a number: %v", name, m[1], err)
				}
				if got != c.want {
					t.Errorf("%s says %q, but the code measures %d %s",
						filepath.Base(name), m[0], c.want, c.what)
				}
			}
		}
	}
}
