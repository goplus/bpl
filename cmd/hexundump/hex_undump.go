package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/qiniu/bpl/hex"
)

// Usage: hexundump <hexdump-file> <binary-file>
//
func main() {

	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: hexundump <hexdump-file> <binary-file>\n\n")
		return
	}
	b, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	w := bytes.NewBuffer(nil)
	hex.UndumpText(w, string(b))
	err = ioutil.WriteFile(os.Args[2], w.Bytes(), 0666)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}
