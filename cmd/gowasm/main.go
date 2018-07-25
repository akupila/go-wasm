package main

import (
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	wasm "github.com/akupila/go-wasm"
)

func main() {
	file := flag.String("file", "", "file to parse (.wasm)")
	flag.Parse()

	if *file == "" {
		flag.Usage()
		os.Exit(2)
	}

	f, err := os.Open(*file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open file: %v", err)
		os.Exit(1)
	}
	defer f.Close()

	mod, err := wasm.Parse(f)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)
	fmt.Fprintf(w, "Index\tName\tSize (bytes)\n")
	for i, s := range mod.Sections {
		fmt.Fprintf(w, "%d\t%s\t%d\n", i, s.Name(), s.Size())
	}
	w.Flush()

	// Much more information is available by type asserting the section:
	// switch section := s.(type) {
	//     case *wasm.SectionCode:
	//         // can now read function bytecode from section.
	// }
}
