package registry

import _ "embed"

// portsYAML is registry/ports.yml — the catalogue.
//
//go:embed ports.yml
var portsYAML []byte

// copyYAML is registry/copy.yml — the English prose every port README shares.
//
//go:embed copy.yml
var copyYAML []byte
