package render

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sp-night/sp-night/internal/theme"
)

// The fidelity gate.
//
// testdata holds the mapping and the finished files of the two ports that were
// already published before this engine existed. Rendering the mapping has to
// reproduce those files byte for byte — including the single trailing newline.
//
// If this test fails, the engine is not faithful to what users have installed,
// and the bug is in the renderer, not in the files. Do not update the golden
// files to make it pass unless the change to the shipped ports is deliberate.
func TestRenderReproducesThePublishedPorts(t *testing.T) {
	pal, roles, err := theme.Load()
	if err != nil {
		t.Fatalf("theme.Load: %v", err)
	}

	for _, port := range []struct {
		dir      string
		template string
		want     []string
	}{
		{
			dir:      "ghostty",
			template: "ghostty.tmpl",
			want: []string{
				"themes/sp_night_noite",
				"themes/sp_night_garoa",
				"themes/sp_night_jaragua",
			},
		},
		{
			dir:      "eza",
			template: "eza.yml.tmpl",
			want: []string{
				"themes/sp_night_noite.yml",
				"themes/sp_night_garoa.yml",
				"themes/sp_night_jaragua.yml",
			},
		},
	} {
		t.Run(port.dir, func(t *testing.T) {
			root := filepath.Join("testdata", port.dir)
			src, err := os.ReadFile(filepath.Join(root, port.template))
			if err != nil {
				t.Fatalf("read template: %v", err)
			}
			tpl, err := Parse(port.template, src)
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}

			files, err := tpl.Render(pal, roles, Meta{})
			if err != nil {
				t.Fatalf("Render: %v", err)
			}
			if len(files) != len(port.want) {
				t.Fatalf("rendered %d file(s), want %d", len(files), len(port.want))
			}

			got := map[string][]byte{}
			for _, f := range files {
				got[filepath.ToSlash(f.Path)] = f.Content
			}

			for _, rel := range port.want {
				content, ok := got[rel]
				if !ok {
					t.Errorf("did not render %s; rendered %v", rel, keys(got))
					continue
				}
				golden, err := os.ReadFile(filepath.Join(root, rel))
				if err != nil {
					t.Fatalf("read golden %s: %v", rel, err)
				}
				if string(content) != string(golden) {
					t.Errorf("%s does not match the published file\n%s", rel, firstDiff(golden, content))
				}
			}
		})
	}
}

// A generated file ends with exactly one newline. Two would show up as a diff
// in every port the first time anyone regenerates.
func TestRenderedFilesEndWithExactlyOneNewline(t *testing.T) {
	pal, roles, err := theme.Load()
	if err != nil {
		t.Fatalf("theme.Load: %v", err)
	}
	for _, name := range []string{"ghostty/ghostty.tmpl", "eza/eza.yml.tmpl"} {
		src, err := os.ReadFile(filepath.Join("testdata", name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		tpl, err := Parse(name, src)
		if err != nil {
			t.Fatalf("Parse %s: %v", name, err)
		}
		files, err := tpl.Render(pal, roles, Meta{})
		if err != nil {
			t.Fatalf("Render %s: %v", name, err)
		}
		for _, f := range files {
			c := f.Content
			if len(c) == 0 || c[len(c)-1] != '\n' {
				t.Errorf("%s does not end with a newline", f.Path)
				continue
			}
			if len(c) > 1 && c[len(c)-2] == '\n' {
				t.Errorf("%s ends with more than one newline", f.Path)
			}
		}
	}
}

// The published templates must also pass the roles-only rule. Neither of them
// declares raw_palette, so a stray .C would be a finding.
func TestPublishedTemplatesAskForRolesOnly(t *testing.T) {
	for _, name := range []string{"ghostty/ghostty.tmpl", "eza/eza.yml.tmpl"} {
		src, err := os.ReadFile(filepath.Join("testdata", name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		tpl, err := Parse(name, src)
		if err != nil {
			t.Fatalf("Parse %s: %v", name, err)
		}
		for _, f := range tpl.Lint() {
			t.Errorf("%s:%s", name, f)
		}
	}
}

func keys[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

// firstDiff reports the first differing line, which is far more useful than
// dumping two whole config files into the test log.
func firstDiff(want, got []byte) string {
	wl, gl := splitLines(string(want)), splitLines(string(got))
	for i := range max(len(wl), len(gl)) {
		var w, g string
		if i < len(wl) {
			w = wl[i]
		}
		if i < len(gl) {
			g = gl[i]
		}
		if w != g {
			return "  line " + itoa(i+1) + "\n    want: " + w + "\n    got:  " + g
		}
	}
	return "  files differ only in trailing bytes"
}

func splitLines(s string) []string {
	var out []string
	start := 0
	for i := range len(s) {
		if s[i] == '\n' {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		out = append(out, s[start:])
	}
	return out
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}
