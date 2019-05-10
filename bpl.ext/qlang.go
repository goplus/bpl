package bpl

import (
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"

	"qlang.io/exec.v2"
	"qlang.io/qlang.spec.v1"
	"qlang.io/qlang/bytes"
	"qlang.io/qlang/crypto/hmac"
	"qlang.io/qlang/crypto/md5"
	"qlang.io/qlang/crypto/sha1"
	"qlang.io/qlang/crypto/sha256"
	"qlang.io/qlang/encoding/hex"
	"qlang.io/qlang/encoding/json"
	"qlang.io/qlang/errors"
	"qlang.io/qlang/io"
	"qlang.io/qlang/strings"

	// import qlang builtin
	_ "qlang.io/qlang/builtin"
	qstrconv "qlang.io/qlang/strconv"
)

// -----------------------------------------------------------------------------

func exit(code int) {

	panic(code)
}

func init() {

	osExports := map[string]interface{}{
		"exit": exit,
	}

	httpExports := map[string]interface{}{
		"readRequest":  http.ReadRequest,
		"readResponse": http.ReadResponse,
	}

	var ioutilExports = map[string]interface{}{
		"nopCloser": ioutil.NopCloser,
		"readAll":   ioutil.ReadAll,
		"discard":   ioutil.Discard,
	}

	qlang.Import("", exports)
	qlang.Import("bytes", bytes.Exports)
	qlang.Import("md5", md5.Exports)
	qlang.Import("sha1", sha1.Exports)
	qlang.Import("sha256", sha256.Exports)
	qlang.Import("hmac", hmac.Exports)
	qlang.Import("errors", errors.Exports)
	qlang.Import("json", json.Exports)
	qlang.Import("hex", hex.Exports)
	qlang.Import("io", io.Exports)
	qlang.Import("ioutil", ioutilExports)
	qlang.Import("os", osExports)
	qlang.Import("http", httpExports)
	qlang.Import("strconv", qstrconv.Exports)
	qlang.Import("strings", strings.Exports)
}

// Fntable returns the qlang compiler's function table. It is required by tpl.Interpreter engine.
//
func (p *Compiler) Fntable() map[string]interface{} {

	return qlang.Fntable
}

// -----------------------------------------------------------------------------

func castInt(a interface{}) (int, bool) {

	switch a1 := a.(type) {
	case int:
		return a1, true
	case int32:
		return int(a1), true
	case int64:
		return int(a1), true
	case int16:
		return int(a1), true
	case int8:
		return int(a1), true
	case uint:
		return int(a1), true
	case uint32:
		return int(a1), true
	case uint64:
		return int(a1), true
	case uint16:
		return int(a1), true
	case uint8:
		return int(a1), true
	}
	return 0, false
}

func toInt(a interface{}, msg string) int {

	if v, ok := castInt(a); ok {
		return v
	}
	panic(msg)
}

func toBool(a interface{}, msg string) bool {

	if v, ok := a.(bool); ok {
		return v
	}
	if v, ok := castInt(a); ok {
		return v != 0
	}
	panic(msg)
}

// CallFn generates a function call instruction. It is required by tpl.Interpreter engine.
//
func (p *Compiler) CallFn(fn interface{}) {

	p.code.Block(exec.Call(fn))
}

func eq(a, b interface{}) bool {

	if a1, ok := castInt(a); ok {
		switch b1 := b.(type) {
		case int:
			return a1 == b1
		}
	}
	if a1, ok := a.(string); ok {
		switch b1 := b.(type) {
		case string:
			return a1 == b1
		}
	}
	panicUnsupportedOp2("==", a, b)
	return false
}

func and(a, b bool) bool {

	return a && b
}

func or(a, b bool) bool {

	return a || b
}

func panicUnsupportedOp2(op string, a, b interface{}) interface{} {

	ta := typeString(a)
	tb := typeString(b)
	panic("unsupported operator: " + ta + op + tb)
}

func typeString(a interface{}) string {

	if a == nil {
		return "nil"
	}
	return reflect.TypeOf(a).String()
}

// -----------------------------------------------------------------------------

func (p *Compiler) popArity() int {

	return p.popConstInt()
}

func (p *Compiler) popConstInt() int {

	if v, ok := p.gstk.Pop(); ok {
		if val, ok := v.(int); ok {
			return val
		}
	}
	panic("no int")
}

func (p *Compiler) arity(arity int) {

	p.gstk.Push(arity)
}

func (p *Compiler) call() {

	variadic := p.popArity()
	arity := p.popArity()
	if variadic != 0 {
		if arity == 0 {
			panic("what do you mean of `...`?")
		}
		p.code.Block(exec.CallFnv(arity))
	} else {
		p.code.Block(exec.CallFn(arity))
	}
}

func (p *Compiler) ref(name string) {

	var instr exec.Instr
	if v, ok := p.consts[name]; ok {
		instr = exec.Push(v)
	} else {
		instr = exec.Ref(name)
	}
	p.code.Block(instr)
}

func (p *Compiler) mref(name string) {

	p.code.Block(exec.MemberRef(name))
}

func (p *Compiler) pushi(v int) {

	p.code.Block(exec.Push(v))
}

func (p *Compiler) pushs(lit string) {

	v, err := strconv.Unquote(lit)
	if err != nil {
		panic("invalid string `" + lit + "`: " + err.Error())
	}
	p.code.Block(exec.Push(v))
}

func (p *Compiler) pushc(lit string) {

	v, multibyte, tail, err := strconv.UnquoteChar(lit[1:len(lit)-1], '\'')
	if err != nil {
		panic("invalid char `" + lit + "`: " + err.Error())
	}
	if tail != "" || multibyte {
		panic("invalid char: " + lit)
	}
	p.code.Block(exec.Push(byte(v)))
}

func (p *Compiler) cpushi(v int) {

	p.gstk.Push(v)
}

func (p *Compiler) fnConst(name string) {

	p.consts[name] = p.popConstInt()
}

// -----------------------------------------------------------------------------

func (p *Compiler) fnMap() {

	arity := p.popArity()
	p.code.Block(exec.Call(qlang.MapFrom, arity*2))
}

func (p *Compiler) fnSlice() {

	arity := p.popArity()
	p.code.Block(exec.SliceFrom(arity))
}

func (p *Compiler) index() {

	arity2 := p.popArity()
	arityMid := p.popArity()
	arity1 := p.popArity()

	if arityMid == 0 {
		if arity1 == 0 {
			panic("call operator[] without index")
		}
		p.code.Block(exec.Get)
	} else {
		p.code.Block(exec.Op3(qlang.SubSlice, arity1 != 0, arity2 != 0))
	}
}

// -----------------------------------------------------------------------------

// DumpCode is mode how to dump code.
// 1 means to dump code with `rem` instruction; 2 means to dump clean code; 0 means don't dump code.
//
var DumpCode int

func (p *Compiler) codeLine(src interface{}) {

	ipt := p.ipt
	if ipt == nil {
		return
	}

	f := ipt.FileLine(src)
	p.code.CodeLine(f.File, f.Line)
	if DumpCode == 1 {
		text := string(ipt.Source(src))
		p.code.Block(exec.Rem(f.File, f.Line, text))
	}
}

// -----------------------------------------------------------------------------
