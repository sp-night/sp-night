// Package palette holds the SP Night contract as data, embedded into the
// binary. A port never vendors these files: it declares a mapping, and the
// tool brings the colours with it.
//
// The three files are the whole contract. sp_night.json is the only place in
// the project where a hex is written down; roles.json names a palette key for
// every semantic role; schema.json describes the shape of the first two so an
// editor can validate them while you type.
package palette

import _ "embed"

// SpNight is palette/sp_night.json — 22 colours per flavour, and nothing else
// in the project is allowed to spell a hex.
//
//go:embed sp_night.json
var SpNight []byte

// Roles is palette/roles.json — the semantic layer. A template asks for a
// role; changing a role here repaints every port at once.
//
//go:embed roles.json
var Roles []byte

// Schema is palette/schema.json, published so editors can validate the two
// files above. It is not the gate: the gate is Palette.Validate, which runs on
// every command and catches what a JSON schema cannot express.
//
//go:embed schema.json
var Schema []byte
