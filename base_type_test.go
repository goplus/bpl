package bpl_test

import (
	"encoding/binary"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/qiniu/bpl"
	"qiniupkg.com/x/bufiox.v7"
)

func TestBaseType(t *testing.T) {

	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, 123)
	in := bufiox.NewReaderBuffer(b)

	ctx := bpl.NewContext()
	named := &bpl.Member{Name: "foo", Type: bpl.Int64}
	v, err := named.Match(in, ctx)
	if err != nil {
		t.Fatal("Member.Match failed:", err)
	}
	if v != int64(123) {
		t.Fatal("v != 123")
	}
	if v, ok := ctx.Var("foo"); !ok || v != int64(123) {
		t.Fatal("v != 123 - ", reflect.TypeOf(v), v, ok)
	}
}

func TestCString(t *testing.T) {

	b := []byte("Hello, world!")
	b = append(b, 0)
	in := bufiox.NewReaderBuffer(b)

	v, err := bpl.CString.Match(in, nil)
	if err != nil {
		t.Fatal("CString.Match failed:", err)
	}
	if v != "Hello, world!" {
		t.Fatal("CString.Match result:", v)
	}

	b2, err := json.Marshal(v)
	if err != nil {
		t.Fatal("json.Marshal failed:", err)
	}
	if string(b2) != `"Hello, world!"` {
		t.Fatal("json.Marshal result:", b2)
	}
}

func TestUintbe(t *testing.T) {

	b := []byte{1, 2, 3, 4}
	in := bufiox.NewReaderBuffer(b)
	v, err := bpl.Uintbe(3).Match(in, nil)
	if err != nil {
		t.Fatal("Uintbe.Match failed:", err)
	}
	if v != uint(0x010203) {
		t.Fatal("v != 0x010203:", v)
	}
}

func TestUintle(t *testing.T) {

	b := []byte{1, 2, 3, 4}
	in := bufiox.NewReaderBuffer(b)
	v, err := bpl.Uintle(3).Match(in, nil)
	if err != nil {
		t.Fatal("Uintbe.Match failed:", err)
	}
	if v != uint(0x030201) {
		t.Fatal("v != 0x030201:", v)
	}
}
