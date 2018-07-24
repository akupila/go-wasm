package wasm

import (
	"encoding/binary"
	"io"
)

func read(r io.Reader, v interface{}) error {
	return binary.Read(r, binary.LittleEndian, v)
}

func readByte(r io.Reader) (byte, error) {
	b := make([]byte, 1)
	if _, err := r.Read(b); err != nil {
		return 0, err
	}
	return b[0], nil
}

func readVarUint1(r io.Reader, v *uint8) error {
	return binary.Read(r, binary.LittleEndian, v)
}

func readVarUint7(r io.Reader, v *uint8) error {
	if err := binary.Read(r, binary.LittleEndian, v); err != nil {
		return err
	}
	*v &= 0xFE
	return nil
}

func readVarUint32(r io.Reader, v *uint32) error {
	var shift uint32
	for {
		b, err := readByte(r)
		if err != nil {
			return err
		}
		*v |= uint32(b&0x7F) << shift
		if (b & 0x80) == 0 {
			break
		}
		shift += 7
	}

	return nil
}

func readVarInt1(r io.Reader, v *int8) error {
	return binary.Read(r, binary.LittleEndian, v)
}

func readVarInt7(r io.Reader, v *int8) error {
	if err := binary.Read(r, binary.LittleEndian, v); err != nil {
		return err
	}
	*v &= 0x7F
	return nil
}

func readVarInt32(r io.Reader, v *int32) error {
	var shift uint32
	for {
		b, err := readByte(r)
		if err != nil {
			return err
		}
		*v |= int32(b&0x7F) << shift
		if (b & 0x80) == 0 {
			break
		}
		shift += 7
	}

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

func readOpCode(r io.Reader, v *OpCode) error {
	var t int8
	if err := readVarInt7(r, &t); err != nil {
		return err
	}
	*v = OpCode(t)
	return nil
}
