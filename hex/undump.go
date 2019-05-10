package hex

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"io"
	"os"
	"strconv"
	"strings"
)

func undump1(w io.Writer, text string) { // bd c2 c1 24 93 55 2a 4d

	b, err := hex.DecodeString(strings.Replace(text, " ", "", -1))
	if err != nil {
		panic(err)
	}
	w.Write(b)
}

func undumpLine(w io.Writer, line string) {

	parts := strings.SplitN(line, "  ", 4)
	max := 3
	if len(parts) < max {
		max = len(parts)
	}
	addr := parts[0]
	if len(addr) < 8 {
		return
	}
	_, err := strconv.ParseInt(addr, 16, 64)
	if err != nil {
		return
	}
	for i := 1; i < max; i++ {
		undump1(w, parts[i])
	}
}

// Undump reverts `hexdump -C binary` result back to a binary data.
//
func Undump(w io.Writer, in *bufio.Reader, filter string) { // filter = [REQ]

	fskip := true
	for {
		line, err := in.ReadString('\n')
		if filter != "" {
			if strings.HasPrefix(line, "[INFO]") {
				fskip = !strings.HasPrefix(line[6:], filter)
				continue
			}
			if fskip {
				continue
			}
		}
		undumpLine(w, line)
		if err != nil {
			if err != io.EOF {
				panic(err)
			}
			return
		}
	}
}

// UndumpText reverts `hexdump -C binary` result back to a binary data.
//
func UndumpText(w io.Writer, text string) {

	lines := strings.Split(text, "\n")
	for _, line := range lines {
		undumpLine(w, line)
	}
}

// TextReader returns a reader that reverts `hexdump -C binary` result back to a binary data.
//
func TextReader(text string) *bytes.Reader {

	var w bytes.Buffer
	UndumpText(&w, text)
	return bytes.NewReader(w.Bytes())
}

// Reader returns a reader that reverts `hexdump -C binary` result back to a binary data.
//
func Reader(f io.Reader, filter string) (ret io.ReadCloser, err error) {

	pr, pw := io.Pipe()
	go func() {
		in := bufio.NewReader(f)
		Undump(pw, in, filter)
		pw.Close()
	}()

	return pr, nil
}

// Open returns a reader that reverts `hexdump -C binary` result back to a binary data.
//
func Open(fname string, filter string) (ret io.ReadCloser, err error) {

	f, err := os.Open(fname)
	if err != nil {
		return
	}

	pr, pw := io.Pipe()
	go func() {
		in := bufio.NewReader(f)
		Undump(pw, in, filter)
		pw.Close()
		f.Close()
	}()

	return pr, nil
}
