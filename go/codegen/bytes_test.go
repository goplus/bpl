package codegen

import (
	"bytes"
	"testing"
)

func doTestBytesFrom(t *testing.T, name string, b []byte, expected string) {

	w := bytes.NewBuffer(nil)
	err := BytesFrom(w, name, b)
	if err != nil {
		t.Fatal("BytesFrom failed:", err)
	}
	ret := w.String()
	if ret != expected {
		t.Fatal("bytes:", ret)
	}
}

const expected0 = `var foo = []byte{
}
`

const expected1 = `var foo = []byte{
	0x02, 0x03, 0x03, 0x09, 0x00,
}
`

const expected2 = `var foo = []byte{
	0x02, 0x03, 0x03, 0x09, 0x00, 0x02, 0xe8, 0x60,
	0xf8,
}
`

func TestBytesFrom(t *testing.T) {

	doTestBytesFrom(t, "foo", nil, expected0)
	doTestBytesFrom(t, "foo", []byte{2, 3, 3, 9, 0}, expected1)
	doTestBytesFrom(t, "foo", []byte{2, 3, 3, 9, 0, 2, 0xe8, 96, 0xf8}, expected2)
}
