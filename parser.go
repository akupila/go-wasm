package wasm

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
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

// TypePayload is the payload for a Type section.
type TypePayload struct {
	Entries []*TypeEntry `json:"entries,omitempty"`
}

// TypeEntry is an entry for a type definition.
type TypeEntry struct {
	Form        OpCode   `json:"form,omitempty"`
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

// GlobalType is the Type field when Kind = ExtKindTable.
type TableType struct {
	ElemType OpCode           `json:"elem_type,omitempty"`
	Limits   *ResizableLimits `json:"limits,omitempty"`
}

// GlobalType is the Type field when Kind = ExtKindMemory.
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

func (p *Parser) readFilePreamble(r *bufio.Reader) error {
	var header uint32
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return fmt.Errorf("reader header: %v", err)
	}
	if header != magicnumber {
		return fmt.Errorf("not a wasm file, expected header 0x%x, got 0x%x", magicnumber, header)
	}
	if err := binary.Read(r, binary.LittleEndian, &p.mod.Version); err != nil {
		return fmt.Errorf("read version: %v", err)
	}
	p.state = stateBeginWASM
	return nil
}

func (p *Parser) readSectionHeader(r *bufio.Reader) error {
	var s Section
	var payloadLen, nameLen uint32
	if err := binary.Read(r, binary.LittleEndian, &s.ID); err != nil {
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
		if err := binary.Read(r, binary.LittleEndian, &name); err != nil {
			return fmt.Errorf("read section name: %v", err)
		}
		s.Name = string(name)
	}

	payloadLen -= uint32(len(s.Name))            // sizeof name
	payloadLen -= uint32(varUint32Size(nameLen)) // sizeof name_len

	var err error

	switch s.ID {
	case SectionCustom:
		s.Payload = make([]byte, payloadLen)
		if err := binary.Read(r, binary.LittleEndian, s.Payload); err != nil {
			return fmt.Errorf("read section payload: %v", err)
		}
	case SectionType:
		s.Payload, err = readTypePayload(r)
	case SectionImport:
		s.Payload, err = readImportPayload(r)
	default:
		// Skip section
		if _, err := r.Discard(int(payloadLen)); err != nil {
			return fmt.Errorf("discard section payload: %v", err)
		}
	}

	if err != nil {
		return fmt.Errorf("read section %s: %v", s.ID, err)
	}

	p.mod.Sections = append(p.mod.Sections, &s)
	p.state = stateEndSection

	return nil
}

func readTypePayload(r *bufio.Reader) (interface{}, error) {
	var count uint32
	if err := readVarUint32(r, &count); err != nil {
		return nil, fmt.Errorf("read section count: %v", err)
	}

	pl := TypePayload{
		Entries: make([]*TypeEntry, count),
	}

	for i := uint32(0); i < count; i++ {
		var e TypeEntry

		if err := readOpCode(r, &e.Form); err != nil {
			return nil, fmt.Errorf("read form: %v", err)
		}

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

		pl.Entries[i] = &e
	}

	return &pl, nil
}

func readImportPayload(r *bufio.Reader) (interface{}, error) {
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
		if err := binary.Read(r, binary.LittleEndian, modName); err != nil {
			return nil, fmt.Errorf("read module name")
		}
		e.Module = string(modName)

		var fieldLen uint32
		if err := readVarUint32(r, &fieldLen); err != nil {
			return nil, fmt.Errorf("read field length: %v", err)
		}

		fieldName := make([]byte, fieldLen)
		if err := binary.Read(r, binary.LittleEndian, fieldName); err != nil {
			return nil, fmt.Errorf("read field name")
		}
		e.Field = string(fieldName)

		kind, err := r.ReadByte()
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
			if err := binary.Read(r, binary.LittleEndian, &t.Mutable); err != nil {
				return nil, fmt.Errorf("read global mutability: %v", err)
			}
		}

		pl.Entries[i] = &e
	}

	return &pl, nil
}

func readOpCode(r *bufio.Reader, v *OpCode) error {
	var t int8
	if err := readVarInt7(r, &t); err != nil {
		return err
	}
	*v = OpCode(t)
	return nil
}

func readResizableLimits(r *bufio.Reader) (*ResizableLimits, error) {
	var l ResizableLimits

	var hasMax bool
	if err := binary.Read(r, binary.LittleEndian, &hasMax); err != nil {
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
