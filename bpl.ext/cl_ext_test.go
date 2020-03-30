package bpl

import (
	"encoding/json"
	"testing"

	"github.com/qiniu/bpl/binary"
	"github.com/qiniu/x/bufiox"
)

// -----------------------------------------------------------------------------

const codeArray = `

sub1 = int8 uint16

subType = {
    array [2]cstring
}

doc = [sub1 uint32 float32 cstring subType float64]
`

type subType2 struct {
	Foo string
	Bar string
}

type fooType2 struct {
	A int8
	B uint16
	C uint32
	D float32
	E string
	F subType2
	G float64
}

func TestArray(t *testing.T) {

	foo := &fooType2{
		A: 1, B: 2, C: 3, D: 3.14, E: "Hello", F: subType2{Foo: "foo", Bar: "bar"}, G: 7.52,
	}
	b, err := binary.Marshal(&foo)
	if err != nil {
		t.Fatal("binary.Marshal failed:", err)
	}
	if len(b) != 33 {
		t.Fatal("len(b) != 33, len:", len(b), "data:", string(b))
	}

	r, err := NewFromString(codeArray, "")
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
	if string(ret) != `[null,3,3.14,"Hello",{"array":["foo","bar"]},7.52]` {
		t.Fatal("ret:", string(ret))
	}
}

// -----------------------------------------------------------------------------

const codeArray2 = `

headerType = {
    type int32
    _    int32
    n    int32
    m    int32
}

recType = {
    h     headerType
    array [h.n + h.m]cstring
}

doc = [int32] *[recType]
`

type headerType struct {
	Type int32
	Len  int32
	N    int32
	M    int32
}

type recType1 struct {
	H  headerType
	A1 string
	A2 string
	A3 string
}

type recType2 struct {
	H  headerType
	A1 string
	A2 string
}

type fooType3 struct {
	N  int32
	R1 recType1
	R2 recType2
}

func TestArray2(t *testing.T) {

	foo := &fooType3{
		N: 2,
		R1: recType1{
			H: headerType{
				Type: 1,
				N:    1,
				M:    2,
			},
			A1: "hello",
			A2: "world",
			A3: "bpl",
		},
		R2: recType2{
			H: headerType{
				Type: 2,
				N:    1,
				M:    1,
			},
			A1: "foo",
			A2: "bar",
		},
	}
	b, err := binary.Marshal(&foo)
	if err != nil {
		t.Fatal("binary.Marshal failed:", err)
	}
	if len(b) != 60 {
		t.Fatal("len(b) != 60, len:", len(b), "data:", string(b))
	}

	r, err := NewFromString(codeArray2, "")
	if err != nil {
		t.Fatal("New failed:", err)
	}
	v, err := r.MatchBuffer(b)
	if err != nil {
		t.Fatal("Match failed:", err, "v:", v)
	}
	ret, err := json.Marshal(v)
	if err != nil {
		t.Fatal("json.Marshal failed:", err)
	}
	if string(ret) != `[2,{"array":["hello","world","bpl"],"h":{"m":2,"n":1,"type":1}},{"array":["foo","bar"],"h":{"m":1,"n":1,"type":2}}]` {
		t.Fatal("ret:", string(ret))
	}
}

// -----------------------------------------------------------------------------

const codeCase = `

headerType = {
	type int32
	_    int32
	n    int32
	m    int32
}

recType = {
	h headerType
	case h.type {
		1: {t1 [3]cstring}
		2: {t2 [2]cstring}
	}
}

doc = [int32] *[recType]
`

func TestCase(t *testing.T) {

	SetCaseType = false
	foo := &fooType3{
		N: 2,
		R1: recType1{
			H: headerType{
				Type: 1,
				N:    1,
				M:    2,
			},
			A1: "hello",
			A2: "world",
			A3: "bpl",
		},
		R2: recType2{
			H: headerType{
				Type: 2,
				N:    1,
				M:    1,
			},
			A1: "foo",
			A2: "bar",
		},
	}
	b, err := binary.Marshal(&foo)
	if err != nil {
		t.Fatal("binary.Marshal failed:", err)
	}
	if len(b) != 60 {
		t.Fatal("len(b) != 60, len:", len(b), "data:", string(b))
	}

	r, err := NewFromString(codeCase, "")
	if err != nil {
		t.Fatal("New failed:", err)
	}
	v, err := r.MatchBuffer(b)
	if err != nil {
		t.Fatal("Match failed:", err)
	}
	ret, err := json.Marshal(v)
	if err != nil {
		t.Fatal("json.Marshal failed:", err)
	}
	if string(ret) != `[2,{"h":{"m":2,"n":1,"type":1},"t1":["hello","world","bpl"]},{"h":{"m":1,"n":1,"type":2},"t2":["foo","bar"]}]` {
		t.Fatal("ret:", string(ret))
	}
}

// -----------------------------------------------------------------------------

const codeCase2 = `

const (
	NVAL = 1
)

headerType = {
	type int32
	do println("type:", type)

	_ int32
	n int32
	m int32
	assert n == NVAL
}

recType = {h headerType} case h.type {
	1: {t1 [3]cstring}
	2: {t2 [2]cstring}
}

doc = [int32] *[recType]
`

func TestCase2(t *testing.T) {

	SetCaseType = false
	foo := &fooType3{
		N: 2,
		R1: recType1{
			H: headerType{
				Type: 1,
				N:    1,
				M:    2,
			},
			A1: "hello",
			A2: "world",
			A3: "bpl",
		},
		R2: recType2{
			H: headerType{
				Type: 2,
				N:    1,
				M:    1,
			},
			A1: "foo",
			A2: "bar",
		},
	}
	b, err := binary.Marshal(&foo)
	if err != nil {
		t.Fatal("binary.Marshal failed:", err)
	}
	if len(b) != 60 {
		t.Fatal("len(b) != 60, len:", len(b), "data:", string(b))
	}

	r, err := NewFromString(codeCase2, "")
	if err != nil {
		t.Fatal("New failed:", err)
	}
	v, err := r.MatchBuffer(b)
	if err != nil {
		t.Fatal("Match failed:", err)
	}
	ret, err := json.Marshal(v)
	if err != nil {
		t.Fatal("json.Marshal failed:", err)
	}
	if string(ret) != `[2,{"h":{"m":2,"n":1,"type":1},"t1":["hello","world","bpl"]},{"h":{"m":1,"n":1,"type":2},"t2":["foo","bar"]}]` {
		t.Fatal("ret:", string(ret))
	}
}

// -----------------------------------------------------------------------------

const codeRead = `

headerType = {
	type int32
	len  int32
	_    int32
	_    int32
}

recType = {
	h headerType
	read h.len - sizeof(headerType) do case h.type {
		1: {t1 [2]cstring}
		2: {t2 [1]cstring}
	}
}

doc = [int32] *[recType]
`

func TestRead(t *testing.T) {

	SetCaseType = false
	foo := &fooType3{
		N: 2,
		R1: recType1{
			H: headerType{
				Type: 1,
				Len:  16 + 16,
				N:    1,
				M:    2,
			},
			A1: "hello", // 6
			A2: "world", // 6
			A3: "bpl",   // 4
		},
		R2: recType2{
			H: headerType{
				Type: 2,
				Len:  16 + 8,
				N:    1,
				M:    1,
			},
			A1: "foo", // 4
			A2: "bar", // 4
		},
	}
	b, err := binary.Marshal(&foo)
	if err != nil {
		t.Fatal("binary.Marshal failed:", err)
	}
	if len(b) != 60 {
		t.Fatal("len(b) != 60, len:", len(b), "data:", string(b))
	}

	r, err := NewFromString(codeRead, "")
	if err != nil {
		t.Fatal("New failed:", err)
	}
	v, err := r.MatchBuffer(b)
	if err != nil {
		t.Fatal("Match failed:", err)
	}
	ret, err := json.Marshal(v)
	if err != nil {
		t.Fatal("json.Marshal failed:", err)
	}
	if string(ret) != `[2,{"h":{"len":32,"type":1},"t1":["hello","world"]},{"h":{"len":24,"type":2},"t2":["foo"]}]` {
		t.Fatal("ret:", string(ret))
	}
}

// -----------------------------------------------------------------------------

const codeIf = `

headerType = {
	type int32
	len  int32
	_    int32
	_    int32
}

recType = {
	h headerType
	read h.len - sizeof(headerType) do {
		if h.type == 1 {
			t1 [2]cstring
		} elif h.type == 2 {
			t2 [1]cstring
		}
	}
}

doc = [int32] *[recType]
`

func TestIf(t *testing.T) {

	foo := &fooType3{
		N: 2,
		R1: recType1{
			H: headerType{
				Type: 1,
				Len:  16 + 16,
				N:    1,
				M:    2,
			},
			A1: "hello", // 6
			A2: "world", // 6
			A3: "bpl",   // 4
		},
		R2: recType2{
			H: headerType{
				Type: 2,
				Len:  16 + 8,
				N:    1,
				M:    1,
			},
			A1: "foo", // 4
			A2: "bar", // 4
		},
	}
	b, err := binary.Marshal(&foo)
	if err != nil {
		t.Fatal("binary.Marshal failed:", err)
	}
	if len(b) != 60 {
		t.Fatal("len(b) != 60, len:", len(b), "data:", string(b))
	}

	r, err := NewFromString(codeIf, "")
	if err != nil {
		t.Fatal("New failed:", err)
	}
	v, err := r.MatchBuffer(b)
	if err != nil {
		t.Fatal("Match failed:", err)
	}
	ret, err := json.Marshal(v)
	if err != nil {
		t.Fatal("json.Marshal failed:", err)
	}
	if string(ret) != `[2,{"h":{"len":32,"type":1},"t1":["hello","world"]},{"h":{"len":24,"type":2},"t2":["foo"]}]` {
		t.Fatal("ret:", string(ret))
	}
}

// -----------------------------------------------------------------------------

const codeLet = `

doc = {
	let a = 1
	let a, b = [2, 3]
}
`

func TestLet(t *testing.T) {

	r, err := NewFromString(codeLet, "")
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
	if string(ret) != `{"a":2,"b":3}` {
		t.Fatal("ret:", string(ret))
	}
}

// -----------------------------------------------------------------------------

const codeRet = `

uint32be = {
	return 1
}

doc = {
	a uint32be
}
`

func TestReturn(t *testing.T) {

	r, err := NewFromString(codeRet, "")
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
	if string(ret) != `{"a":1}` {
		t.Fatal("ret:", string(ret))
	}
}

// -----------------------------------------------------------------------------

const codeRet2 = `

uint32be = {
    b4 uint8
    b3 uint8
    b2 uint8
    b1 uint8
	if true {
	    return (b4 << 24) | (b3 << 16) | (b2 << 8) | b1
	}
}

doc = {
	a uint32be
}
`

func TestReturn2(t *testing.T) {

	b := []byte{1, 2, 3, 4}

	r, err := NewFromString(codeRet2, "")
	if err != nil {
		t.Fatal("New failed:", err)
	}
	v, err := r.MatchBuffer(b)
	if err != nil {
		t.Fatal("Match failed:", err)
	}
	ret, err := json.Marshal(v)
	if err != nil {
		t.Fatal("json.Marshal failed:", err)
	}
	if string(ret) != `{"a":16909060}` {
		t.Fatal("ret:", string(ret))
	}
}

// -----------------------------------------------------------------------------

const codeGlobal = `

record = {
	do set(msgs, 1234, 35, 123, 36)
	let a = msgs[123]
}

init = {
	global msgs = mkmap("int:int")
}

doc = init record
`

func TestGlobal(t *testing.T) {

	b := []byte{1, 2, 3, 4}

	r, err := NewFromString(codeGlobal, "")
	if err != nil {
		t.Fatal("New failed:", err)
	}
	v, err := r.MatchBuffer(b)
	if err != nil {
		t.Fatal("Match failed:", err)
	}
	ret, err := json.Marshal(v)
	if err != nil {
		t.Fatal("json.Marshal failed:", err)
	}
	if string(ret) != `{"a":36}` {
		t.Fatal("ret:", string(ret))
	}
}

// -----------------------------------------------------------------------------

const codeGlobal2 = `

record = {
	let a = 3
	let b = 4
}

init = {
	global a = 1
}

doc = init record
`

func TestGlobal2(t *testing.T) {

	b := []byte{1, 2, 3, 4}

	r, err := NewFromString(codeGlobal2, "")
	if err != nil {
		t.Fatal("New failed:", err)
	}
	in := bufiox.NewReaderBuffer(b)
	ctx := NewContext()
	v, err := r.SafeMatch(in, ctx)
	if err != nil {
		t.Fatal("Match failed:", err)
	}
	ret, err := json.Marshal(v)
	if err != nil {
		t.Fatal("json.Marshal failed:", err)
	}
	if v, ok := ctx.Globals.Var("a"); !ok || v != 3 {
		t.Fatal("v != 3: v =", v, ok)
	}
	if string(ret) != `{"b":4}` {
		t.Fatal("ret:", string(ret))
	}
}

// -----------------------------------------------------------------------------

const codeMap = `

record = {
	do set(msgs, "foo", 35, "bar", 36)
	do println("msgs:", msgs)
	let a = get(msgs, "bar")
}

init = {
	global msgs = {}
}

doc = init record
`

func TestMap(t *testing.T) {

	b := []byte{1, 2, 3, 4}

	r, err := NewFromString(codeMap, "")
	if err != nil {
		t.Fatal("New failed:", err)
	}
	v, err := r.MatchBuffer(b)
	if err != nil {
		t.Fatal("Match failed:", err)
	}
	ret, err := json.Marshal(v)
	if err != nil {
		t.Fatal("json.Marshal failed:", err)
	}
	if string(ret) != `{"a":36}` {
		t.Fatal("ret:", string(ret))
	}
}

// -----------------------------------------------------------------------------

const codeUnset = `

record = {
	let a = 1
	let b = 2
	if true {
		do unset("a")
	}
}

doc = record
`

func TestUnset(t *testing.T) {

	b := []byte{1, 2, 3, 4}

	r, err := NewFromString(codeUnset, "")
	if err != nil {
		t.Fatal("New failed:", err)
	}
	v, err := r.MatchBuffer(b)
	if err != nil {
		t.Fatal("Match failed:", err)
	}
	ret, err := json.Marshal(v)
	if err != nil {
		t.Fatal("json.Marshal failed:", err)
	}
	if string(ret) != `{"b":2}` {
		t.Fatal("ret:", string(ret))
	}
}

// -----------------------------------------------------------------------------
