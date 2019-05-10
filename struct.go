package bpl

import (
	"bufio"
	"fmt"
	"reflect"
	"strings"

	"qiniupkg.com/x/log.v7"
)

// -----------------------------------------------------------------------------

type ret func(ctx *Context) (v interface{}, err error)

func (p ret) Match(in *bufio.Reader, ctx *Context) (v interface{}, err error) {

	v, err = p(ctx)
	if err != nil {
		return
	}
	ctx.dom = v
	return
}

func (p ret) RetType() reflect.Type {

	return TyInterface
}

func (p ret) SizeOf() int {

	return -1
}

// Return returns a matching unit that returns fnRet(ctx).
//
func Return(fnRet func(ctx *Context) (v interface{}, err error)) Ruler {

	return ret(fnRet)
}

// -----------------------------------------------------------------------------

// A Member is typeinfo of a `Struct` member.
//
type Member struct {
	Name string
	Type Ruler
}

// Match is required by a matching unit. see Ruler interface.
//
func (p *Member) Match(in *bufio.Reader, ctx *Context) (v interface{}, err error) {

	v, err = p.Type.Match(in, ctx.NewSub())
	if err != nil {
		return
	}
	if p.Name != "_" {
		ctx.SetVar(p.Name, v)
	}
	return
}

// RetType returns matching result type.
//
func (p *Member) RetType() reflect.Type {

	return p.Type.RetType()
}

// SizeOf is required by a matching unit. see Ruler interface.
//
func (p *Member) SizeOf() int {

	return p.Type.SizeOf()
}

// -----------------------------------------------------------------------------

type structType struct {
	rulers []Ruler
	size   int
}

func (p *structType) Match(in *bufio.Reader, ctx *Context) (v interface{}, err error) {

	for _, r := range p.rulers {
		_, err = r.Match(in, ctx)
		if err != nil {
			return
		}
	}
	return ctx.Dom(), nil
}

func (p *structType) RetType() reflect.Type {

	return TyInterface
}

func (p *structType) SizeOf() int {

	if p.size == -2 {
		p.size = p.sizeof()
	}
	return p.size
}

func (p *structType) sizeof() int {

	size := 0
	for _, r := range p.rulers {
		if n := r.SizeOf(); n < 0 {
			size = -1
			break
		} else {
			size += n
		}
	}
	return size
}

// Struct returns a compound matching unit.
//
func Struct(members []Ruler) Ruler {

	n := len(members)
	if n == 0 {
		return Nil
	}

	return &structType{rulers: members, size: -2}
}

// -----------------------------------------------------------------------------

func structFrom(t reflect.Type) (r Ruler, err error) {

	n := t.NumField()
	rulers := make([]Ruler, n)
	for i := 0; i < n; i++ {
		sf := t.Field(i)
		r, err = TypeFrom(sf.Type)
		if err != nil {
			log.Warn("bpl.TypeFrom failed:", err)
			return
		}
		rulers[i] = &Member{Name: strings.ToLower(sf.Name), Type: r}
	}
	return Struct(rulers), nil
}

// TypeFrom creates a matching unit from a Go type.
//
func TypeFrom(t reflect.Type) (r Ruler, err error) {

retry:
	kind := t.Kind()
	switch {
	case kind == reflect.Struct:
		return structFrom(t)
	case kind >= reflect.Int8 && kind <= reflect.Float64:
		return BaseType(kind), nil
	case kind == reflect.String:
		return CString, nil
	case kind == reflect.Ptr:
		t = t.Elem()
		goto retry
	}
	return nil, fmt.Errorf("bpl.TypeFrom: unsupported type - %v", t)
}

// -----------------------------------------------------------------------------
