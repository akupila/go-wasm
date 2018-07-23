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

type Parser struct {
	state state
	mod   Module
}

type Section struct {
	ID      uint8
	Name    string
	Payload []byte
}

type Module struct {
	Version  uint32
	Sections []*Section
}

// See: https://github.com/WebAssembly/design/blob/master/BinaryEncoding.md
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
			p.readSectionHeader(r)
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
	s := Section{}
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
	s.Payload = make([]byte, payloadLen)
	if err := binary.Read(r, binary.LittleEndian, s.Payload); err != nil {
		return fmt.Errorf("read section payload: %v", err)
	}
	fmt.Println(s.ID, len(s.Payload))
	p.mod.Sections = append(p.mod.Sections, &s)
	p.state = stateEndSection
	return nil
}

// varUint32Size returns the size in bytes of a varuint32
func varUint32Size(v uint32) int {
	s := 0
	for v > 0 {
		s++
		v = v >> 8
	}
	return s
}

func readVarUint1(r io.Reader, v *uint8) error {
	return binary.Read(r, binary.LittleEndian, v)
}

func readVarUint7(r io.Reader, v *uint8) error {
	if err := binary.Read(r, binary.LittleEndian, &v); err != nil {
		return err
	}
	*v &= 0xFe
	return nil
}

func readVarUint32(r io.ByteReader, v *uint32) error {
	var shift uint32
	for {
		b, err := r.ReadByte()
		if err != nil {
			return err
		}
		*v |= uint32(b&0x7F) << shift
		shift += 7
		if (b & 0x80) == 0 {
			break
		}
	}

	return nil
}
