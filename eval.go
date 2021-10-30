package wasm

import (
	"bytes"
	"fmt"
	"io"
)

// Eval tries to evaluate the (constant) expression in the buffer. It returns
// the evaluation stack at the end of the sequence or an error, if there was
// one.
func Eval(r *bytes.Buffer) ([]interface{}, error) {
	// Right now, this is only meant to parse very simple expressions like the
	// one used for the start offset of a data segment (which is an expression).
	var stack []interface{}
	for {
		b, err := r.ReadByte()
		if err == io.EOF {
			return nil, io.ErrUnexpectedEOF
		}
		if b == 0x0B { // end
			// End of expression.
			break
		}
		switch b {
		case 0x41: // i32.const
			var n int32
			err := readVarInt32(r, &n)
			if err != nil {
				return nil, err
			}
			stack = append(stack, n)
		default:
			return nil, fmt.Errorf("unknown opcode: 0x%02X", b)
		}
	}
	return stack, nil
}
