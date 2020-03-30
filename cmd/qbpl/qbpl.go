package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	bpl "github.com/qiniu/bpl/bpl.ext"
	"github.com/qiniu/x/log"
)

var (
	protocol = flag.String("p", "", "protocol file in BPL syntax. default is guessed by extension.")
	output   = flag.String("o", "", "output log file, default is stderr.")
	logmode  = flag.String("l", "", "log mode: short (default) or long.")
)

// qbpl [-p <protocol>.bpl -o <output>.log -l <logmode>] <file>
//
func main() {

	flag.Parse()
	bpl.SetDumpCode(os.Getenv("BPL_DUMPCODE"))

	var in *bufio.Reader
	args := flag.Args()
	if len(args) > 0 {
		file := args[0]
		f, err := os.Open(file)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Open failed:", file)
		}
		defer f.Close()
		in = bufio.NewReader(f)
	} else {
		in = bufio.NewReader(os.Stdin)
	}

	if *protocol == "" {
		if len(args) == 0 {
			fmt.Fprintln(os.Stderr, "Usage: qbpl [-p <protocol>.bpl -o <output>.log -l <logmode>] <file>")
			flag.PrintDefaults()
			return
		}
		ext := filepath.Ext(args[0])
		if ext != "" {
			*protocol = os.Getenv("HOME") + "/.qbpl/formats/" + ext[1:] + ".bpl"
		}
	}

	logflags := bpl.Ldefault
	flong := (*logmode == "long")
	if flong {
		logflags = bpl.Llong
	}

	if *output != "" {
		f, err := os.Create(*output)
		if err != nil {
			log.Fatalln("Create log file failed:", err)
		}
		defer f.Close()
		bpl.SetDumper(f, logflags)
	}
	log.Std = bpl.Dumper

	ruler, err := bpl.NewFromFile(*protocol)
	if err != nil {
		log.Fatalln("bpl.NewFromFile failed:", err)
	}

	ctx := bpl.NewContext()
	_, err = ruler.SafeMatch(in, ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Match failed:", err)
		return
	}
}
