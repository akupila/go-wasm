package wasm

import (
	"bufio"
	"encoding/binary"
	"io"
)

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

func readVarUint32(r *bufio.Reader, v *uint32) error {
	var shift uint32
	for {
		b, err := r.ReadByte()
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

func readVarInt32(r *bufio.Reader, v *int32) error {
	var shift uint32
	for {
		b, err := r.ReadByte()
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
