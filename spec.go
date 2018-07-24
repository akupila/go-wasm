//go:generate stringer -type SectionID -trimprefix Section
//go:generate stringer -type ExternalKind -trimprefix ExtKind
//go:generate stringer -type LangType -trimprefix LangType
//go:generate stringer -type OpCode -trimprefix op
//go:generate stringer -type NameType -trimprefix NameType

package wasm

import (
	"fmt"
)

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

// LangType the type for an entry in the Type section.
type LangType int

const (
	// LangTypeBlock is the language type for a pseudo type representing an empty block_type.
	LangTypeBlock LangType = 0x40

	// LangTypeFunc is the language type for a function.
	LangTypeFunc LangType = 0x60

	// LangTypeAnyFunc is the language type for a function with any signatrue.
	LangTypeAnyFunc LangType = 0x70

	// LangTypeFloat64 is the language type for a 64 bit float.
	LangTypeFloat64 LangType = 0x7C

	// LangTypeFloat32 is the language type for a 32 bit float.
	LangTypeFloat32 LangType = 0x7D

	// LangTypeInt64 is the language type for a 64 bit signed integer.
	LangTypeInt64 LangType = 0x7E

	// LangTypeInt32 is the language type for a 32 bit signed integer.
	LangTypeInt32 LangType = 0x7F
)

func (l LangType) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s (0x%02x)"`, l.String(), byte(l))), nil
}

// OpCode is an operation code.
type OpCode uint8

const (
	opUnreachable       OpCode = iota // 0x00
	opNop                             // 0x01
	opBlock                           // 0x02
	opLoop                            // 0x03
	opIf                              // 0x04
	opElse                            // 0x05
	_                                 // 0x06
	_                                 // 0x07
	_                                 // 0x08
	_                                 // 0x09
	_                                 // 0x0A
	opEnd                             // 0x0B
	opBr                              // 0x0C
	opBrIf                            // 0x0D
	opBrTable                         // 0x0E
	opReturn                          // 0x0F
	opCall                            // 0x10
	opCallIndirect                    // 0x11
	_                                 // 0x12
	_                                 // 0x13
	_                                 // 0x14
	_                                 // 0x15
	_                                 // 0x16
	_                                 // 0x17
	_                                 // 0x18
	_                                 // 0x19
	opDrop                            // 0x1A
	opSelect                          // 0x1B
	_                                 // 0x1C
	_                                 // 0x1D
	_                                 // 0x1E
	_                                 // 0x1F
	opGetLocal                        // 0x20
	opSetLocal                        // 0x21
	opTeeLocal                        // 0x22
	opGetGlobal                       // 0x23
	opSetGlobal                       // 0x24
	_                                 // 0x25
	_                                 // 0x26
	_                                 // 0x27
	opI32Load                         // 0x28
	opI64Load                         // 0x29
	opF32Load                         // 0x2A
	opF64Load                         // 0x2B
	opI32Load8S                       // 0x2C
	opI32Load8U                       // 0x2D
	opI32Load16S                      // 0x2E
	opI32Load16U                      // 0x2F
	opI64Load8S                       // 0x30
	opI64Load8U                       // 0x31
	opI64Load16S                      // 0x32
	opI64Load16U                      // 0x33
	opI64Load32S                      // 0x34
	opI64Load32U                      // 0x35
	opI32Store                        // 0x36
	opI64Store                        // 0x37
	opF32Store                        // 0x38
	opF64Store                        // 0x39
	opI32Store8                       // 0x3A
	opI32Store16                      // 0x3B
	opI64Store8                       // 0x3C
	opI64Store16                      // 0x3D
	opI64Store32                      // 0x3E
	opCurrentMemory                   // 0x3F
	opGrowMemory                      // 0x40
	opI32Const                        // 0x41
	opI64Const                        // 0x42
	opF32Const                        // 0x43
	opF64Const                        // 0x44
	opI32Eqz                          // 0x45
	opI32Eq                           // 0x46
	opI32Ne                           // 0x47
	opI32LtS                          // 0x48
	opI32LtU                          // 0x49
	opI32GtS                          // 0x4A
	opI32GtU                          // 0x4B
	opI32LeS                          // 0x4C
	opI32LeU                          // 0x4D
	opI32GeS                          // 0x4E
	opI32GeU                          // 0x4F
	opI64Eqz                          // 0x50
	opI64Eq                           // 0x51
	opI64Ne                           // 0x52
	opI64LtS                          // 0x53
	opI64LtU                          // 0x54
	opI64GtS                          // 0x55
	opI64GtU                          // 0x56
	opI64LeS                          // 0x57
	opI64LeU                          // 0x58
	opI64GeS                          // 0x59
	opI64GeU                          // 0x5A
	opF32Eq                           // 0x5B
	opF32Ne                           // 0x5C
	opF32Lt                           // 0x5D
	opF32Gt                           // 0x5E
	opF32Le                           // 0x5F
	opF32Ge                           // 0x60
	opF64Eq                           // 0x61
	opF64Ne                           // 0x62
	opF64Lt                           // 0x63
	opF64Gt                           // 0x64
	opF64Le                           // 0x65
	opF64Ge                           // 0x66
	opI32Clz                          // 0x67
	opI32Ctz                          // 0x68
	opI32Popcnt                       // 0x69
	opI32Add                          // 0x6A
	opI32Sub                          // 0x6B
	opI32Mul                          // 0x6C
	opI32DivS                         // 0x6D
	opI32DivU                         // 0x6E
	opI32Rems                         // 0x6F
	opI32Remu                         // 0x70
	opI32And                          // 0x71
	opI32Or                           // 0x72
	opI32Xor                          // 0x73
	opI32Shl                          // 0x74
	opI32ShrS                         // 0x75
	opI32ShrU                         // 0x76
	opI32Rotl                         // 0x77
	opI32Rotr                         // 0x78
	opI64Clz                          // 0x79
	opI64Ctz                          // 0x7A
	opI64Popcnt                       // 0x7B
	opI64Add                          // 0x7C
	opI64Sub                          // 0x7D
	opI64Mul                          // 0x7E
	opI64DivS                         // 0x7F
	opI64DivU                         // 0x80
	opI64RemS                         // 0x81
	opI64RemU                         // 0x82
	opI64And                          // 0x83
	opI64Or                           // 0x84
	opI64Xor                          // 0x85
	opI64Shl                          // 0x86
	opI64ShrS                         // 0x87
	opI64ShrU                         // 0x88
	opI64Rotl                         // 0x89
	opI64Rotr                         // 0x8A
	opF32Abs                          // 0x8B
	opF32Neg                          // 0x8C
	opF32Ceil                         // 0x8D
	opF32Floor                        // 0x8E
	opF32Trunc                        // 0x8F
	opF32Nearest                      // 0x90
	opF32Sqrt                         // 0x91
	opF32Add                          // 0x92
	opF32Sub                          // 0x93
	opF32Mul                          // 0x94
	opF32Div                          // 0x95
	opF32Min                          // 0x96
	opF32Max                          // 0x97
	opF32Copysign                     // 0x98
	opF64Abs                          // 0x99
	opF64Neg                          // 0x9A
	opF64Ceil                         // 0x9B
	opF64Floor                        // 0x9C
	opF64Trunc                        // 0x9D
	opF64Nearest                      // 0x9E
	opF64Sqrt                         // 0x9F
	opF64Add                          // 0xA0
	opF64Sub                          // 0xA1
	opF64Mul                          // 0xA2
	opF64Div                          // 0xA3
	opF64Min                          // 0xA4
	opF64Max                          // 0xA5
	opF64Copysign                     // 0xA6
	opI32WrapI64                      // 0xA7
	opI32TruncSF32                    // 0xA8
	opI32TruncUF32                    // 0xA9
	opI32TruncSF64                    // 0xAA
	opI32TruncUF64                    // 0xAB
	opI64ExtendSI32                   // 0xAC
	opI64ExtendUI32                   // 0xAD
	opI64TruncSF32                    // 0xAE
	opI64TruncUF32                    // 0xAF
	opI64TruncSF64                    // 0xB0
	opI64TruncUF64                    // 0xB1
	opF32ConvertSI32                  // 0xB2
	opF32ConvertUI32                  // 0xB3
	opF32ConvertSI64                  // 0xB4
	opF32ConvertUI64                  // 0xB5
	opF32DemoteF64                    // 0xB6
	opF64ConvertSI32                  // 0xB7
	opF64ConvertUI32                  // 0xB8
	opF64ConvertSI64                  // 0xB9
	opF64ConvertUI64                  // 0xBA
	opF64PromoteF32                   // 0xBB
	opI32ReinterpretF32               // 0xBC
	opI64ReinterpretF64               // 0xBD
	opF32ReinterpretI32               // 0xBE
	opF64ReinterpretI64               // 0xBF
)

func (o OpCode) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s (0x%02x)"`, o.String(), byte(o))), nil
}

type NameType uint8

const (
	NameTypeModule NameType = iota
	NameTypeFunction
	NameTypeLocal
)

func (n NameType) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s (0x%02x)"`, n.String(), byte(n))), nil
}
