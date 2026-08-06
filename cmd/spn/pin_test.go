package main

import (
	"strings"
	"testing"
)

// The mapping is a file a human maintains: comments, blank lines and quoting
// style are all deliberate. Pinning rewrites one value and leaves the rest byte
// for byte, so the regeneration pull request shows the bump and nothing else.
func TestReplaceVersionTouchesOnlyTheDeclaredVersion(t *testing.T) {
	src := []byte(`---
spn:
  version: "^1.0"
  matrix: [flavor]
  filename: "themes/sp_night_{{ .Flavor }}.conf"
---
# SP Night — {{ .FlavorLabel }}
#
# Written against spn ^1.0, which is not a thing this line should change.
background {{ .R.ui.bg }}
`)

	out, err := replaceVersion(src, "^1.0", "1.3.0")
	if err != nil {
		t.Fatalf("replaceVersion: %v", err)
	}
	got := string(out)

	if !strings.Contains(got, `version: "1.3.0"`) {
		t.Errorf("the frontmatter was not pinned:\n%s", got)
	}
	// The body mentions the old range in a comment. Rewriting it would edit a
	// human's prose, and worse, would do it silently.
	if !strings.Contains(got, "Written against spn ^1.0, which is not a thing") {
		t.Error("the body's mention of the old version was rewritten")
	}
	if strings.Count(got, "1.3.0") != 1 {
		t.Errorf("the new version appears %d times, want once", strings.Count(got, "1.3.0"))
	}
	for _, keep := range []string{
		"  matrix: [flavor]",
		`  filename: "themes/sp_night_{{ .Flavor }}.conf"`,
		"background {{ .R.ui.bg }}",
	} {
		if !strings.Contains(got, keep) {
			t.Errorf("%q did not survive the rewrite", keep)
		}
	}
}

// An exact pin is the normal case after the first bump, so pinning has to be
// idempotent rather than only understanding ranges.
func TestReplaceVersionRepinsAnExactVersion(t *testing.T) {
	src := []byte("---\nspn:\n  version: \"1.2.0\"\n  matrix: []\n---\nbody\n")
	out, err := replaceVersion(src, "1.2.0", "1.3.0")
	if err != nil {
		t.Fatalf("replaceVersion: %v", err)
	}
	if !strings.Contains(string(out), `version: "1.3.0"`) {
		t.Errorf("not repinned:\n%s", out)
	}
}

func TestReplaceVersionRefusesAFileWithNoFrontmatter(t *testing.T) {
	if _, err := replaceVersion([]byte("background #fff\n"), "^1.0", "1.3.0"); err == nil {
		t.Error("a template with no frontmatter should not be pinned silently")
	}
}
