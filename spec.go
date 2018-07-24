//go:generate stringer -type SectionID -trimprefix Section
//go:generate stringer -type OpCode -trimprefix Op
//go:generate stringer -type ExternalKind -trimprefix ExtKind

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

// OpCode is an operation code.
type OpCode int8

const (
	// OpInt32 is the op code for a 32 bit signed integer.
	OpInt32 OpCode = 0x7F

	// OpInt64 is the op code for a 64 bit signed integer.
	OpInt64 OpCode = 0x7E

	// OpFloat32 is the op code for a 32 bit float.
	OpFloat32 OpCode = 0x7D

	// OpFloat64 is the op code for a 64 bit float.
	OpFloat64 OpCode = 0x7C

	// OpAnyFunc is the op code for a function with any signatrue.
	OpAnyFunc OpCode = 0x70

	// OpFunc is the op code for a function.
	OpFunc OpCode = 0x60

	// OpBlock is the op for a pseudo type representing an empty block_type.
	OpBlock OpCode = 0x40
)

func (o OpCode) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%q", o.String())), nil
}

// ExternalKind defines the type for an external import.
type ExternalKind uint8

const (
	// ExtKindFunction indicates a Function import or definition.
	ExtKindFunction ExternalKind = iota

	// ExtKindTable indicates a Table import or definition.
	ExtKindTable

	// ExtKindMemory indicates a Memory import or definition.
	ExtKindMemory

	// ExtKindGlobal indicates a Global import or definition.
	ExtKindGlobal
)

func (e ExternalKind) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%q", e.String())), nil
}
