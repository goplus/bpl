package bpl

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"reflect"
	"runtime/debug"

	"qiniupkg.com/x/bufiox.v7"
	"qlang.io/exec.v2"
)

// -----------------------------------------------------------------------------

// A Globals represents global variables.
//
type Globals struct {
	Impl map[string]interface{}
}

// NewGlobals returns a `Globals` instance.
//
func NewGlobals() Globals {

	return Globals{
		Impl: make(map[string]interface{}),
	}
}

// GetAndSetVar gets old value of a global variable and sets new value to it.
//
func (p Globals) GetAndSetVar(name string, v interface{}) (old interface{}, ok bool) {

	old, ok = p.Impl[name]
	p.Impl[name] = v
	return
}

// SetVar sets a global variable to new value.
//
func (p Globals) SetVar(name string, v interface{}) {

	p.Impl[name] = v
}

// Var returns value of a global variable.
//
func (p Globals) Var(name string) (v interface{}, ok bool) {

	v, ok = p.Impl[name]
	return
}

// -----------------------------------------------------------------------------

// A Context represents the matching context of bpl.
//
type Context struct {
	dom     interface{}
	Stack   *exec.Stack
	Parent  *Context
	Globals Globals
}

// NewContext returns a new matching Context.
//
func NewContext() *Context {

	gbl := NewGlobals()
	stk := exec.NewStack()
	return &Context{Globals: gbl, Stack: stk}
}

// NewSub returns a new sub Context.
//
func (p *Context) NewSub() *Context {

	return &Context{Parent: p, Globals: p.Globals, Stack: p.Stack}
}

func (p *Context) requireVarSlice() []interface{} {

	var vars []interface{}
	if p.dom == nil {
		vars = make([]interface{}, 0, 4)
	} else if domv, ok := p.dom.([]interface{}); ok {
		vars = domv
	} else {
		panic("dom type isn't []interface{}")
	}
	return vars
}

// SetVar sets a new variable to matching context.
//
func (p *Context) SetVar(name string, v interface{}) {

	if _, ok := p.Globals.Var(name); ok {
		panic(fmt.Errorf("variable `%s` exists globally", name))
	}

	var vars map[string]interface{}
	if p.dom == nil {
		vars = make(map[string]interface{})
		p.dom = vars
	} else if domv, ok := p.dom.(map[string]interface{}); ok {
		if _, ok = domv[name]; ok {
			panic(fmt.Errorf("variable `%s` exists in dom", name))
		}
		vars = domv
	} else {
		panic("dom type isn't map[string]interface{}")
	}
	vars[name] = v
}

// LetVar sets a variable to matching context.
//
func (p *Context) LetVar(name string, v interface{}) {

	if _, ok := p.Globals.Var(name); ok {
		p.Globals.SetVar(name, v)
		return
	}

	var vars map[string]interface{}
	if p.dom == nil {
		vars = make(map[string]interface{})
		p.dom = vars
	} else if domv, ok := p.dom.(map[string]interface{}); ok {
		vars = domv
	} else {
		panic("dom type isn't map[string]interface{}")
	}
	vars[name] = v
}

// Var gets a variable from matching context.
//
func (p *Context) Var(name string) (v interface{}, ok bool) {

	vars, ok := p.dom.(map[string]interface{})
	if ok {
		v, ok = vars[name]
	} else {
		panic("dom type isn't map[string]interface{}")
	}
	return
}

// SetDom set matching result of matching result.
//
func (p *Context) SetDom(v interface{}) {

	if p.dom == nil {
		p.dom = v
	} else {
		panic("dom was assigned already")
	}
}

// Dom returns matching result.
//
func (p *Context) Dom() interface{} {

	return p.dom
}

// -----------------------------------------------------------------------------

// A Ruler interface is required to a matching unit.
//
type Ruler interface {
	// Match matches input stream `in`, and returns matching result.
	Match(in *bufio.Reader, ctx *Context) (v interface{}, err error)

	// RetType returns matching result type.
	RetType() reflect.Type

	// SizeOf returns expected length of result. If length is variadic, it returns -1.
	SizeOf() int
}

// MatchStream matches a stream.
//
func MatchStream(r Ruler, in *bufio.Reader, ctx *Context) (v interface{}, err error) {

	glbs := ctx.Globals
	old, ok := glbs.GetAndSetVar("BPL_IN", in)
	v, err = r.Match(in, ctx)
	if ok {
		glbs.SetVar("BPL_IN", old)
	}
	return
}

// -----------------------------------------------------------------------------

type fileLine struct {
	r    Ruler
	file string
	line int
}

type errorAt struct {
	Err error
	Buf []byte
}

func (p *errorAt) Error() string {

	b := make([]byte, 0, 32)
	b = append(b, p.Err.Error()...)
	b = append(b, '\n')

	w := bytes.NewBuffer(b)
	d := hex.Dumper(w)
	d.Write(p.Buf)
	d.Close()
	return string(w.Bytes())
}

func (p *fileLine) Match(in *bufio.Reader, ctx *Context) (v interface{}, err error) {

	v, err = doMatch(p.r, in, ctx)
	if err != nil {
		if _, ok := err.(*exec.Error); !ok {
			err = &exec.Error{
				Err:   &errorAt{Err: err, Buf: bufiox.Buffer(in)},
				File:  p.file,
				Line:  p.line,
				Stack: debug.Stack(),
			}
		}
	}
	return
}

func (p *fileLine) RetType() reflect.Type {

	return p.r.RetType()
}

func (p *fileLine) SizeOf() int {

	return p.r.SizeOf()
}

func doMatch(R Ruler, in *bufio.Reader, ctx *Context) (v interface{}, err error) {

	defer func() {
		if e := recover(); e != nil {
			switch v := e.(type) {
			case string:
				err = errors.New(v)
			case error:
				err = v
			default:
				panic(e)
			}
		}
	}()

	return R.Match(in, ctx)
}

// FileLine is a matching rule that reports error file line when error occurs.
//
func FileLine(file string, line int, R Ruler) Ruler {

	if _, ok := R.(*fileLine); ok {
		return R
	}
	return &fileLine{r: R, file: file, line: line}
}

// -----------------------------------------------------------------------------
