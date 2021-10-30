package wasm

import (
	"bytes"
	"reflect"
	"testing"
)

func TestEval(t *testing.T) {
	tests := []struct {
		buf    []byte
		result []interface{}
	}{
		{[]byte{0x41, 0x80, 0x80, 0x04, 0x0B}, []interface{}{int32(0x10000)}},
		{[]byte{0x41, 0xA0, 0xFE, 0x04, 0x0B}, []interface{}{int32(0x13f20)}},
		{[]byte{0x0B}, nil},
	}

	for i, tc := range tests {
		r := bytes.NewBuffer(tc.buf)
		result, err := Eval(r)
		if err != nil {
			t.Errorf("failed to run test %d: %w", i, err)
			continue
		}
		if !reflect.DeepEqual(result, tc.result) {
			t.Errorf("expected %v but got %v", tc.result, result)
		}
	}
}
