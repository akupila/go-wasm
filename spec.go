//go:generate stringer -type SectionID -trimprefix Section

package wasm

import "fmt"

// SectionID the id of a section in the wasm file.
type SectionID uint8

const (
	// SectionCustom is a custom section.
	SectionCustom SectionID = iota

	// SectionType contains function signature declarations.
	SectionType

	// SectionImport contains import declarations.
	SectionImport

	// SectionFunction contains function declarations.
	SectionFunction

	// SectionTable contains an indirect function table and other tables.
	SectionTable

	// SectionMemory contains memory attributes.
	SectionMemory

	// SectionGlobal contains global declarations.
	SectionGlobal

	// SectionExport contains exports from the WASM module.
	SectionExport

	// SectionElement starts function declarations.
	SectionStart

	// SectionElement contains elements.
	SectionElement

	// SectionCode contains function bodies.
	SectionCode

	// SectionData contains data segments.
	SectionData
)

func (s SectionID) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%q", s.String())), nil
}
