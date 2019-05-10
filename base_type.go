package bpl

import (
	"bufio"
	"encoding/binary"
	"io"
	"reflect"
	"unsafe"

	"qiniupkg.com/x/log.v7"
)

// -----------------------------------------------------------------------------

var (
	tyInt            = reflect.TypeOf(int(0))
	tyInt8           = reflect.TypeOf(int8(0))
	tyInt16          = reflect.TypeOf(int16(0))
	tyInt32          = reflect.TypeOf(int32(0))
	tyInt64          = reflect.TypeOf(int64(0))
	tyUint           = reflect.TypeOf(uint(0))
	tyUint8          = reflect.TypeOf(uint8(0))
	tyUint16         = reflect.TypeOf(uint16(0))
	tyUint32         = reflect.TypeOf(uint32(0))
	tyUint64         = reflect.TypeOf(uint64(0))
	tyFloat32        = reflect.TypeOf(float32(0))
	tyFloat64        = reflect.TypeOf(float64(0))
	tyString         = reflect.TypeOf(string(""))
	tyByteSlice      = reflect.TypeOf([]byte(nil))
	TyInterface      = reflect.TypeOf((*interface{})(nil)).Elem()
	tyInterfaceSlice = reflect.SliceOf(TyInterface)
)

// -----------------------------------------------------------------------------

// A BaseType represents a matching unit of a builtin fixed size type.
//
type BaseType uint

type baseTypeInfo struct {
	read   func(in *bufio.Reader) (v interface{}, err error)
	newn   func(n int) interface{}
	typ    reflect.Type
	sizeOf int
}

var baseTypes = [...]baseTypeInfo{
	reflect.Int8:    {readInt8, newInt8n, tyInt8, 1},
	reflect.Int16:   {readInt16, newInt16n, tyInt16, 2},
	reflect.Int32:   {readInt32, newInt32n, tyInt32, 4},
	reflect.Int64:   {readInt64, newInt64n, tyInt64, 8},
	reflect.Uint8:   {readUint8, newUint8n, tyUint8, 1},
	reflect.Uint16:  {readUint16, newUint16n, tyUint16, 2},
	reflect.Uint32:  {readUint32, newUint32n, tyUint32, 4},
	reflect.Uint64:  {readUint64, newUint64n, tyUint64, 8},
	reflect.Float32: {readFloat32, newFloat32n, tyFloat32, 4},
	reflect.Float64: {readFloat64, newFloat64n, tyFloat64, 8},
}

func readInt8(in *bufio.Reader) (v interface{}, err error) {

	t, err := in.ReadByte()
	return int8(t), err
}

func readUint8(in *bufio.Reader) (v interface{}, err error) {

	return in.ReadByte()
}

func readInt16(in *bufio.Reader) (v interface{}, err error) {

	t1, err := in.ReadByte()
	if err != nil {
		return
	}
	t2, err := in.ReadByte()
	return (int16(t2) << 8) | int16(t1), err
}

func readUint16(in *bufio.Reader) (v interface{}, err error) {

	t1, err := in.ReadByte()
	if err != nil {
		return
	}
	t2, err := in.ReadByte()
	return (uint16(t2) << 8) | uint16(t1), err
}

func readInt32(in *bufio.Reader) (v interface{}, err error) {

	t, err := in.Peek(4)
	if err != nil {
		return
	}
	v = int32(binary.LittleEndian.Uint32(t))
	in.Discard(4)
	return
}

func readUint32(in *bufio.Reader) (v interface{}, err error) {

	t, err := in.Peek(4)
	if err != nil {
		return
	}
	v = binary.LittleEndian.Uint32(t)
	in.Discard(4)
	return
}

func readInt64(in *bufio.Reader) (v interface{}, err error) {

	t, err := in.Peek(8)
	if err != nil {
		return
	}
	v = int64(binary.LittleEndian.Uint64(t))
	in.Discard(8)
	return
}

func readUint64(in *bufio.Reader) (v interface{}, err error) {

	t, err := in.Peek(8)
	if err != nil {
		return
	}
	v = binary.LittleEndian.Uint64(t)
	in.Discard(8)
	return
}

func readFloat32(in *bufio.Reader) (v interface{}, err error) {

	t, err := in.Peek(4)
	if err != nil {
		return
	}
	v = *(*float32)(unsafe.Pointer(&t[0]))
	in.Discard(4)
	return
}

func readFloat64(in *bufio.Reader) (v interface{}, err error) {

	t, err := in.Peek(8)
	if err != nil {
		return
	}
	v = *(*float64)(unsafe.Pointer(&t[0]))
	in.Discard(8)
	return
}

func newInt8n(n int) interface{} {

	return make([]int8, n)
}

func newUint8n(n int) interface{} {

	return make([]uint8, n)
}

func newInt16n(n int) interface{} {

	return make([]int16, n)
}

func newUint16n(n int) interface{} {

	return make([]uint16, n)
}

func newInt32n(n int) interface{} {

	return make([]int32, n)
}

func newUint32n(n int) interface{} {

	return make([]uint32, n)
}

func newInt64n(n int) interface{} {

	return make([]int64, n)
}

func newUint64n(n int) interface{} {

	return make([]uint64, n)
}

func newFloat32n(n int) interface{} {

	return make([]float32, n)
}

func newFloat64n(n int) interface{} {

	return make([]float64, n)
}

// Match is required by a matching unit. see Ruler interface.
//
func (p BaseType) Match(in *bufio.Reader, ctx *Context) (v interface{}, err error) {

	v, err = baseTypes[p].read(in)
	return
}

// RetType returns matching result type.
//
func (p BaseType) RetType() reflect.Type {

	return baseTypes[p].typ
}

// SizeOf is required by a matching unit. see Ruler interface.
//
func (p BaseType) SizeOf() int {

	return baseTypes[p].sizeOf
}

var (
	// Int8 is the matching unit for int8
	Int8 = BaseType(reflect.Int8)

	// Int16 is the matching unit for int16
	Int16 = BaseType(reflect.Int16)

	// Int32 is the matching unit for int32
	Int32 = BaseType(reflect.Int32)

	// Int64 is the matching unit for int64
	Int64 = BaseType(reflect.Int64)

	// Uint8 is the matching unit for uint8
	Uint8 = BaseType(reflect.Uint8)

	// Uint16 is the matching unit for uint16
	Uint16 = BaseType(reflect.Uint16)

	// Uint32 is the matching unit for uint32
	Uint32 = BaseType(reflect.Uint32)

	// Uint64 is the matching unit for uint64
	Uint64 = BaseType(reflect.Uint64)

	// Float32 is the matching unit for float32
	Float32 = BaseType(reflect.Float32)

	// Float64 is the matching unit for float64
	Float64 = BaseType(reflect.Float64)
)

// -----------------------------------------------------------------------------

type cstring int

func (p cstring) Match(in *bufio.Reader, ctx *Context) (v interface{}, err error) {

	b, err := in.ReadBytes(0)
	if err != nil {
		return
	}
	return string(b[:len(b)-1]), nil
}

func (p cstring) RetType() reflect.Type {

	return tyString
}

func (p cstring) SizeOf() int {

	return -1
}

// CString is a matching unit that matches a C style string.
//
var CString Ruler = cstring(0)

// -----------------------------------------------------------------------------

type charType int

func (p charType) Match(in *bufio.Reader, ctx *Context) (v interface{}, err error) {

	return in.ReadByte()
}

func (p charType) RetType() reflect.Type {

	return tyUint8
}

func (p charType) SizeOf() int {

	return 1
}

// Char is a matching unit that matches a character.
//
var Char Ruler = charType(0)

// -----------------------------------------------------------------------------

type fixedType struct {
	typ reflect.Type
}

func (p *fixedType) Match(in *bufio.Reader, ctx *Context) (v interface{}, err error) {

	typ := p.typ
	size := typ.Size()
	val := reflect.New(typ)
	b := (*[1 << 30]byte)(unsafe.Pointer(val.UnsafeAddr()))
	_, err = io.ReadFull(in, b[:size])
	if err != nil {
		log.Warn("fixedType.Match: io.ReadFull failed -", err)
		return
	}
	return val.Interface(), nil
}

func (p *fixedType) RetType() reflect.Type {

	return p.typ
}

func (p *fixedType) SizeOf() int {

	return int(p.typ.Size())
}

// FixedType returns a matching unit that matches a C style fixed size struct.
//
func FixedType(t reflect.Type) Ruler {

	return &fixedType{typ: t}
}

// -----------------------------------------------------------------------------

type uintbe int

func (p uintbe) Match(in *bufio.Reader, ctx *Context) (v interface{}, err error) {

	t, err := in.Peek(int(p))
	if err != nil {
		return
	}
	var val uint
	for i := 0; i < int(p); i++ {
		val = (val << 8) | uint(t[i])
	}
	in.Discard(int(p))
	return val, nil
}

func (p uintbe) RetType() reflect.Type {

	return tyUint
}

func (p uintbe) SizeOf() int {

	return int(p)
}

// Uintbe returns a matching unit that matches a uintbe(n) type.
//
func Uintbe(n int) Ruler {

	if n < 2 || n > 8 {
		panic("Uintbe: invalid argument (n >= 2 && n <= 8)")
	}
	return uintbe(n)
}

// -----------------------------------------------------------------------------

type uintle int

func (p uintle) Match(in *bufio.Reader, ctx *Context) (v interface{}, err error) {

	t, err := in.Peek(int(p))
	if err != nil {
		return
	}
	var val uint
	for i := int(p); i > 0; {
		i--
		val = (val << 8) | uint(t[i])
	}
	in.Discard(int(p))
	return val, nil
}

func (p uintle) RetType() reflect.Type {

	return tyUint
}

func (p uintle) SizeOf() int {

	return int(p)
}

// Uintle returns a matching unit that matches a uintle(n) type.
//
func Uintle(n int) Ruler {

	if n < 2 || n > 8 {
		panic("Uintle: invalid argument (n >= 2 && n <= 8)")
	}
	return uintle(n)
}

// Uint24 is a matching unit that matches a uintle(3) type.
//
var Uint24 = Uintle(3)

// -----------------------------------------------------------------------------

func float32frombits(b uint32) float32 { return *(*float32)(unsafe.Pointer(&b)) }

type float32be int

func (p float32be) Match(in *bufio.Reader, ctx *Context) (v interface{}, err error) {

	t, err := in.Peek(4)
	if err != nil {
		return
	}
	v = float32frombits(binary.BigEndian.Uint32(t))
	in.Discard(4)
	return
}

func (p float32be) RetType() reflect.Type {

	return tyFloat32
}

func (p float32be) SizeOf() int {

	return 4
}

// Float32be returns a matching unit that matches a float32be type.
//
var Float32be Ruler = float32be(0)

// -----------------------------------------------------------------------------

func float64frombits(b uint64) float64 { return *(*float64)(unsafe.Pointer(&b)) }

type float64be int

func (p float64be) Match(in *bufio.Reader, ctx *Context) (v interface{}, err error) {

	t, err := in.Peek(8)
	if err != nil {
		return
	}
	v = float64frombits(binary.BigEndian.Uint64(t))
	in.Discard(8)
	return
}

func (p float64be) RetType() reflect.Type {

	return tyFloat64
}

func (p float64be) SizeOf() int {

	return 8
}

// Float64be returns a matching unit that matches a float64be type.
//
var Float64be Ruler = float64be(0)

// -----------------------------------------------------------------------------
