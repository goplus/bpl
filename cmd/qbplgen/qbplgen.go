package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/goplus/bpl/go/codegen"
)

// qbplgen <protocol>.bpl qbpl|qbplproxy
//
func main() {

	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "Usage: qbplgen <protocol>.bpl qbpl|qbplproxy")
		return
	}

	protocol := os.Args[1]
	if path.Ext(protocol) == "" {
		baseDir := os.Getenv("HOME") + "/.qbpl/formats/"
		protocol = baseDir + protocol + ".bpl"
	}

	qbpl := os.Args[2]
	qdestname := strings.TrimPrefix(qbpl, "qbpl")
	srcbase := os.Getenv("QBOXROOT") + "/bpl/src/qiniupkg.com/text/"
	qbpl = srcbase + qbpl + "/" + qbpl + ".go"
	src, err := ioutil.ReadFile(qbpl)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	src = bytes.Replace(src, []byte("flag.Parse()\n"), []byte("flag.Parse()\n\t*protocol = \"_.bpl\"\n"), 1)
	src = bytes.Replace(src, []byte("bpl.NewFromFile(*protocol)"), []byte("bpl.New(BPL_PROTOCOL, \"\")"), 1)
	src = bytes.Replace(src, []byte("-p <protocol>.bpl "), []byte{}, -1)

	qdest := srcbase + "/q" + strings.TrimSuffix(path.Base(protocol), ".bpl") + qdestname
	err = os.MkdirAll(qdest, 0700)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	f, err := os.Create(qdest + "/protocol.go")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(3)
	}
	defer f.Close()
	f.WriteString("package main\n\n")
	err = codegen.BytesFromFile(f, "BPL_PROTOCOL", protocol)
	if err != nil {
		fmt.Fprintln(os.Stderr, "codegen.BytesFromFile:", err)
		os.Exit(3)
	}

	err = ioutil.WriteFile(qdest+"/main.go", src, 0777)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(4)
	}
}
