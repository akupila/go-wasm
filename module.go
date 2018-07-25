package wasm

// A Module represents a parsed WASM module.
type Module struct {
	// Sections contains the sections in the parsed file, in the order they
	// appear in the file. A valid  but empty file will have zero sections.
	//
	// The items in the slice will be a mix of the SectionXXX types.
	Sections []Section
}
