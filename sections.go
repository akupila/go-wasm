package wasm

// A Module represents a parsed WASM module.
type Module struct {
	// Sections contains the sections in the parsed file, in the order they
	// appear in the file.
	//
	// The items in the slice will be a mix of the SectionXXX types.
	Sections []interface{}
}

// SectionCustom is a custom or name section added by the compiler that
// generated the WASM file.
type SectionCustom struct {
	// Name is the name of the section.
	Name string
	// Payload is the raw payload for the section.
	Payload []byte
}

// SectionType is a section for type definitions. The section declares all
// function signatures that will be used in the module.
type SectionType struct {
	Entries []funcType
}

type funcType struct {
	// form is the value for a func type constructor (0x60)
	Form int8
	// Params contains the parameter types of the function.
	Params []valueType
	// ReturnCount returns the number of results from the function.
	// The value will be 0 or 1. Future version may allow more:
	// https://github.com/WebAssembly/design/issues/1146
	ReturnCount uint8
	// ReturnType is the result type if ReturnCount > 0
	ReturnTypes []valueType
}

// SectionImport is the section for declaring imports. It declares all imports
// that will be used in the module.
type SectionImport struct {
	Entries []importEntry
}

type importEntry struct {
	Module string
	Field  string
	Kind   externalKind
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
	ElemType elemType
	Limits   resizableLimits
}

type globalType struct {
	ContentType valueType
	Mutable     bool
}

type resizableLimits struct {
	// Initial is the initial length of the memory.
	Initial uint32
	// Maximum is the maximum length of the memory. May not be set.
	Maximum uint32
}

// SectionFunction declares the signatures of all functions in the modules.
// The definitions of the functions will be in the code section.
type SectionFunction struct {
	// Types contains a sequence of indices into the type section.
	Types []uint32
}

// SectionTable declares a table section. A table is similar to linear memory,
// whose elements, instead of being bytes, are opaque values of a particular
// table element. This allows the table to contain values -- like GC
// references, raw OS handles, or native pointers -- that are accessed by
// WebAssembly code indirectly through an integer index.
//
// See: https://github.com/WebAssembly/design/blob/master/Semantics.md#table
type SectionTable struct {
	Entries []memoryType
}

// SectionMemory declares a memory section. The section provides an internal
// definition of one linear memory.
//
// See: https://github.com/WebAssembly/design/blob/master/Modules.md#linear-memory-section
type SectionMemory struct {
	Entries []memoryType
}

// SectionGlobal provides an internal definition of global variables.
//
// See: https://github.com/WebAssembly/design/blob/master/Modules.md#global-section
type SectionGlobal struct {
	Globals []globalVariable
}

type globalVariable struct {
	Type globalType
	Init []byte
}

// SectionExport declares exports from the WASM module.
//
// See: https://github.com/WebAssembly/design/blob/master/Modules.md#exports
type SectionExport struct {
	Entries []exportEntry
}

type exportEntry struct {
	Field string
	Kind  externalKind
	Index uint32
}

// SectionStart defines the start node, if the module has a start node defined.
//
// See: https://github.com/WebAssembly/design/blob/master/Modules.md#module-start-function
type SectionStart struct {
	Index uint32
}

// SectionElement defines element segments that initialize elements of imported
// or internally-defined tables with any other definition in the module.
//
// See: https://github.com/WebAssembly/design/blob/master/Modules.md#elements-section
type SectionElement struct {
	Entries []elemSegment
}

type elemSegment struct {
	Index  uint32
	Offset []byte
	Elems  []uint32
}

// SectionCode contains a function body for every functino in the module.
type SectionCode struct {
	Bodies []functionBody
}

type functionBody struct {
	Locals []localEntry
	Code   []byte
}

type localEntry struct {
	Count uint32
	Type  OpCode
}

// SectionData declares the initialized data that is loaded into the linear memory.
type SectionData struct {
	Entries []dataSegment
}

type dataSegment struct {
	Index  uint32
	Offset []uint8
	Data   []byte
}

// SectionName is a custom section that provides debugging information, by
// matching indices to human readable names.
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

type externalKind uint8

const (
	// ExtKindFunction indicates a Function import or definition.
	ExtKindFunction externalKind = iota
	// ExtKindTable indicates a Table import or definition.
	ExtKindTable
	// ExtKindMemory indicates a Memory import or definition.
	ExtKindMemory
	// ExtKindGlobal indicates a Global import or definition.
	ExtKindGlobal
)

type NameType uint8

const (
	NameTypeModule NameType = iota
	NameTypeFunction
	NameTypeLocal
)

// varint7
type valueType int8

// varint7
type langType int8

// varint7
type elemType int8
