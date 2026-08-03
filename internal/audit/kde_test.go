package audit

import (
	"io"
	"strings"
	"testing"
)

// The real regression: a generic kcolorscheme mapping put 1.11:1 text on
// Dolphin's alternating rows because KDE crosses every Foreground with every
// Background in a section. The check reads the built file, so it catches this
// no matter how the template arrived at it.
func TestKDESchemesCatchesTheAlternatingRowBug(t *testing.T) {
	scheme := `[Colors:View]
BackgroundNormal=21,23,35
BackgroundAlternate=29,31,45
ForegroundNormal=211,215,235
ForegroundInactive=33,35,47
`
	var buf strings.Builder
	n := KDESchemes(map[string][]byte{"dist/kde/sp_night_noite.colors": []byte(scheme)}, &buf)
	if n == 0 {
		t.Fatal("KDESchemes passed a near-invisible foreground")
	}
	out := buf.String()
	if !strings.Contains(out, "ForegroundInactive") {
		t.Errorf("report does not name the offending key:\n%s", out)
	}
	// It has to be reported against BOTH backgrounds of the section, which is
	// the crossing behaviour the check exists for.
	if got := strings.Count(out, "ForegroundInactive"); got != 2 {
		t.Errorf("ForegroundInactive reported %d time(s), want 2 (Normal and Alternate)", got)
	}
}

func TestKDESchemesAcceptsAReadableSection(t *testing.T) {
	scheme := `[Colors:View]
BackgroundNormal=21,23,35
BackgroundAlternate=29,31,45
ForegroundNormal=211,215,235
ForegroundActive=242,152,74
`
	if n := KDESchemes(map[string][]byte{"a.colors": []byte(scheme)}, io.Discard); n != 0 {
		t.Errorf("%d failure(s) on a readable section", n)
	}
}

// Only *.colors files are KDE schemes. A ghostty theme that happens to contain
// the word Background must not be parsed as one.
func TestKDESchemesIgnoresOtherTargets(t *testing.T) {
	other := map[string][]byte{
		"themes/sp_night_noite":      []byte("background = 151723\nforeground = d3d7eb\n"),
		"themes/sp_night_noite.yml":  []byte("BackgroundNormal=0,0,0\nForegroundNormal=1,1,1\n"),
		"themes/sp_night_noite.conf": []byte("[Colors:View]\nBackgroundNormal=0,0,0\nForegroundNormal=1,1,1\n"),
	}
	if n := KDESchemes(other, io.Discard); n != 0 {
		t.Errorf("%d failure(s) from files that are not KDE schemes", n)
	}
}

// [Inactive] variants share a section's colours and are audited with it.
func TestKDESchemesReadsInactiveVariants(t *testing.T) {
	scheme := `[Colors:Window][Inactive]
BackgroundNormal=21,23,35
ForegroundNormal=25,27,39
`
	if n := KDESchemes(map[string][]byte{"a.colors": []byte(scheme)}, io.Discard); n == 0 {
		t.Error("an [Inactive] section was not audited")
	}
}

// Malformed input must clamp rather than wrap a channel around. 999 wrapping
// into a uint8 would become 231 and quietly change the measurement; clamped it
// is white, and white on black is the maximum 21:1.
func TestKDESchemesClampsOutOfRangeChannels(t *testing.T) {
	scheme := `[Colors:View]
BackgroundNormal=999,999,999
ForegroundNormal=0,0,0
`
	if n := KDESchemes(map[string][]byte{"a.colors": []byte(scheme)}, io.Discard); n != 0 {
		t.Errorf("black on clamped white is 21:1 and must pass, got %d failure(s)", n)
	}
}
