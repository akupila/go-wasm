//go:generate stringer -type LanguageType -trimprefix LanguageType_
//go:generate stringer -type ValueType -trimprefix ValueType_

package wasm

// A Module represents a parsed WASM module.
type Module struct {
	// Sections contains the sections in the parsed file, in the order they
	// appear in the file.
	//
	// The items in the slice will be a mix of the SectionXXX types.
	Sections []interface{}
}

// SectionCustom is a custom or name section.
//
// A name section provides debug information, a generic custom section is
// something the compiler generated and not part of the spec.
type SectionCustom struct {
	// Name is the name of the section.
	Name string
	// Payload is the raw payload for the section.
	Payload []byte
}

// SectionType is a section for type definitions. The section declares all
// function signatrues that will be used in the module.
type SectionType struct {
	Entries []typeEntry
}

type typeEntry struct {
	// form is the value for a func type constructor (0x60)
	Form LanguageType
	// Params contains the parameter types of the function.
	Params []ValueType
	// ReturnCount returns the number of results from the function.
	// The value will be 0 or 1. Future version may allow more:
	// https://github.com/WebAssembly/design/issues/1146
	ReturnCount uint8
	// ReturnType is the result type if ReturnCount > 0
	ReturnTypes []ValueType
}

type SectionImport struct {
	Entries []importEntry
}

type importEntry struct {
	Module string
	Field  string
	Kind   ExternalKind
	// One of these is set, depending on the Kind
	FunctionType *functionType
	TableType    *tableType
	MemoryType   *memoryType
	GlobalType   *globalType
}

type functionType struct {
	Index uint32
}

type memoryType struct {
	Limits resizableLimits
}

type tableType struct {
	ElemType uint8
	Limits   resizableLimits
}

type globalType struct {
	ContentType uint8
	Mutable     bool
}

type resizableLimits struct {
	// Initial is the initial length of the memory.
	Initial uint32
	// Maximum is the maximum length of the memory. May not be set.
	Maximum uint32
}

type SectionFunction struct {
	// Types contains a sequence of indices into the type section.
	Types []uint32
}

type SectionTable struct {
	Entries []memoryType
}

type SectionMemory struct {
	Entries []memoryType
}

type SectionGlobal struct {
	Globals []globalVariable
}

type globalVariable struct {
	Type globalType
	Init []byte
}

type SectionExport struct {
	Entries []exportEntry
}

type exportEntry struct {
	Field string
	Kind  ExternalKind
	Index uint32
}

type SectionStart struct {
	Index uint32
}

type SectionElement struct {
	Entries []elemSegment
}

type elemSegment struct {
	Index  uint32
	Offset []byte
	Elems  []uint32
}

type SectionCode struct {
	Bodies []functionBody
}

type functionBody struct {
	Locals []localEntry
	Code   []byte
}

type localEntry struct {
	Count uint32
	Type  LanguageType
}

type SectionData struct {
	Entries []dataSegment
}

type dataSegment struct {
	Index  uint32
	Offset []uint8
	Data   []byte
}

type SectionName struct {
	Name      string
	Module    string
	Functions *nameMap
	Locals    *locals
}

type nameMap struct {
	Names []naming
}

type naming struct {
	Index uint32
	Name  string
}

type locals struct {
	Funcs []localName
}

type localName struct {
	Index    uint32
	LocalMap nameMap
}

// ExternalKind defines the type for an external import.  type ExternalKind uint8
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

type LanguageType uint8

const (
	LanguageType_i32     LanguageType = 0x7f
	LanguageType_i64     LanguageType = 0x7e
	LanguageType_f32     LanguageType = 0x7d
	LanguageType_f64     LanguageType = 0x7c
	LanguageType_anyfunc LanguageType = 0x70
	LanguageType_func    LanguageType = 0x60
	LanguageType_block   LanguageType = 0x40
)

type ValueType uint8

const (
	ValueType_i32 ValueType = ValueType(LanguageType_i32)
	ValueType_i64 ValueType = ValueType(LanguageType_i64)
	ValueType_f32 ValueType = ValueType(LanguageType_f32)
	ValueType_f64 ValueType = ValueType(LanguageType_f64)
)

type NameType uint8

const (
	NameTypeModule NameType = iota
	NameTypeFunction
	NameTypeLocal
)
