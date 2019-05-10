package bpl_test

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/qiniu/bpl"
	bin "github.com/qiniu/bpl/binary"
	"qiniupkg.com/x/bufiox.v7"
)

type fixedType struct {
	A int8
	B uint16
	C uint32
	D float32
}

func TestFixedStruct(t *testing.T) {
	v := fixedType{
		A: 1,
		B: 2,
		C: 3,
		D: 3.14,
	}

	b := new(bytes.Buffer)
	err := binary.Write(b, binary.LittleEndian, &v)
	if err != nil {
		t.Fatal("binary.Write failed:", err)
	}
	if b.Len() != 11 {
		t.Fatal("len != 11")
	}

	members := []bpl.Ruler{
		&bpl.Member{Name: "a", Type: bpl.Int8},
		&bpl.Member{Name: "b", Type: bpl.Uint16},
		&bpl.Member{Name: "c", Type: bpl.Uint32},
		&bpl.Member{Name: "d", Type: bpl.Float32},
	}
	struc := bpl.Struct(members)
	if struc.SizeOf() != 11 {
		t.Fatal("struct.size != 11 - ", struc.SizeOf())
	}

	in := bufiox.NewReaderBuffer(b.Bytes())
	if in.Buffered() != 11 {
		t.Fatal("len != 11")
	}

	ctx := bpl.NewContext()
	ret, err := struc.Match(in, ctx)
	if err != nil {
		t.Fatal("struc.Match failed:", err)
	}
	text, err := json.Marshal(ret)
	if err != nil {
		t.Fatal("json.Marshal failed:", err)
	}
	if string(text) != `{"a":1,"b":2,"c":3,"d":3.14}` {
		t.Fatal("json.Marshal result:", string(text))
	}
}

type subType struct {
	Foo string
}

type fooType struct {
	A int8
	B uint16
	C uint32
	D float32
	E string
	F subType
	G float64
}

var (
	rSubType  = bpl.Seq(bpl.CString)
	rFooType  = bpl.Seq(bpl.Int8, bpl.Uint16, bpl.Uint32, bpl.Float32, bpl.CString, rSubType, bpl.Float64)
	rFooType2 = bpl.And(bpl.Seq(bpl.Int8, bpl.Uint16), bpl.Uint32, bpl.Float32, bpl.CString, bpl.Seq(rSubType), bpl.Float64)
)

func TestStruct(t *testing.T) {

	foo := &fooType{
		A: 1, B: 2, C: 3, D: 3.14, E: "Hello", F: subType{Foo: "foo"}, G: 7.52,
		// 1 + 2 + 4 + 4 + 6 + 4 + 8 = 29
	}
	b, err := bin.Marshal(&foo)
	if err != nil {
		t.Fatal("binary.Marshal failed:", err)
	}
	if len(b) != 29 {
		t.Fatal("len(b) != 29, len:", len(b), "data:", string(b))
	}

	fooTyp := reflect.TypeOf(foo)
	r, err := bpl.TypeFrom(fooTyp)
	if err != nil {
		t.Fatal("bpl.TypeFrom failed:", err)
	}
	in := bufiox.NewReaderBuffer(b)
	ctx := bpl.NewContext()
	v, err := r.Match(in, ctx)
	if err != nil {
		t.Fatal("Match failed:", err, "len:", len(b))
	}
	ret, err := json.Marshal(v)
	if err != nil {
		t.Fatal("json.Marshal failed:", err)
	}
	if string(ret) != `{"a":1,"b":2,"c":3,"d":3.14,"e":"Hello","f":{"foo":"foo"},"g":7.52}` {
		t.Fatal("ret:", string(ret))
	}
}

func TestSeq(t *testing.T) {

	foo := &fooType{
		A: 1, B: 2, C: 3, D: 3.14, E: "Hello", F: subType{Foo: "foo"}, G: 7.52,
		// 1 + 2 + 4 + 4 + 6 + 4 + 8 = 29
	}
	b, err := bin.Marshal(&foo)
	if err != nil {
		t.Fatal("binary.Marshal failed:", err)
	}
	in := bufiox.NewReaderBuffer(b)
	v, err := rFooType.Match(in, bpl.NewContext())
	if err != nil {
		t.Fatal("Match failed:", err, "len:", len(b))
	}
	ret, err := json.Marshal(v)
	if err != nil {
		t.Fatal("json.Marshal failed:", err)
	}
	if string(ret) != `[1,2,3,3.14,"Hello",["foo"],7.52]` {
		t.Fatal("ret:", string(ret))
	}
}

func TestSeq2(t *testing.T) {

	foo := &fooType{
		A: 1, B: 2, C: 3, D: 3.14, E: "Hello", F: subType{Foo: "foo"}, G: 7.52,
		// 1 + 2 + 4 + 4 + 6 + 4 + 8 = 29
	}
	b, err := bin.Marshal(&foo)
	if err != nil {
		t.Fatal("binary.Marshal failed:", err)
	}
	in := bufiox.NewReaderBuffer(b)
	ctx := bpl.NewContext()
	_, err = rFooType2.Match(in, ctx)
	if err != nil {
		t.Fatal("Match failed:", err, "len:", len(b))
	}
	v := ctx.Dom()
	ret, err := json.Marshal(v)
	if err != nil {
		t.Fatal("json.Marshal failed:", err)
	}
	if string(ret) != `[1,2,["foo"]]` {
		t.Fatal("ret:", string(ret))
	}
}
