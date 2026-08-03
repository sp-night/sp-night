package audit

import (
	"fmt"
	"io"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/sp-night/sp-night/internal/color"
)

// Auditing generated KDE colour schemes.
//
// KDE draws text by crossing ANY Foreground* with ANY Background* of the same
// section — which is how a generic kcolorscheme mapping produced 1.11:1 text
// on Dolphin's alternating rows. Here the tool parses the file it has just
// built and fails the build if any pair in any section lands below AA.
//
// The check reads the OUTPUT rather than the template, so it keeps holding
// even if the template's logic changes. No other theme project does this; the
// real Breeze has pairs at 2–3:1 in Selection.

var (
	kdeSectionRe = regexp.MustCompile(`(?s)\[Colors:(\w+)\](?:\[Inactive\])?([^[]*)`)
	kdeEntryRe   = regexp.MustCompile(`(?m)^(\w+)=(\d+),(\d+),(\d+)$`)
)

type kdeEntry struct {
	key string
	r   uint8
	g   uint8
	b   uint8
}

// KDESchemes checks every fg/bg pair of every generated *.colors file and
// returns the number of fatal pairs.
func KDESchemes(built map[string][]byte, w io.Writer) int {
	// Sorted so the report is stable across runs; a map is not.
	paths := make([]string, 0, len(built))
	for p := range built {
		if strings.HasSuffix(p, ".colors") {
			paths = append(paths, p)
		}
	}
	sort.Strings(paths)

	fatal := 0
	for _, path := range paths {
		for _, sec := range kdeSectionRe.FindAllStringSubmatch(string(built[path]), -1) {
			name, body := sec[1], sec[2]

			var bgs, fgs []kdeEntry
			for _, e := range kdeEntryRe.FindAllStringSubmatch(body, -1) {
				entry := kdeEntry{key: e[1], r: atoi8(e[2]), g: atoi8(e[3]), b: atoi8(e[4])}
				switch {
				case strings.HasPrefix(e[1], "Background"):
					bgs = append(bgs, entry)
				case strings.HasPrefix(e[1], "Foreground"):
					fgs = append(fgs, entry)
				}
			}

			for _, bg := range bgs {
				for _, fg := range fgs {
					ratio := contrastOf(fg, bg)
					if ratio < LevelAA {
						fatal++
						fmt.Fprintf(w, "    ✗ kde %-12s %-22s on %-22s %5.2f:1  (floor %.1f)\n",
							name, fg.key, bg.key, ratio, LevelAA)
					}
				}
			}
		}
	}
	return fatal
}

func contrastOf(fg, bg kdeEntry) float64 {
	return color.Contrast(
		color.RGB{R: fg.r, G: fg.g, B: fg.b},
		color.RGB{R: bg.r, G: bg.g, B: bg.b},
	)
}

func atoi8(s string) uint8 {
	n, _ := strconv.Atoi(s)
	if n < 0 {
		return 0
	}
	if n > 255 {
		return 255
	}
	return uint8(n)
}
