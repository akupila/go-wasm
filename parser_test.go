package wasm

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var update = flag.Bool("update", false, "Update golden files")

func TestParser(t *testing.T) {
	tt := []struct {
		file        string
		numSections int
	}{
		{"empty.wasm", 0},
		{"helloworld.wasm", 12},
	}

	for _, tc := range tt {
		t.Run(tc.file, func(t *testing.T) {
			f, done := open(t, tc.file)
			defer done()

			actual, err := Parse(f)
			if err != nil {
				t.Fatal(err)
			}

			if len(actual.Sections) != tc.numSections {
				t.Fatalf("Number of sections does not match; expected %d, actual %d", tc.numSections, len(actual.Sections))
			}

			for i, s := range actual.Sections {
				name := strings.TrimSuffix(tc.file, filepath.Ext(tc.file))
				name = fmt.Sprintf("golden/%s-%02d.json", name, i)

				j, err := json.MarshalIndent(s, "", "\t")
				if err != nil {
					t.Fatal(err)
				}

				assertGolden(t, j, name)
			}
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

func assertGolden(t *testing.T, b []byte, name string) {
	t.Helper()

	tf := filepath.Join("testdata", name)
	if *update {
		if err := ioutil.WriteFile(tf, b, 0644); err != nil {
			t.Fatal(err)
		}
		return
	}

	g, err := ioutil.ReadFile(tf)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(g, b) {
		addr := 0
		for i, v := range g {
			if b[i] != v {
				addr = i
				break
			}
		}
		t.Errorf("Golden file %s does not match; difference at address 0x%06x", tf, addr)
	}
}
