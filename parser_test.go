package wasm

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParser(t *testing.T) {
	tt := []struct {
		file     string
		sections []string
	}{
		{"empty.wasm", nil},
		{"helloworld.wasm", []string{"Custom", "Type", "Import", "Function", "Table", "Memory", "Global", "Export", "Element", "Code", "Data", "Name"}},
	}

	for _, tc := range tt {
		t.Run(tc.file, func(t *testing.T) {
			f, done := open(t, tc.file)
			defer done()

			actual, err := Parse(f)
			if err != nil {
				t.Fatal(err)
			}

			if len(actual.Sections) != len(tc.sections) {
				t.Fatalf("Number of sections does not match; expected %d, actual %d", len(tc.sections), len(actual.Sections))
			}

			for i, sec := range actual.Sections {
				n := strings.TrimPrefix(fmt.Sprintf("%T", sec), "*wasm.Section")
				if tc.sections[i] != n {
					t.Errorf("Section %d/%d type doesn not match; expected %q, actual %q", i+1, len(tc.sections), tc.sections[i], n)
				}
			}

			// TODO(akupila): add more assertions
		})
	}
}

var filename = "testdata/helloworld.wasm"

func Example_parseFile() {
	f, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	mod, err := Parse(f)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("The file has %d sections\n", len(mod.Sections))
	// Output:
	// The file has 12 sections
}

func BenchmarkParser(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		f, done := open(b, "helloworld.wasm")
		defer done()

		b.StartTimer()
		_, err := Parse(f)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func open(t testing.TB, name string) (io.Reader, func()) {
	f, err := os.Open(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return f, func() {
		if err := f.Close(); err != nil {
			t.Fatal(err)
		}
	}
}
