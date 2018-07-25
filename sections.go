package wasm

// A Module represents a parsed WASM module.
type Module struct {
	// Sections contains the sections in the parsed file, in the order they
	// appear in the file. A valid  but empty file will have zero sections.
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

// SectionType declares all function type definitions used in the module.
type SectionType struct {
	// Entries are the entries in a Type section. Each entry declares one type.
	Entries []FuncType
}

// A FuncType is the description of a function signature.
type FuncType struct {
	// Form is the value for a func type constructor (always 0x60, the op code
	// for a function).
	Form int8

	// Params contains the parameter types of the function.
	Params []valueType

	// ReturnCount returns the number of results from the function.
	// The value will be 0 or 1.
	//
	// Future version may allow more: https://github.com/WebAssembly/design/issues/1146
	ReturnCount uint8

	// ReturnType is the result type if ReturnCount > 0.
	ReturnTypes []valueType
}

// SectionImport declares all imports defined by the module.
type SectionImport struct {
	Entries []ImportEntry
}

// ImportEntry describes an individual import to the module.
type ImportEntry struct {
	// Module is the name of the module.
	Module string

	// Field is the field name being imported.
	Field string

	// Kind specified the type of import. The import type value will be set
	// depending on the kind, the other ones will be nil.
	Kind ExternalKind

	// FunctionType describes a function import, if Kind == ExtKindFunction.
	FunctionType *FunctionType

	// TableType describes a table import, if Kind == ExtKindTable.
	TableType *TableType

	// MemoryType describes a memory import, if Kind == ExtKindMemory.
	MemoryType *MemoryType

	// GlobalType describes a global import, if Kind == ExtKindGlobal.
	GlobalType *GlobalType
}

// FunctionType the type for a function import.
type FunctionType struct {
	// Index is the index of the function signature.
	Index uint32
}

// MemoryType is the type for a memory import.
type MemoryType struct {
	// Limits contains memory limits defined by the import.
	Limits ResizableLimits
}

// TableType is the type for a table import.
type TableType struct {
	// ElemType specifies the type of the elements.
	ElemType elemType

	// Limits specifies the resizable limits of the table.
	Limits ResizableLimits
}

// GlobalType is the type for a global import.
type GlobalType struct {
	// ContentType is the type of the value.
	ContentType valueType

	// Mutable is true if the global value can be modified.
	Mutable bool
}

// ResizableLimits describes the limits of a table or memory.
type ResizableLimits struct {
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
// https://github.com/WebAssembly/design/blob/master/Semantics.md#table
type SectionTable struct {
	Entries []MemoryType
}

// SectionMemory declares a memory section. The section provides an internal
// definition of one linear memory.
//
// https://github.com/WebAssembly/design/blob/master/Modules.md#linear-memory-section
type SectionMemory struct {
	Entries []MemoryType
}

// SectionGlobal provides an internal definition of global variables.
//
// https://github.com/WebAssembly/design/blob/master/Modules.md#global-section
type SectionGlobal struct {
	Globals []GlobalVariable
}

// A GlobalVariable is a global variable defined by the module.
type GlobalVariable struct {
	// Type is the type of the global variable.
	Type GlobalType

	// Init is an init expression (wasm bytecode) to set the initial value of
	// the global variable.
	Init []byte
}

// SectionExport declares exports from the WASM module.
//
// https://github.com/WebAssembly/design/blob/master/Modules.md#exports
type SectionExport struct {
	Entries []ExportEntry
}

// ExportEntry specifies an individual export from the module.
type ExportEntry struct {
	// Field is the name of the field being exported.
	Field string

	// Kind is the kind of export.
	Kind ExternalKind

	// Index is the index into the corresponding index space.
	//
	// https://github.com/WebAssembly/design/blob/master/Modules.md#function-index-space
	Index uint32
}

// SectionStart defines the start node, if the module has a start node defined.
//
// https://github.com/WebAssembly/design/blob/master/Modules.md#module-start-function
type SectionStart struct {
	// Index is the index to the start function in the function index space.
	//
	// https://github.com/WebAssembly/design/blob/master/Modules.md#function-index-space
	Index uint32
}

// SectionElement defines element segments that initialize elements of imported
// or internally-defined tables with any other definition in the module.
//
// https://github.com/WebAssembly/design/blob/master/Modules.md#elements-section
type SectionElement struct {
	// Entries contains the elements.
	Entries []ElemSegment
}

// An ElemSegment is an element segment. It initializes a table with initial
// values.
type ElemSegment struct {
	// Index is the table index.
	Index uint32

	// Offset is an init expression (wasm bytecode) to compute the offset at
	// which to place the elements.
	Offset []byte

	// Elems contains the sequence of function indicies.
	Elems []uint32
}

// SectionCode contains a function body for every function in the module.
type SectionCode struct {
	// Bodies contains all function bodies.
	Bodies []FunctionBody
}

// A FunctionBody is the body of a function.
type FunctionBody struct {
	// Locals define the local variables of the function.
	Locals []LocalEntry

	// Code is the wasm bytecode of the function.
	Code []byte
}

// LocalEntry is a local variable in a function.
type LocalEntry struct {
	// Count specifies the number of the following type.
	Count uint32

	// Type is the type of the variable.
	Type valueType
}

// SectionData declares the initialized data that is loaded into the linear
// memory.
type SectionData struct {
	// Entries contains the data segment entries.
	Entries []DataSegment
}

// A DataSegment is a segment of data in the Data section that is loaded into
// linear memory.
type DataSegment struct {
	// Index is the linear memory index.
	//
	// https://github.com/WebAssembly/design/blob/master/Modules.md#linear-memory-index-space
	Index uint32

	// Offset is an init expression (wasm bytecode) that computes the offset to
	// place the data.
	Offset []byte

	// Data is the raw data to be placed in memory.
	Data []byte
}

// SectionName is a custom section that provides debugging information, by
// matching indices to human readable names.
type SectionName struct {
	// Name is the name of the name section. The value is always "name".
	Name string

	// Module is the name of the WASM module.
	Module string

	// Functions contains function name mappings.
	Functions *NameMap

	// Locals contains local function name mappings.
	Locals *Locals
}

// A NameMap is a map that maps an index to a name.
type NameMap struct {
	// Names contains a list of mappings in the NameMap.
	Names []Naming
}

// Naming is a single function naming. It maps an index to a human readable
// function name.
type Naming struct {
	// Index is the index that is being named.
	Index uint32

	// Name is a UTF-8 name.
	Name string
}

// Locals assigns name maps to a subset of functions in the function index
// space (imports and module-defined).
type Locals struct {
	// Funcs are the functions to be named.
	Funcs []LocalName
}

// LocalName a name mapping for a local function name.
type LocalName struct {
	// Index is the index of the function whose locals are being named.
	Index uint32

	// LocalMap is the name mapping for the function.
	LocalMap NameMap
}

// ExternalKind is set as the Kind for an import entry. The value specifies
// what type of import it is.
type ExternalKind uint8

const (
	// ExtKindFunction is an imported function.
	ExtKindFunction ExternalKind = iota

	// ExtKindTable is an imported table.
	ExtKindTable

	// ExtKindMemory is imported memory.
	ExtKindMemory

	// ExtKindGlobal is an imported global.
	ExtKindGlobal
)

// name types are used to identify the type in a Name section.
const (
	nameTypeModule   uint8 = iota // 0x00
	nameTypeFunction              // 0x01
	nameTypeLocal                 // 0x02
)

// varint7
type valueType int8

// varint7
type langType int8

// varint7
type elemType int8
