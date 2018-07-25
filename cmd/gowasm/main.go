package main

import (
	"flag"
	"fmt"
	"os"

	wasm "github.com/akupila/go-wasm"
	"github.com/kr/pretty"
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

	_ = mod

	for _, sec := range mod.Sections {
		switch s := sec.(type) {
		case *wasm.SectionImport:
			pretty.Println(s)
		}
	}

	// pretty.Println(mod)

	// j, err := json.MarshalIndent(mod, "", "\t")
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(string(j))

	// for _, sec := range mod.Sections {
	// 	fmt.Printf("%T\n", sec)
	// 	switch s := sec.(type) {
	// 	case *wasm.SectionCode:
	// 		fmt.Println("bodies:", len(s.Bodies))
	// 	}
	// }
}
