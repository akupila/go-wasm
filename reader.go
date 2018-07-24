package wasm

import (
	"io"
)

// reader wraps io.Reader and keeps track of the current position in the input.
type reader struct {
	rd io.Reader // reader provided by the client
	i  int       // current index
}

func newReader(r io.Reader) *reader {
	return &reader{r, 0}
}

// Index returns the current position in the file.
func (r *reader) Index() int {
	return r.i
}

// Read reads bytes into p and returns the number of bytes read.
// The number of bytes read are recorded in the reader.
func (r *reader) Read(p []byte) (int, error) {
	n, err := r.rd.Read(p)
	if err != nil {
		return 0, err
	}
	r.i += n
	return n, nil
}
