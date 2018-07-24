package wasm

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
)

const magicnumber = 0x6d736100 // \0asm

type state int

const (
	stateInitial state = iota
	stateBeginWASM
	stateEndSection
)

// A Parser parses wasm files.
//
// See: https://github.com/WebAssembly/design/blob/master/BinaryEncoding.md
type Parser struct {
	state state
	mod   Module
}

// A Module represents a WASM module.
type Module struct {
	Version  uint32     `json:"version,omitempty"`
	Sections []*Section `json:"sections,omitempty"`
}

// A Section is a single section in the WASM file.
type Section struct {
	ID      SectionID   `json:"id,omitempty"`
	Name    string      `json:"name,omitempty"`
	Payload interface{} `json:"payload,omitempty"`
}

// TypeEntry is an entry for a type definition.
type TypeEntry struct {
	Form        LangType `json:"form,omitempty"`
	ParamTypes  []OpCode `json:"param_types,omitempty"`
	ReturnTypes []OpCode `json:"return_types,omitempty"`
}

// ImportPayload is the payload for an Import section.
type ImportPayload struct {
	Entries []*ImportEntry `json:"entries,omitempty"`
}

// ImportEntry is an imported entry.
type ImportEntry struct {
	Module string       `json:"module,omitempty"`
	Field  string       `json:"field,omitempty"`
	Kind   ExternalKind `json:"kind,omitempty"`
	Type   interface{}  `json:"type,omitempty"`
}

// FunctionType is the Type field when Kind = ExtKindFunction.
type FunctionType struct {
	Index uint32 `json:"index,omitempty"`
}

// GlobalType is the Type field when Kind = ExtKindGlobal.
type GlobalType struct {
	ContentType OpCode `json:"content_type,omitempty"`
	Mutable     bool   `json:"mutable,omitempty"`
}

// TableType is the Type field when Kind = ExtKindTable.
type TableType struct {
	ElemType OpCode           `json:"elem_type,omitempty"`
	Limits   *ResizableLimits `json:"limits,omitempty"`
}

// MemoryType is the Type field when Kind = ExtKindMemory.
type MemoryType struct {
	Limits *ResizableLimits `json:"limits,omitempty"`
}

// ResizableLimits describes the limits of a table or memory.
type ResizableLimits struct {
	// Initial is the initial length of the memory.
	Initial uint32 `json:"initial,omitempty"`

	// Maximum is the maximum length of the memory. May not be set.
	Maximum uint32 `json:"maximum,omitempty"`
}

// A GlobalVariable is declared in the Globals section.
type GlobalVariable struct {
	Type GlobalType `json:"type,omitempty"`
	// Init is a wasm expression for setting the initial value of the variable.
	Init []OpCode `json:"init,omitempty"`
}

// An ExportEntry is an entry in the Exports section.
type ExportEntry struct {
	Field string       `json:"field,omitempty"`
	Kind  ExternalKind `json:"kind,omitempty"`
	Index uint32       `json:"index,omitempty"`
}

// An ElemSegment is a segment in the Element section.
type ElemSegment struct {
	// Index is the table index
	Index uint32 `json:"index"`
	// Offset is an expression that computes the offset to place the elements.
	Offset []OpCode `json:"offset"`
	// Elems is a sequence of function indices.
	Elems []uint32 `json:"elems,omitempty"`
}

// Parse parses the input to a WASM module.
func (p *Parser) Parse(rd io.Reader) (*Module, error) {
	r := bufio.NewReader(rd)
	for {
		if _, err := r.Peek(1); err != nil {
			if err == io.EOF {
				return &p.mod, nil
			}
		}
		switch p.state {
		case stateInitial:
			if err := p.readFilePreamble(r); err != nil {
				return nil, fmt.Errorf("read header: %v", err)
			}
		case stateBeginWASM, stateEndSection:
			if err := p.readSectionHeader(r); err != nil {
				return nil, fmt.Errorf("read section: %v", err)
			}
		}
	}
}

func (p *Parser) readFilePreamble(r io.Reader) error {
	var header uint32
	if err := read(r, &header); err != nil {
		return fmt.Errorf("reader header: %v", err)
	}
	if header != magicnumber {
		return fmt.Errorf("not a wasm file, expected header 0x%08x, got 0x%08x", magicnumber, header)
	}
	if err := read(r, &p.mod.Version); err != nil {
		return fmt.Errorf("read version: %v", err)
	}
	p.state = stateBeginWASM
	return nil
}

func (p *Parser) readSectionHeader(r io.Reader) error {
	var s Section
	var payloadLen, nameLen uint32
	if err := read(r, &s.ID); err != nil {
		return fmt.Errorf("read section id: %v", err)
	}
	if err := readVarUint32(r, &payloadLen); err != nil {
		return fmt.Errorf("read section payload length: %v", err)
	}
	if s.ID == 0 {
		if err := readVarUint32(r, &nameLen); err != nil {
			return fmt.Errorf("read section name length: %v", err)
		}
		name := make([]byte, nameLen)
		if err := read(r, &name); err != nil {
			return fmt.Errorf("read section name: %v", err)
		}
		s.Name = string(name)
	}

	payloadLen -= uint32(len(s.Name))            // sizeof name
	payloadLen -= uint32(varUint32Size(nameLen)) // sizeof name_len

	var err error

	switch s.ID {
	// case SectionCustom:
	// 	s.Payload = make([]byte, payloadLen)
	// 	if err := read(r, s.Payload); err != nil {
	// 		return fmt.Errorf("read custom section payload: %v", err)
	// 	}
	// case SectionType:
	// 	s.Payload, err = readTypePayload(r)
	// case SectionImport:
	// 	s.Payload, err = readImportPayload(r)
	// case SectionFunction:
	// 	s.Payload, err = readFunctionPayload(r)
	// case SectionTable:
	// 	s.Payload, err = readTablePayload(r)
	// case SectionMemory:
	// 	s.Payload, err = readMemoryPayload(r)
	// case SectionGlobal:
	// 	s.Payload, err = readGlobalPayload(r)
	// case SectionExport:
	// 	s.Payload, err = readExportPayload(r)
	// case SectionStart:
	// 	var index uint32
	// 	if err := readVarUint32(r, &index); err != nil {
	// 		return fmt.Errorf("read start section index: %v", err)
	// 	}
	// 	s.Payload = index
	case SectionElement:
		s.Payload, err = readElementSection(r)
	default:
		// Skip section
		offset := int64(payloadLen)
		if _, err := io.CopyN(ioutil.Discard, r, offset); err != nil {
			return fmt.Errorf("discard %s section payload, %d bytes: %v", s.ID, offset, err)
		}
	}

	if err != nil {
		return fmt.Errorf("read section %s: %v", s.ID, err)
	}

	p.mod.Sections = append(p.mod.Sections, &s)
	p.state = stateEndSection

	return nil
}

func readTypePayload(r io.Reader) (interface{}, error) {
	var count uint32
	if err := readVarUint32(r, &count); err != nil {
		return nil, fmt.Errorf("read section count: %v", err)
	}

	pl := make([]*TypeEntry, count)

	for i := uint32(0); i < count; i++ {
		var e TypeEntry

		form, err := readByte(r)
		if err != nil {
			return nil, fmt.Errorf("read form: %v", err)
		}
		e.Form = LangType(form)

		var paramCount uint32
		if err := readVarUint32(r, &paramCount); err != nil {
			return nil, fmt.Errorf("read func param count: %v", err)
		}

		e.ParamTypes = make([]OpCode, paramCount)
		for i := range e.ParamTypes {
			if err := readOpCode(r, &e.ParamTypes[i]); err != nil {
				return nil, fmt.Errorf("read function param %d: %v", i, err)
			}
		}

		var retCount uint8
		if err := readVarUint1(r, &retCount); err != nil {
			return nil, fmt.Errorf("read number of returns from function: %v", err)
		}

		e.ReturnTypes = make([]OpCode, retCount)
		for i := range e.ReturnTypes {
			if err := readOpCode(r, &e.ReturnTypes[i]); err != nil {
				return nil, fmt.Errorf("read function return type %d: %v", i, err)
			}
		}

		pl[i] = &e
	}

	return &pl, nil
}

func readImportPayload(r io.Reader) (interface{}, error) {
	var count uint32
	if err := readVarUint32(r, &count); err != nil {
		return nil, fmt.Errorf("read section count: %v", err)
	}

	pl := ImportPayload{
		Entries: make([]*ImportEntry, count),
	}

	for i := uint32(0); i < count; i++ {
		var e ImportEntry

		var moduleLen uint32
		if err := readVarUint32(r, &moduleLen); err != nil {
			return nil, fmt.Errorf("read module length: %v", err)
		}

		modName := make([]byte, moduleLen)
		if err := read(r, modName); err != nil {
			return nil, fmt.Errorf("read module name")
		}
		e.Module = string(modName)

		var fieldLen uint32
		if err := readVarUint32(r, &fieldLen); err != nil {
			return nil, fmt.Errorf("read field length: %v", err)
		}

		fieldName := make([]byte, fieldLen)
		if err := read(r, fieldName); err != nil {
			return nil, fmt.Errorf("read field name")
		}
		e.Field = string(fieldName)

		kind, err := readByte(r)
		if err != nil {
			return nil, fmt.Errorf("read kind: %v", err)
		}
		e.Kind = ExternalKind(kind)

		switch e.Kind {
		case ExtKindFunction:
			var t FunctionType
			if err := readVarUint32(r, &t.Index); err != nil {
				return nil, fmt.Errorf("read external function type index: %v", err)
			}
			e.Type = t
		case ExtKindTable:
			var t TableType
			if err := readOpCode(r, &t.ElemType); err != nil {
				return nil, fmt.Errorf("read table element type: %v", err)
			}
			limits, err := readResizableLimits(r)
			if err != nil {
				return nil, fmt.Errorf("read table resizable limits: %v", err)
			}
			t.Limits = limits
			e.Type = t
		case ExtKindMemory:
			var t MemoryType
			limits, err := readResizableLimits(r)
			if err != nil {
				return nil, fmt.Errorf("read memory resizable limits: %v", err)
			}
			t.Limits = limits
			e.Type = t
		case ExtKindGlobal:
			var t GlobalType
			if err := readOpCode(r, &t.ContentType); err != nil {
				return nil, fmt.Errorf("read global content type: %v", err)
			}
			if err := read(r, &t.Mutable); err != nil {
				return nil, fmt.Errorf("read global mutability: %v", err)
			}
		}

		pl.Entries[i] = &e
	}

	return &pl, nil
}

func readFunctionPayload(r io.Reader) (interface{}, error) {
	var count uint32
	if err := readVarUint32(r, &count); err != nil {
		return nil, fmt.Errorf("read section count: %v", err)
	}

	pl := make([]uint32, count)

	for i := uint32(0); i < count; i++ {
		if err := readVarUint32(r, &pl[i]); err != nil {
			return nil, err
		}
	}

	return &pl, nil
}

func readTablePayload(r io.Reader) (interface{}, error) {
	var count uint32
	if err := readVarUint32(r, &count); err != nil {
		return nil, fmt.Errorf("read section count: %v", err)
	}

	pl := make([]*TableType, count)

	for i := uint32(0); i < count; i++ {
		var t TableType
		if err := readOpCode(r, &t.ElemType); err != nil {
			return nil, fmt.Errorf("read table element type: %v", err)
		}
		limits, err := readResizableLimits(r)
		if err != nil {
			return nil, fmt.Errorf("read table resizable limits: %v", err)
		}
		t.Limits = limits
		pl[i] = &t
	}

	return &pl, nil
}

func readMemoryPayload(r io.Reader) (interface{}, error) {
	var count uint32
	if err := readVarUint32(r, &count); err != nil {
		return nil, fmt.Errorf("read section count: %v", err)
	}

	pl := make([]*MemoryType, count)

	for i := uint32(0); i < count; i++ {
		var t MemoryType
		limits, err := readResizableLimits(r)
		if err != nil {
			return nil, fmt.Errorf("read memory resizable limits: %v", err)
		}
		t.Limits = limits
		pl[i] = &t
	}

	return &pl, nil
}

func readGlobalPayload(r io.Reader) (interface{}, error) {
	var count uint32
	if err := readVarUint32(r, &count); err != nil {
		return nil, fmt.Errorf("read section count: %v", err)
	}

	pl := make([]*GlobalVariable, count)

	for i := uint32(0); i < count; i++ {
		var t GlobalVariable
		if err := readOpCode(r, &t.Type.ContentType); err != nil {
			return nil, fmt.Errorf("read global content type: %v", err)
		}
		if err := read(r, &t.Type.Mutable); err != nil {
			return nil, fmt.Errorf("read global mutability: %v", err)
		}
		if err := readExpr(r, &t.Init); err != nil {
			return nil, fmt.Errorf("read global init expression: %v", err)
		}
		pl[i] = &t
	}

	return &pl, nil
}

func readExportPayload(r io.Reader) (interface{}, error) {
	var count uint32
	if err := readVarUint32(r, &count); err != nil {
		return nil, fmt.Errorf("read section count: %v", err)
	}

	pl := make([]*ExportEntry, count)

	for i := uint32(0); i < count; i++ {
		var e ExportEntry

		var nameLen uint32
		if err := readVarUint32(r, &nameLen); err != nil {
			return nil, fmt.Errorf("read name length: %v", err)
		}

		name := make([]byte, nameLen)
		if err := read(r, name); err != nil {
			return nil, fmt.Errorf("read name")
		}
		e.Field = string(name)

		kind, err := readByte(r)
		if err != nil {
			return nil, fmt.Errorf("read kind: %v", err)
		}
		e.Kind = ExternalKind(kind)

		if err := readVarUint32(r, &e.Index); err != nil {
			return nil, fmt.Errorf("read index: %v", err)
		}

		pl[i] = &e
	}

	return &pl, nil
}

func readElementSection(r io.Reader) (interface{}, error) {
	var count uint32
	if err := readVarUint32(r, &count); err != nil {
		return nil, fmt.Errorf("read section count: %v", err)
	}

	pl := make([]*ElemSegment, count)

	for i := uint32(0); i < count; i++ {
		var e ElemSegment

		if err := readVarUint32(r, &e.Index); err != nil {
			return nil, fmt.Errorf("read element index: %v", err)
		}

		if err := readExpr(r, &e.Offset); err != nil {
			return nil, fmt.Errorf("read elemenet offset expression: %v", err)
		}

		var numElem uint32
		if err := readVarUint32(r, &numElem); err != nil {
			return nil, fmt.Errorf("read number of elements: %v", err)
		}
		e.Elems = make([]uint32, int(numElem))
		for i := range e.Elems {
			if err := readVarUint32(r, &e.Elems[i]); err != nil {
				return nil, fmt.Errorf("read element function index %d: %v", i, err)
			}
		}

		pl[i] = &e
	}

	return &pl, nil
}

func readResizableLimits(r io.Reader) (*ResizableLimits, error) {
	var l ResizableLimits

	var hasMax bool
	if err := read(r, &hasMax); err != nil {
		return nil, fmt.Errorf("flags: %v", err)
	}
	if err := readVarUint32(r, &l.Initial); err != nil {
		return nil, fmt.Errorf("initial: %v", err)
	}
	if hasMax {
		if err := readVarUint32(r, &l.Maximum); err != nil {
			return nil, fmt.Errorf("maximum: %v", err)
		}
	}

	return &l, nil
}

func readExpr(r io.Reader, v *[]OpCode) error {
	for {
		b, err := readByte(r)
		if err != nil {
			return err
		}
		if b == byte(opEnd) {
			break
		}
		*v = append(*v, OpCode(b))
	}
	return nil
}
