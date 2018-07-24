package wasm

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"
)

// magicnumber is a magic number which must appear as the very first bytes of a
// wasm file.
const magicnumber = 0x6d736100 // \0asm

// sectionID the id of a section in the wasm file.
type sectionID uint8

const (
	secCustom   sectionID = iota // 0x00
	secType                      // 0x01
	secImport                    // 0x02
	secFunction                  // 0x03
	secTable                     // 0x04
	secMemory                    // 0x05
	secGlobal                    // 0x06
	secExport                    // 0x07
	secStart                     // 0x08
	secElement                   // 0x09
	secCode                      // 0x0A
	secData                      // 0x0B
)

type parser struct {
	r *reader
}

// Parse parses the input to a WASM module.
func Parse(r io.Reader) (*Module, error) {
	p := &parser{
		r: newReader(r),
	}

	if err := p.parsePreamble(); err != nil {
		return nil, err
	}

	// Parse file sections
	var m Module
	for {
		s, err := p.parseSection()
		if err != nil {
			if strings.HasSuffix(err.Error(), io.EOF.Error()) {
				break
			}
			return nil, fmt.Errorf("[0x%06x] parse section: %v", p.r.Index(), err)
		}
		if s == nil {
			// Unknown section
			continue
		}
		m.Sections = append(m.Sections, &s)
	}
	return &m, nil
}

func (p *parser) parsePreamble() error {
	var h, v uint32
	if err := read(p.r, &h); err != nil {
		return fmt.Errorf("could not read file header")
	}
	if h != magicnumber {
		return fmt.Errorf("not a wasm file")
	}
	if err := read(p.r, &v); err != nil {
		return fmt.Errorf("could not version")
	}
	if v != 1 {
		return fmt.Errorf("unsupported version %d", v)
	}
	return nil
}

func (p *parser) parseSection() (interface{}, error) {
	sID, err := readByte(p.r)
	if err != nil {
		return nil, fmt.Errorf("read section id: %v", err)
	}

	switch sectionID(sID) {
	case secCustom:
		return p.parseCustomSection()
	case secType:
		return p.parseTypeSection()
	case secImport:
		return p.parseImportSection()
	case secFunction:
		return p.parseFunctionSection()
	case secTable:
		return p.parseTableSection()
	case secMemory:
		return p.parseMemorySection()
	case secGlobal:
		return p.parseGlobalSection()
	case secExport:
		return p.parseExportSection()
	case secStart:
		return p.parseStartSection()
	case secElement:
		return p.parseElementSection()
	case secCode:
		return p.parseCodeSection()
	case secData:
		return p.parseDataSection()
	default:
		var offset uint32
		if err := readVarUint32(p.r, &offset); err != nil {
			return nil, fmt.Errorf("read type section payload length: %v", err)
		}
		if _, err := io.CopyN(ioutil.Discard, p.r, int64(offset)); err != nil {
			return nil, fmt.Errorf("discard section payload, %d bytes: %v", offset, err)
		}
		if sID > byte(secData) {
			// This happens if the previous section was not read to the end,
			// indicating a bug in that section parser.
			return nil, fmt.Errorf("data corrupted; section id 0x%02x not valid", sID)
		}
		fmt.Printf("[0x%06x] skipping unknown section 0x%02x\n", p.r.Index(), sID)
		return nil, nil
	}
}

func (p *parser) parseCustomSection() (interface{}, error) {
	var pl uint32
	if err := readVarUint32(p.r, &pl); err != nil {
		return nil, fmt.Errorf("read section payload length: %v", err)
	}

	var nl uint32
	if err := readVarUint32(p.r, &nl); err != nil {
		return nil, fmt.Errorf("read section name length: %v", err)
	}

	b := make([]byte, nl)
	if err := read(p.r, &b); err != nil {
		return nil, fmt.Errorf("read section name: %v", err)
	}
	name := string(b)

	pl -= uint32(nl)                // sizeof name
	pl -= uint32(varUint32Size(nl)) // sizeof name_len

	if name == "name" {
		// A name section is a special custom section meant for debugging
		// purposes. It's defined in the spec so we'll parse it.
		return p.parseNameSection(name, pl)
	}

	s := SectionCustom{
		Name: name,
	}

	// set raw bytes
	s.Payload = make([]byte, pl)
	if err := read(p.r, s.Payload); err != nil {
		return nil, fmt.Errorf("read custom section payload: %v", err)
	}

	return &s, nil
}

func (p *parser) parseTypeSection() (interface{}, error) {
	var s SectionType

	err := p.parseMultiSection(func(uint32) error {
		var e typeEntry

		var form uint8
		if err := read(p.r, &form); err != nil {
			return fmt.Errorf("read form: %v", err)
		}
		e.Form = LanguageType(form)

		var pc uint32
		if err := readVarUint32(p.r, &pc); err != nil {
			return fmt.Errorf("read func param count: %v", err)
		}
		e.Params = make([]ValueType, pc)
		for i := range e.Params {
			if err := read(p.r, &e.Params[i]); err != nil {
				return fmt.Errorf("read function param type: %v", err)
			}
		}

		var rc uint8
		if err := read(p.r, &rc); err != nil {
			return fmt.Errorf("read number of returns from function: %v", err)
		}
		e.ReturnTypes = make([]ValueType, rc)
		for i := range e.ReturnTypes {
			if err := read(p.r, &e.ReturnTypes[i]); err != nil {
				return fmt.Errorf("read function return type: %v", err)
			}
		}

		s.Entries = append(s.Entries, e)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &s, nil
}

func (p *parser) parseImportSection() (interface{}, error) {
	var s SectionImport

	err := p.parseMultiSection(func(uint32) error {
		var e importEntry

		var moduleLen uint32
		if err := readVarUint32(p.r, &moduleLen); err != nil {
			return fmt.Errorf("read module length: %v", err)
		}

		modName := make([]byte, moduleLen)
		if err := read(p.r, modName); err != nil {
			return fmt.Errorf("read module name: %v", err)
		}
		e.Module = string(modName)

		var fieldLen uint32
		if err := readVarUint32(p.r, &fieldLen); err != nil {
			return fmt.Errorf("read field length: %v", err)
		}

		fieldName := make([]byte, fieldLen)
		if err := read(p.r, fieldName); err != nil {
			return fmt.Errorf("read field name")
		}
		e.Field = string(fieldName)

		var kind uint8
		if err := read(p.r, &kind); err != nil {
			return fmt.Errorf("read kind: %v", err)
		}
		e.Kind = ExternalKind(kind)

		switch e.Kind {
		case ExtKindFunction:
			e.FunctionType = &functionType{}
			if err := readVarUint32(p.r, &e.FunctionType.Index); err != nil {
				return fmt.Errorf("read function type index: %v", err)
			}
		case ExtKindTable:
			e.TableType = &tableType{}
			if err := read(p.r, &e.TableType.ElemType); err != nil {
				return fmt.Errorf("read table element type: %v", err)
			}
			if err := p.parseResizableLimits(&e.TableType.Limits); err != nil {
				return fmt.Errorf("read table resizable limits: %v", err)
			}
		case ExtKindMemory:
			e.MemoryType = &memoryType{}
			if err := p.parseResizableLimits(&e.MemoryType.Limits); err != nil {
				return fmt.Errorf("read memory resizable limits: %v", err)
			}
		case ExtKindGlobal:
			e.GlobalType = &globalType{}
			if err := read(p.r, &e.GlobalType.ContentType); err != nil {
				return fmt.Errorf("read global content type: %v", err)
			}
			if err := read(p.r, &e.GlobalType.Mutable); err != nil {
				return fmt.Errorf("read global mutability: %v", err)
			}
		}

		s.Entries = append(s.Entries, e)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &s, nil
}

func (p *parser) parseFunctionSection() (interface{}, error) {
	var s SectionFunction

	err := p.parseMultiSection(func(uint32) error {
		var t uint32
		if err := readVarUint32(p.r, &t); err != nil {
			return fmt.Errorf("read function type: %v", err)
		}

		s.Types = append(s.Types, t)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &s, nil
}

func (p *parser) parseTableSection() (interface{}, error) {
	var s SectionTable

	err := p.parseMultiSection(func(uint32) error {
		var e memoryType

		if err := p.parseResizableLimits(&e.Limits); err != nil {
			return fmt.Errorf("read memory resizable limits: %v", err)
		}

		s.Entries = append(s.Entries, e)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &s, nil
}

func (p *parser) parseMemorySection() (interface{}, error) {
	var s SectionMemory

	err := p.parseMultiSection(func(uint32) error {
		var e memoryType

		if err := p.parseResizableLimits(&e.Limits); err != nil {
			return fmt.Errorf("read memory resizable limits: %v", err)
		}

		s.Entries = append(s.Entries, e)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &s, nil
}

func (p *parser) parseGlobalSection() (interface{}, error) {
	var s SectionGlobal

	err := p.parseMultiSection(func(uint32) error {
		var e globalVariable

		if err := read(p.r, &e.Type.ContentType); err != nil {
			return fmt.Errorf("read global content type: %v", err)
		}

		if err := read(p.r, &e.Type.Mutable); err != nil {
			return fmt.Errorf("read global mutability: %v", err)
		}

		if err := readUntil(p.r, byte(opEnd), &e.Init); err != nil {
			return fmt.Errorf("read global init expression: %v", err)
		}

		s.Globals = append(s.Globals, e)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &s, nil
}

func (p *parser) parseExportSection() (interface{}, error) {
	var s SectionExport

	err := p.parseMultiSection(func(uint32) error {
		var e exportEntry

		var fl uint32
		if err := readVarUint32(p.r, &fl); err != nil {
			return fmt.Errorf("read field length: %v", err)
		}

		f := make([]byte, fl)
		if err := read(p.r, f); err != nil {
			return fmt.Errorf("read field")
		}
		e.Field = string(f)

		var kind uint8
		if err := read(p.r, &kind); err != nil {
			return fmt.Errorf("read kind: %v", err)
		}
		e.Kind = ExternalKind(kind)

		if err := readVarUint32(p.r, &e.Index); err != nil {
			return fmt.Errorf("read index: %v", err)
		}

		s.Entries = append(s.Entries, e)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &s, nil
}

func (p *parser) parseStartSection() (interface{}, error) {
	var s SectionStart

	if err := readVarUint32(p.r, &s.Index); err != nil {
		return nil, fmt.Errorf("read start index: %v", err)
	}

	return &s, nil
}

func (p *parser) parseElementSection() (interface{}, error) {
	var s SectionElement

	err := p.parseMultiSection(func(uint32) error {
		var e elemSegment

		if err := readVarUint32(p.r, &e.Index); err != nil {
			return fmt.Errorf("read element index: %v", err)
		}

		if err := readUntil(p.r, byte(opEnd), &e.Offset); err != nil {
			return fmt.Errorf("read offset expression: %v", err)
		}

		var numElem uint32
		if err := readVarUint32(p.r, &numElem); err != nil {
			return fmt.Errorf("read number of elements: %v", err)
		}
		e.Elems = make([]uint32, int(numElem))
		for i := range e.Elems {
			if err := readVarUint32(p.r, &e.Elems[i]); err != nil {
				return fmt.Errorf("read element function index %d: %v", i, err)
			}
		}

		s.Entries = append(s.Entries, e)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &s, nil
}

func (p *parser) parseCodeSection() (interface{}, error) {
	var s SectionCode

	err := p.parseMultiSection(func(length uint32) error {
		var e functionBody

		var bs uint32
		if err := readVarUint32(p.r, &bs); err != nil {
			return fmt.Errorf("read body size: %v", err)
		}

		end := p.r.Index() + int(bs)

		var localCount uint32
		if err := readVarUint32(p.r, &localCount); err != nil {
			return fmt.Errorf("read local count: %v", err)
		}
		e.Locals = make([]localEntry, localCount)

		for i := range e.Locals {
			var l localEntry

			if err := readVarUint32(p.r, &l.Count); err != nil {
				return fmt.Errorf("read local entry count: %v", err)
			}

			var t uint8
			if err := read(p.r, &t); err != nil {
				return fmt.Errorf("read local entry value type: %v", err)
			}
			l.Type = LanguageType(t)

			e.Locals[i] = l
		}

		numBytes := end - p.r.Index()
		e.Code = make([]byte, numBytes)
		if err := read(p.r, e.Code); err != nil {
			return fmt.Errorf("read function bytecode: %v", err)
		}

		s.Bodies = append(s.Bodies, e)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &s, nil
}

func (p *parser) parseDataSection() (interface{}, error) {
	var s SectionData

	err := p.parseMultiSection(func(length uint32) error {
		var e dataSegment

		if err := readVarUint32(p.r, &e.Index); err != nil {
			return fmt.Errorf("read data segment index: %v", err)
		}

		if err := readUntil(p.r, byte(opEnd), &e.Offset); err != nil {
			return fmt.Errorf("read data section offset initializer: %v", err)
		}

		var size uint32
		if err := readVarUint32(p.r, &size); err != nil {
			return fmt.Errorf("read data section size: %v", err)
		}

		e.Data = make([]byte, size)
		if err := read(p.r, e.Data); err != nil {
			return fmt.Errorf("read data section data: %v", err)
		}

		s.Entries = append(s.Entries, e)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &s, nil
}

func (p *parser) parseNameSection(name string, n uint32) (interface{}, error) {
	s := SectionName{
		Name: name,
	}

	var t uint8
	if err := read(p.r, &t); err != nil {
		return nil, fmt.Errorf("read name type: %v", err)
	}

	var pl uint32
	if err := readVarUint32(p.r, &pl); err != nil {
		return nil, fmt.Errorf("read payload length: %v", err)
	}

	switch NameType(t) {
	case NameTypeModule:
		var l uint32
		if err := readVarUint32(p.r, &l); err != nil {
			return nil, fmt.Errorf("read module name length: %v", err)
		}

		name := make([]byte, l)
		if err := read(p.r, name); err != nil {
			return nil, fmt.Errorf("read module name: %v", err)
		}

		s.Module = string(name)
	case NameTypeFunction:
		s.Functions = &nameMap{}
		if err := p.parseNameMap(s.Functions); err != nil {
			return nil, fmt.Errorf("read function name map: %v", err)
		}
	case NameTypeLocal:
		var count uint32
		if err := readVarUint32(p.r, &count); err != nil {
			return nil, fmt.Errorf("read local func name count: %v", err)
		}

		s.Locals = &locals{
			Funcs: make([]localName, count),
		}

		for i := range s.Locals.Funcs {
			var l localName
			if err := readVarUint32(p.r, &l.Index); err != nil {
				return nil, fmt.Errorf("read local func index: %v", err)
			}
			if err := p.parseNameMap(&l.LocalMap); err != nil {
				return nil, fmt.Errorf("read local name map: %v", err)
			}
			s.Locals.Funcs[i] = l
		}
	default:
		return nil, fmt.Errorf("unknown name section 0x%02x", t)
	}

	return &s, nil
}

func (p *parser) parseResizableLimits(l *resizableLimits) error {
	var hasMax bool
	if err := read(p.r, &hasMax); err != nil {
		return fmt.Errorf("flags: %v", err)
	}
	if err := readVarUint32(p.r, &l.Initial); err != nil {
		return fmt.Errorf("initial: %v", err)
	}
	if !hasMax {
		return nil
	}
	if err := readVarUint32(p.r, &l.Maximum); err != nil {
		return fmt.Errorf("maximum: %v", err)
	}
	return nil
}

// parseMultiSection reads the section payload length, then the number of
// entries in the section and calls the call back n times with the payload
// size. All sections except custom start with this pattern.
//
// If f returns an error, further processing is not done and the error is
// returned to the caller.
func (p *parser) parseMultiSection(f func(n uint32) error) error {
	var pl uint32
	if err := readVarUint32(p.r, &pl); err != nil {
		return fmt.Errorf("read type section payload length: %v", err)
	}

	var n uint32
	if err := readVarUint32(p.r, &n); err != nil {
		return fmt.Errorf("read section count: %v", err)
	}

	for i := uint32(0); i < n; i++ {
		if err := f(pl); err != nil {
			return fmt.Errorf("entry %d: %v", i, err)
		}
	}
	return nil
}

func (p *parser) parseNameMap(v *nameMap) error {
	var count uint32
	if err := readVarUint32(p.r, &count); err != nil {
		return fmt.Errorf("read name map count: %v", err)
	}

	v.Names = make([]naming, count)

	for i := range v.Names {
		var n naming

		if err := readVarUint32(p.r, &n.Index); err != nil {
			return fmt.Errorf("read naming index: %v", err)
		}

		var l uint32
		if err := readVarUint32(p.r, &l); err != nil {
			return fmt.Errorf("read naming length: %v", err)
		}

		name := make([]byte, l)
		if err := read(p.r, name); err != nil {
			return fmt.Errorf("read name: %v", err)
		}

		n.Name = string(name)

		v.Names[i] = n
	}
	return nil
}
