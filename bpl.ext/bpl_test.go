package bpl

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/goplus/bpl/binary"
	qlang "github.com/xushiwei/qlang/spec"
)

// -----------------------------------------------------------------------------

func TestHexdump(t *testing.T) {

	b := bytes.NewBuffer(nil)
	d := hex.Dumper((*filterWriter)(b))
	d.Write([]byte{
		0x60, 0x60, 0x60, 0x60, 0x60, 0x60, 0x60, 0x60, 0x60, 0x60, 0x60, 0x60, 0x60, 0x60, 0x60, 0x60,
		0x60, 0x60, 0x60, 0x60, 0x60, 0x60, 0x60, 0x60, 0x60, 0x60, 0x60, 0x60, 0x60,
	})
	d.Close()
	if string(b.Bytes()) != `00000000  60 60 60 60 60 60 60 60  60 60 60 60 60 60 60 60  |................|
00000010  60 60 60 60 60 60 60 60  60 60 60 60 60           |.............|
` {
		t.Fatal("Hexdump failed")
	}
}

// -----------------------------------------------------------------------------

const codeBasic = `

sub1 = int8 uint16

subType = cstring

doc = [sub1 uint32 float32 cstring subType float64]
`

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

func TestBasic(t *testing.T) {

	foo := &fooType{
		A: 1, B: 2, C: 3, D: 3.14, E: "Hello", F: subType{Foo: "foo"}, G: 7.52,
		// 1 + 2 + 4 + 4 + 6 + 4 + 8 = 29
	}
	b, err := binary.Marshal(&foo)
	if err != nil {
		t.Fatal("binary.Marshal failed:", err)
	}
	if len(b) != 29 {
		t.Fatal("len(b) != 29, len:", len(b), "data:", string(b))
	}

	r, err := NewFromString(codeBasic, "")
	if err != nil {
		t.Fatal("New failed:", err)
	}
	v, err := r.MatchBuffer(b)
	if err != nil {
		t.Fatal("Match failed:", err, "len:", len(b))
	}
	ret, err := json.Marshal(v)
	if err != nil {
		t.Fatal("json.Marshal failed:", err)
	}
	if string(ret) != `[null,3,3.14,"Hello","foo",7.52]` {
		t.Fatal("ret:", string(ret))
	}
}

// -----------------------------------------------------------------------------

const codeBasic2 = `

sub1 = [int8 uint16]

subType = cstring

doc = [sub1 uint32] float32 cstring [subType] float64
`

func TestBasic2(t *testing.T) {

	foo := &fooType{
		A: 1, B: 2, C: 3, D: 3.14, E: "Hello", F: subType{Foo: "foo"}, G: 7.52,
		// 1 + 2 + 4 + 4 + 6 + 4 + 8 = 29
	}
	b, err := binary.Marshal(&foo)
	if err != nil {
		t.Fatal("binary.Marshal failed:", err)
	}
	if len(b) != 29 {
		t.Fatal("len(b) != 29, len:", len(b), "data:", string(b))
	}

	r, err := NewFromString(codeBasic2, "")
	if err != nil {
		t.Fatal("New failed:", err)
	}
	v, err := r.MatchBuffer(b)
	if err != nil {
		t.Fatal("Match failed:", err, "len:", len(b))
	}
	ret, err := json.Marshal(v)
	if err != nil {
		t.Fatal("json.Marshal failed:", err)
	}
	if string(ret) != `[[1,2],3,"foo"]` {
		t.Fatal("ret:", string(ret))
	}
}

// -----------------------------------------------------------------------------

const codeStruct = `

sub1 = {/C
	int8   a
	uint16 b
}

subType = {
	f cstring
	assert f == "foo"
}

doc = {
	sub1 sub1
	c    uint32
	d    float32
	e    [5]char
	_    byte
	f    subType
	_    float64
	assert e == "Hello"
}
`

func TestStruct(t *testing.T) {

	foo := &fooType{
		A: 1, B: 2, C: 3, D: 3.14, E: "Hello", F: subType{Foo: "foo"}, G: 7.52,
		// 1 + 2 + 4 + 4 + 6 + 4 + 8 = 29
	}
	b, err := binary.Marshal(&foo)
	if err != nil {
		t.Fatal("binary.Marshal failed:", err)
	}
	if len(b) != 29 {
		t.Fatal("len(b) != 29, len:", len(b), "data:", string(b))
	}

	r, err := NewFromString(codeStruct, "")
	if err != nil {
		t.Fatal("New failed:", err)
	}
	v, err := r.MatchBuffer(b)
	if err != nil {
		t.Fatal("Match failed:", err, "len:", len(b))
	}
	ret, err := json.Marshal(v)
	if err != nil {
		t.Fatal("json.Marshal failed:", err)
	}
	if string(ret) != `{"c":3,"d":3.14,"e":"Hello","f":{"f":"foo"},"sub1":{"a":1,"b":2}}` {
		t.Fatal("ret:", string(ret))
	}
}

// -----------------------------------------------------------------------------

const codeDump = `

doc = {
	let a = 1
	let _b = 3
}
`

func TestDump(t *testing.T) {

	r, err := NewFromString(codeDump, "")
	if err != nil {
		t.Fatal("New failed:", err)
	}
	v, err := r.MatchBuffer(nil)
	if err != nil {
		t.Fatal("Match failed:", err)
	}
	ret, err := json.Marshal(v)
	if err != nil {
		t.Fatal("json.Marshal failed:", err)
	}
	if string(ret) != `{"_b":3,"a":1}` {
		t.Fatal("ret:", string(ret))
	}
	var b bytes.Buffer
	DumpDom(&b, v, 0)
	if b.String() != "{\n  a: 1\n}" {
		t.Fatal("dump:", b.String())
	}
}

// -----------------------------------------------------------------------------

const codeDump2 = `

doc = {
	return undefined
}
`

func TestDump2(t *testing.T) {

	r, err := NewFromString(codeDump2, "")
	if err != nil {
		t.Fatal("New failed:", err)
	}
	v, err := r.MatchBuffer(nil)
	if err != nil {
		t.Fatal("Match failed:", err)
	}
	if v != qlang.Undefined {
		t.Fatal("v:", v)
	}
	var b bytes.Buffer
	DumpDom(&b, v, 0)
	if b.String() != "\"```undefined```\"" {
		t.Fatal("dump:", b.String())
	}
}

// -----------------------------------------------------------------------------

const codeRtmp1 = `

AMF0_NULL = {
	return nil
}

AMF0_NUMBER = {
    val float64be
	return val
}

AMF0_STRING = {
    len uint16be
    val [len]char
	return val
}

AMF0_TYPE = {
    marker byte
    case marker {
        0x00: AMF0_NUMBER
        0x02: AMF0_STRING
		0x05: AMF0_NULL
    }
}

AMF0_CMDDATA = {
    cmd           AMF0_TYPE
    transactionId AMF0_TYPE
    value         *AMF0_TYPE
}

doc = {
    msg AMF0_CMDDATA
}
`

func TestRtmp1(t *testing.T) {

	buf := []byte{
		0x02, 0x00, 0x08, 0x6f, 0x6e, 0x42, 0x57, 0x44,
		0x6f, 0x6e, 0x65, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x05,
	}
	r, err := NewFromString(codeRtmp1, "")
	if err != nil {
		t.Fatal("New failed:", err)
	}
	v, err := r.MatchBuffer(buf)
	if err != nil {
		t.Fatal("Match failed:", err)
	}
	ret, err := json.Marshal(v)
	if err != nil {
		t.Fatal("json.Marshal failed:", err)
	}
	if string(ret) != `{"msg":{"cmd":"onBWDone","transactionId":0,"value":[null]}}` {
		t.Fatal("ret:", string(ret))
	}
}

// -----------------------------------------------------------------------------
