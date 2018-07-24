package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

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

	parser := &wasm.Parser{}
	m, err := parser.Parse(f)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	_ = m

	j, err := json.MarshalIndent(m, "", "\t")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(j))
}
