//go:generate stringer -trimprefix sec -type sectionID

package wasm

import (
	"fmt"
	"io"
	"io/ioutil"
)

// magicnumber is a magic number which must appear as the very first bytes of a
// wasm file.
const magicnumber = 0x6d736100 // \0asm

// opEnd is the op code for a section end
const opEnd = 0x0b

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

var errDone = fmt.Errorf("done")

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
		err := p.parseSection(&m.Sections)
		if err != nil {
			if err == errDone {
				break
			}
			return nil, fmt.Errorf("[0x%06x] parse section: %v", p.r.Index(), err)
		}
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

func (p *parser) parseSection(ss *[]Section) error {
	var i uint8
	if err := readVarUint7(p.r, &i); err != nil {
		if err == io.EOF {
			return errDone
		}
		return fmt.Errorf("read section id: %v", err)
	}
	sid := sectionID(i)

	var s Section
	var err error

	base := &section{
		id:   sid,
		name: sid.String(),
	}

	if err := readVarUint32(p.r, &base.size); err != nil {
		return fmt.Errorf("read type section payload length: %v", err)
	}

	switch sid {
	case secCustom:
		s, err = p.parseCustomSection(base)
	case secType:
		s, err = p.parseTypeSection(base)
	case secImport:
		s, err = p.parseImportSection(base)
	case secFunction:
		s, err = p.parseFunctionSection(base)
	case secTable:
		s, err = p.parseTableSection(base)
	case secMemory:
		s, err = p.parseMemorySection(base)
	case secGlobal:
		s, err = p.parseGlobalSection(base)
	case secExport:
		s, err = p.parseExportSection(base)
	case secStart:
		s, err = p.parseStartSection(base)
	case secElement:
		s, err = p.parseElementSection(base)
	case secCode:
		s, err = p.parseCodeSection(base)
	case secData:
		s, err = p.parseDataSection(base)
	default:
		if _, err := io.CopyN(ioutil.Discard, p.r, int64(base.size)); err != nil {
			return fmt.Errorf("discard section payload, %d bytes: %v", base.size, err)
		}
		if sid > secData {
			// This happens if the previous section was not read to the end,
			// indicating a bug in that section parser.
			return fmt.Errorf("data corrupted; section id 0x%02x not valid", sid)
		}
		// Skip unknown section
		return nil
	}
	if err != nil {
		return err
	}

	if s != nil {
		*ss = append(*ss, s)
	}

	return nil
}

func (p *parser) parseCustomSection(base *section) (Section, error) {
	var nl uint32
	if err := readVarUint32(p.r, &nl); err != nil {
		return nil, fmt.Errorf("read section name length: %v", err)
	}

	b := make([]byte, nl)
	if err := read(p.r, &b); err != nil {
		return nil, fmt.Errorf("read section name: %v", err)
	}
	name := string(b)

	base.size -= uint32(nl)                // sizeof name
	base.size -= uint32(varUint32Size(nl)) // sizeof name_len

	if name == "name" {
		// A name section is a special custom section meant for debugging
		// purposes. It's defined in the spec so we'll parse it.
		return p.parseNameSection(base, name, base.size)
	}

	s := SectionCustom{
		section:     base,
		SectionName: name,
	}

	// set raw bytes
	s.Payload = make([]byte, base.size)
	if err := read(p.r, s.Payload); err != nil {
		return nil, fmt.Errorf("read custom section payload: %v", err)
	}

	return &s, nil
}

func (p *parser) parseTypeSection(base *section) (*SectionType, error) {
	s := SectionType{section: base}

	err := p.loopCount(func() error {
		var e FuncType

		if err := readVarInt7(p.r, &e.Form); err != nil {
			return fmt.Errorf("read form: %v", err)
		}

		p.loopCount(func() error {
			var param int8
			if err := readVarInt7(p.r, &param); err != nil {
				return fmt.Errorf("read function param type: %v", err)
			}
			e.Params = append(e.Params, param)
			return nil
		})

		var rc uint8
		if err := readVarUint1(p.r, &rc); err != nil {
			return fmt.Errorf("read number of returns from function: %v", err)
		}
		e.ReturnTypes = make([]int8, rc)
		for i := range e.ReturnTypes {
			if err := readVarInt7(p.r, &e.ReturnTypes[i]); err != nil {
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

func (p *parser) parseImportSection(base *section) (*SectionImport, error) {
	s := SectionImport{section: base}

	err := p.loopCount(func() error {
		var e ImportEntry

		var ml uint32
		if err := readVarUint32(p.r, &ml); err != nil {
			return fmt.Errorf("read module length: %v", err)
		}

		mn := make([]byte, ml)
		if err := read(p.r, mn); err != nil {
			return fmt.Errorf("read module name: %v", err)
		}
		e.Module = string(mn)

		var fl uint32
		if err := readVarUint32(p.r, &fl); err != nil {
			return fmt.Errorf("read field length: %v", err)
		}

		fn := make([]byte, fl)
		if err := read(p.r, fn); err != nil {
			return fmt.Errorf("read field name")
		}
		e.Field = string(fn)

		var kind uint8
		if err := read(p.r, &kind); err != nil {
			return fmt.Errorf("read kind: %v", err)
		}
		e.Kind = ExternalKind(kind)

		switch e.Kind {
		case ExtKindFunction:
			e.FunctionType = &FunctionType{}
			if err := readVarUint32(p.r, &e.FunctionType.Index); err != nil {
				return fmt.Errorf("read function type index: %v", err)
			}
		case ExtKindTable:
			e.TableType = &TableType{}
			if err := readVarInt7(p.r, &e.TableType.ElemType); err != nil {
				return fmt.Errorf("read table element type: %v", err)
			}

			if err := p.parseResizableLimits(&e.TableType.Limits); err != nil {
				return fmt.Errorf("read table resizable limits: %v", err)
			}
		case ExtKindMemory:
			e.MemoryType = &MemoryType{}
			if err := p.parseResizableLimits(&e.MemoryType.Limits); err != nil {
				return fmt.Errorf("read memory resizable limits: %v", err)
			}
		case ExtKindGlobal:
			e.GlobalType = &GlobalType{}
			if err := readVarInt7(p.r, &e.GlobalType.ContentType); err != nil {
				return fmt.Errorf("read global content type: %v", err)
			}

			var m uint8
			if err := readVarUint1(p.r, &m); err != nil {
				return fmt.Errorf("read global mutability: %v", err)
			}
			e.GlobalType.Mutable = m == 1
		}

		s.Entries = append(s.Entries, e)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &s, nil
}

func (p *parser) parseFunctionSection(base *section) (*SectionFunction, error) {
	s := SectionFunction{section: base}

	err := p.loopCount(func() error {
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

func (p *parser) parseTableSection(base *section) (*SectionTable, error) {
	s := SectionTable{section: base}

	err := p.loopCount(func() error {
		var e TableType

		if err := p.parseTableType(&e); err != nil {
			return fmt.Errorf("read table type: %v", err)
		}

		s.Entries = append(s.Entries, e)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &s, nil
}

func (p *parser) parseMemorySection(base *section) (*SectionMemory, error) {
	s := SectionMemory{section: base}

	err := p.loopCount(func() error {
		var e MemoryType

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

func (p *parser) parseGlobalSection(base *section) (*SectionGlobal, error) {
	s := SectionGlobal{section: base}

	err := p.loopCount(func() error {
		var e GlobalVariable

		if err := readVarInt7(p.r, &e.Type.ContentType); err != nil {
			return fmt.Errorf("read global content type: %v", err)
		}

		if err := read(p.r, &e.Type.Mutable); err != nil {
			return fmt.Errorf("read global mutability: %v", err)
		}

		if err := readUntil(p.r, opEnd, &e.Init); err != nil {
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

func (p *parser) parseExportSection(base *section) (*SectionExport, error) {
	s := SectionExport{section: base}

	err := p.loopCount(func() error {
		var e ExportEntry

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
		if err := readVarUint7(p.r, &kind); err != nil {
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

func (p *parser) parseStartSection(base *section) (*SectionStart, error) {
	s := SectionStart{section: base}

	if err := readVarUint32(p.r, &s.Index); err != nil {
		return nil, fmt.Errorf("read start index: %v", err)
	}

	return &s, nil
}

func (p *parser) parseElementSection(base *section) (*SectionElement, error) {
	s := SectionElement{section: base}

	err := p.loopCount(func() error {
		var e ElemSegment

		if err := readVarUint32(p.r, &e.Index); err != nil {
			return fmt.Errorf("read element index: %v", err)
		}

		if err := readUntil(p.r, opEnd, &e.Offset); err != nil {
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

func (p *parser) parseCodeSection(base *section) (*SectionCode, error) {
	s := SectionCode{section: base}

	err := p.loopCount(func() error {
		var e FunctionBody

		var bs uint32
		if err := readVarUint32(p.r, &bs); err != nil {
			return fmt.Errorf("read body size: %v", err)
		}

		end := p.r.Index() + int(bs)

		p.loopCount(func() error {
			var l LocalEntry

			if err := readVarUint32(p.r, &l.Count); err != nil {
				return fmt.Errorf("read local entry count: %v", err)
			}
			if err := read(p.r, &l.Type); err != nil {
				return fmt.Errorf("read local entry value type: %v", err)
			}

			e.Locals = append(e.Locals, l)

			return nil
		})

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

func (p *parser) parseDataSection(base *section) (*SectionData, error) {
	s := SectionData{section: base}

	err := p.loopCount(func() error {
		var e DataSegment

		if err := readVarUint32(p.r, &e.Index); err != nil {
			return fmt.Errorf("read data segment index: %v", err)
		}

		if err := readUntil(p.r, opEnd, &e.Offset); err != nil {
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

// name types are used to identify the type in a Name section.
const (
	nameTypeModule   uint8 = iota // 0x00
	nameTypeFunction              // 0x01
	nameTypeLocal                 // 0x02
)

func (p *parser) parseNameSection(base *section, name string, n uint32) (*SectionName, error) {
	s := SectionName{
		section:     base,
		SectionName: name,
	}

	var t uint8
	if err := read(p.r, &t); err != nil {
		return nil, fmt.Errorf("read name type: %v", err)
	}

	var pl uint32
	if err := readVarUint32(p.r, &pl); err != nil {
		return nil, fmt.Errorf("read payload length: %v", err)
	}

	switch t {
	case nameTypeModule:
		var l uint32
		if err := readVarUint32(p.r, &l); err != nil {
			return nil, fmt.Errorf("read module name length: %v", err)
		}

		name := make([]byte, l)
		if err := read(p.r, name); err != nil {
			return nil, fmt.Errorf("read module name: %v", err)
		}

		s.Module = string(name)
	case nameTypeFunction:
		s.Functions = &NameMap{}
		if err := p.parseNameMap(s.Functions); err != nil {
			return nil, fmt.Errorf("read function name map: %v", err)
		}
	case nameTypeLocal:
		s.Locals = &Locals{}
		p.loopCount(func() error {
			var l LocalName
			if err := readVarUint32(p.r, &l.Index); err != nil {
				return fmt.Errorf("read local func index: %v", err)
			}
			if err := p.parseNameMap(&l.LocalMap); err != nil {
				return fmt.Errorf("read local name map: %v", err)
			}
			s.Locals.Funcs = append(s.Locals.Funcs, l)
			return nil
		})
	default:
		return nil, fmt.Errorf("unknown name type 0x%02x", t)
	}

	return &s, nil
}

func (p *parser) parseResizableLimits(l *ResizableLimits) error {
	var hasMax uint8
	if err := readVarUint1(p.r, &hasMax); err != nil {
		return fmt.Errorf("flags: %v", err)
	}
	if err := readVarUint32(p.r, &l.Initial); err != nil {
		return fmt.Errorf("initial: %v", err)
	}
	if hasMax == 0 {
		return nil
	}
	if err := readVarUint32(p.r, &l.Maximum); err != nil {
		return fmt.Errorf("maximum: %v", err)
	}
	return nil
}

func (p *parser) parseTableType(t *TableType) error {
	refType, err := readByte(p.r)
	if err != nil {
		return fmt.Errorf("read table type limits: %v", err)
	}
	t.ElemType = int8(refType)
	if err := p.parseResizableLimits(&t.Limits); err != nil {
		return fmt.Errorf("read memory resizable limits: %v", err)
	}
	return nil
}

// loopCount reads a varuint32 count and and calls the f n times. All sections
// except custom start with this pattern.
//
// If f returns an error, further processing is not done and the error is
// returned to the caller.
func (p *parser) loopCount(f func() error) error {
	var n uint32
	if err := readVarUint32(p.r, &n); err != nil {
		return fmt.Errorf("read section count: %v", err)
	}

	for i := uint32(0); i < n; i++ {
		if err := f(); err != nil {
			return fmt.Errorf("entry %d: %v", i, err)
		}
	}

	return nil
}

func (p *parser) parseNameMap(v *NameMap) error {
	p.loopCount(func() error {
		var n Naming

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
		v.Names = append(v.Names, n)

		return nil
	})

	return nil
}
