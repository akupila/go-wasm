[![CircleCI](https://circleci.com/gh/akupila/go-wasm.svg?style=svg)](https://circleci.com/gh/akupila/go-wasm)
[![godoc](https://img.shields.io/badge/godoc-Reference-brightgreen.svg?style=flat)](https://godoc.org/github.com/akupila/go-wasm)

# go-wasm

A WebAssembly binary file parser in go.

The parser takes an `io.Reader` and parses a WebAssembly module from it, which
allows the user to see into the binary file. All data is read, future version
may allow to write it out too, which would allow modifying the binary.

For example:

```go
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
}
```

_when passed in a `.wasm` file compiled with go1.11:_

```
Index    Name        Size (bytes)
0        Custom      103
1        Type        58
2        Import      363
3        Function    1588
4        Table       5
5        Memory      5
6        Global      51
7        Export      14
8        Element     3066
9        Code        1174891
10       Data        1169054
11       Custom      45428
```

**Much** more information is available by type asserting on the items in
`.Sections`, for example:

```go
for i, s := range mod.Sections {
    switch section := s.(type) {
        case *wasm.SectionCode:
            // can now read function bytecode from section.
    }
}
```

## Installation

```
go get github.com/akupila/go-wasm/...
```

## Notes

This is a experimental, early and definitely not properly tested. There are
probably bugs. If you find one, please open an issue!
