package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/goplus/bpl/replay"
)

var (
	host = flag.String("s", "", "remote address to dial.")
)

// qreplay -s <host:port> <replay.log>
//
func main() {

	flag.Parse()

	if *host == "" {
		fmt.Fprintln(os.Stderr, "Usage: qreplay -s <host:port> <replay.log>")
		flag.PrintDefaults()
		return
	}

	var in *os.File
	args := flag.Args()
	if len(args) > 0 {
		file := args[0]
		f, err := os.Open(file)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Open failed:", err)
			os.Exit(1)
		}
		defer f.Close()
		in = f
	} else {
		in = os.Stdin
	}

	err := replay.HexRequest(*host, nil, in, "[REQ]")
	if err != nil {
		fmt.Fprintln(os.Stderr, "replay.HexRequest:", err)
		os.Exit(2)
	}
}
