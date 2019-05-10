package codegen

import (
	"fmt"
	"io"
	"io/ioutil"
)

// -----------------------------------------------------------------------------

const hextable = "0123456789abcdef"

// BytesFrom generates a byte slice variable from a byte slice.
//
func BytesFrom(w io.Writer, name string, b []byte) (err error) {

	_, err = fmt.Fprintf(w, "var %s = []byte{\n", name)
	if err != nil {
		return
	}
	buf := make([]byte, 8*6+1)
	for i := 0; i < len(b); i += 8 {
		buf[0] = '\t'
		base := 1
		max := 8
		if len(b)-i < max {
			max = len(b) - i
		}
		for j := 0; j < max; j++ {
			c := b[i+j]
			buf[base] = '0'
			buf[base+1] = 'x'
			buf[base+2] = hextable[c>>4]
			buf[base+3] = hextable[c&0xf]
			buf[base+4] = ','
			buf[base+5] = ' '
			base += 6
		}
		buf[base-1] = '\n'
		_, err = w.Write(buf[:base])
		if err != nil {
			return
		}
	}
	_, err = w.Write([]byte{'}', '\n'})
	return
}

// BytesFromFile generates a byte slice variable from a file.
//
func BytesFromFile(w io.Writer, name string, file string) (err error) {

	b, err := ioutil.ReadFile(file)
	if err != nil {
		return
	}
	return BytesFrom(w, name, b)
}

// -----------------------------------------------------------------------------
