package registry

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// Copy is the English prose every port README shares. It lives in one file
// because it was duplicated by hand across the ports, and duplicated prose is
// prose that diverges.
type Copy struct {
	LogoAlt        string            `yaml:"logo_alt"`
	TaglineStrong  string            `yaml:"tagline_strong"`
	Tagline        string            `yaml:"tagline"`
	FlavoursIntro  string            `yaml:"flavours_intro"`
	Flavours       map[string]string `yaml:"flavours"`
	Provenance     string            `yaml:"provenance"`
	MappingSection string            `yaml:"mapping_section"`
}

// LoadCopy parses the shared prose.
func LoadCopy(data []byte) (*Copy, error) {
	var c Copy
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("copy: %w", err)
	}
	if err := c.validate(); err != nil {
		return nil, err
	}
	return &c, nil
}

// EmbeddedCopy loads the prose that shipped with this binary.
func EmbeddedCopy() (*Copy, error) { return LoadCopy(copyYAML) }

func (c *Copy) validate() error {
	for _, f := range []struct{ name, value string }{
		{"logo_alt", c.LogoAlt},
		{"tagline_strong", c.TaglineStrong},
		{"tagline", c.Tagline},
		{"flavours_intro", c.FlavoursIntro},
		{"provenance", c.Provenance},
		{"mapping_section", c.MappingSection},
	} {
		if strings.TrimSpace(f.value) == "" {
			return fmt.Errorf("copy: %s is required", f.name)
		}
	}
	if len(c.Flavours) == 0 {
		return fmt.Errorf("copy: flavours is required — one English blurb per flavour")
	}
	return nil
}

// Blurb returns the English blurb for a flavour. The palette's own descriptions
// are Portuguese and stay that way; these are what the READMEs read.
func (c *Copy) Blurb(flavor string) (string, bool) {
	s, ok := c.Flavours[flavor]
	return s, ok
}
