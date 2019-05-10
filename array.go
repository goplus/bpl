package bpl

import (
	"bufio"
	"io"
	"reflect"
)

// -----------------------------------------------------------------------------

func valueOf(v interface{}, t reflect.Type) reflect.Value {

	if v != nil {
		return reflect.ValueOf(v)
	}
	return reflect.Zero(t)
}

func matchArray1(R Ruler, in *bufio.Reader, ctx *Context, fCheckNil bool) (v interface{}, err error) {

	t := R.RetType()
	ret := reflect.MakeSlice(reflect.SliceOf(t), 0, 4)
	for {
		_, err = in.Peek(1)
		if err != nil {
			if err == io.EOF {
				if fCheckNil {
					return
				}
				return ret.Interface(), nil
			}
			return
		}
		v, err = R.Match(in, ctx.NewSub())
		if err != nil {
			return
		}
		ret = reflect.Append(ret, valueOf(v, t))
		fCheckNil = false
	}
}

func matchArray(R Ruler, n int, in *bufio.Reader, ctx *Context) (v interface{}, err error) {

	if n == 0 {
		return
	}

	t := R.RetType()
	ret := reflect.MakeSlice(reflect.SliceOf(t), 0, n)
	for i := 0; i < n; i++ {
		v, err = R.Match(in, ctx.NewSub())
		if err != nil {
			return
		}
		ret = reflect.Append(ret, valueOf(v, t))
	}
	return ret.Interface(), nil
}

// -----------------------------------------------------------------------------

type array1 struct {
	r Ruler
}

func (p *array1) Match(in *bufio.Reader, ctx *Context) (v interface{}, err error) {

	return matchArray1(p.r, in, ctx, true)
}

func (p *array1) RetType() reflect.Type {

	return reflect.SliceOf(p.r.RetType())
}

func (p *array1) SizeOf() int {

	return -1
}

// Array1 returns a matching unit that matches R+
//
func Array1(R Ruler) Ruler {

	if R == Uint8 {
		return ByteArray1
	}
	return &array1{r: R}
}

// -----------------------------------------------------------------------------

type array0 struct {
	r Ruler
}

func (p *array0) Match(in *bufio.Reader, ctx *Context) (v interface{}, err error) {

	return matchArray1(p.r, in, ctx, false)
}

func (p *array0) RetType() reflect.Type {

	return reflect.SliceOf(p.r.RetType())
}

func (p *array0) SizeOf() int {

	return -1
}

// Array0 returns a matching unit that matches R*
//
func Array0(R Ruler) Ruler {

	if R == Uint8 {
		return ByteArray0
	}
	return &array0{r: R}
}

// Array01 returns a matching unit that matches R?
//
func Array01(R Ruler) Ruler {

	return Repeat01(R)
}

// -----------------------------------------------------------------------------

type array struct {
	r Ruler
	n int
}

func (p *array) Match(in *bufio.Reader, ctx *Context) (v interface{}, err error) {

	n := p.n
	return matchArray(p.r, n, in, ctx)
}

func (p *array) RetType() reflect.Type {

	return reflect.SliceOf(p.r.RetType())
}

func (p *array) SizeOf() int {

	size := p.r.SizeOf()
	if size < 0 {
		return -1
	}
	return p.n * size
}

// Array returns a matching unit that matches R n times.
//
func Array(r Ruler, n int) Ruler {

	//TODO:
	//if t, ok := r.(BaseType); ok {
	//	return &baseArray{r: t, n: n}
	//}
	if r == Char {
		return charArray(n)
	} else if r == Uint8 {
		return byteArray(n)
	}
	return &array{r: r, n: n}
}

// -----------------------------------------------------------------------------

type dynarray struct {
	r Ruler
	n func(ctx *Context) int
}

func (p *dynarray) Match(in *bufio.Reader, ctx *Context) (v interface{}, err error) {

	n := p.n(ctx)
	return matchArray(p.r, n, in, ctx)
}

func (p *dynarray) RetType() reflect.Type {

	return reflect.SliceOf(p.r.RetType())
}

func (p *dynarray) SizeOf() int {

	return -1
}

// Dynarray returns a matching unit that matches R n(ctx) times.
//
func Dynarray(r Ruler, n func(ctx *Context) int) Ruler {

	//TODO:
	//if t, ok := r.(BaseType); ok {
	//	return &baseDynarray{r: t, n: n}
	//}
	if r == Char {
		return charDynarray(n)
	} else if r == Uint8 {
		return byteDynarray(n)
	}
	return &dynarray{r: r, n: n}
}

// -----------------------------------------------------------------------------
