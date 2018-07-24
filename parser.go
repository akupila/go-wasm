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

func readOpCode(r *bufio.Reader, v *OpCode) error {
	var t int8
	if err := readVarInt7(r, &t); err != nil {
		return err
	}
	*v = OpCode(t)
	return nil
}
