package bpl

import (
	"bufio"
	"errors"
	"io"
	"io/ioutil"
	"reflect"

	"github.com/qiniu/x/bufiox"
)

var (
	// ErrVarNotAssigned is returned when TypeVar.Elem is not assigned.
	ErrVarNotAssigned = errors.New("variable is not assigned")

	// ErrVarAssigned is returned when TypeVar.Elem is already assigned.
	ErrVarAssigned = errors.New("variable is already assigned")

	// ErrNotEOF is returned when current position is not at EOF.
	ErrNotEOF = errors.New("current position is not at EOF")
)

// -----------------------------------------------------------------------------

type nilType int

func (p nilType) Match(in *bufio.Reader, ctx *Context) (v interface{}, err error) {

	return nil, nil
}

func (p nilType) RetType() reflect.Type {

	return TyInterface
}

func (p nilType) SizeOf() int {

	return 0
}

// Nil is a matching unit that matches zero bytes.
//
var Nil Ruler = nilType(0)

// -----------------------------------------------------------------------------

type eof int

func (p eof) Match(in *bufio.Reader, ctx *Context) (v interface{}, err error) {

	_, err = in.Peek(1)
	if err == io.EOF {
		return nil, nil
	}
	return nil, ErrNotEOF
}

func (p eof) RetType() reflect.Type {

	return TyInterface
}

func (p eof) SizeOf() int {

	return 0
}

// EOF is a matching unit that matches EOF.
//
var EOF Ruler = eof(0)

// -----------------------------------------------------------------------------

type done int

func (p done) Match(in *bufio.Reader, ctx *Context) (v interface{}, err error) {

	_, err = in.WriteTo(ioutil.Discard)
	return
}

func (p done) RetType() reflect.Type {

	return TyInterface
}

func (p done) SizeOf() int {

	return -1
}

// Done is a matching unit that seeks current position to EOF.
//
var Done Ruler = done(0)

// -----------------------------------------------------------------------------

type and struct {
	rs []Ruler
}

func (p *and) Match(in *bufio.Reader, ctx *Context) (v interface{}, err error) {

	for _, r := range p.rs {
		_, err = r.Match(in, ctx)
		if err != nil {
			return
		}
	}
	return ctx.Dom(), nil
}

func (p *and) RetType() reflect.Type {

	return TyInterface
}

func (p *and) SizeOf() int {

	return -1
}

// And returns a matching unit that matches R1 R2 ... RN
//
func And(rs ...Ruler) Ruler {

	if len(rs) <= 1 {
		if len(rs) == 1 {
			return rs[0]
		}
		return Nil
	}
	return &and{rs: rs}
}

// -----------------------------------------------------------------------------

type seq struct {
	rs []Ruler
}

func (p *seq) Match(in *bufio.Reader, ctx *Context) (v interface{}, err error) {

	ret := ctx.requireVarSlice()
	for _, r := range p.rs {
		v, err = r.Match(in, ctx.NewSub())
		if err != nil {
			return
		}
		ret = append(ret, v)
	}
	ctx.dom = ret
	return ret, nil
}

func (p *seq) RetType() reflect.Type {

	return tyInterfaceSlice
}

func (p *seq) SizeOf() int {

	return -1
}

// Seq returns a matching unit that matches R1 R2 ... RN and returns matching result.
//
func Seq(rs ...Ruler) Ruler {

	return &seq{rs: rs}
}

// -----------------------------------------------------------------------------

type act struct {
	fn func(ctx *Context) error
}

func (p *act) Match(in *bufio.Reader, ctx *Context) (v interface{}, err error) {

	err = p.fn(ctx)
	if err != nil {
		return
	}
	return ctx.Dom(), nil
}

func (p *act) RetType() reflect.Type {

	return TyInterface
}

func (p *act) SizeOf() int {

	return -1
}

// Do returns a matching unit that executes action fn(ctx).
//
func Do(fn func(ctx *Context) error) Ruler {

	return &act{fn: fn}
}

// -----------------------------------------------------------------------------

type dyntype struct {
	r func(ctx *Context) (Ruler, error)
}

func (p *dyntype) Match(in *bufio.Reader, ctx *Context) (v interface{}, err error) {

	r, err := p.r(ctx)
	if err != nil {
		return
	}
	if r != nil {
		return r.Match(in, ctx)
	}
	return
}

func (p *dyntype) RetType() reflect.Type {

	return TyInterface
}

func (p *dyntype) SizeOf() int {

	return -1
}

// Dyntype returns a dynamic matching unit.
//
func Dyntype(r func(ctx *Context) (Ruler, error)) Ruler {

	return &dyntype{r: r}
}

// -----------------------------------------------------------------------------

type read struct {
	n func(ctx *Context) int
	r Ruler
}

func (p *read) Match(in *bufio.Reader, ctx *Context) (v interface{}, err error) {

	n := p.n(ctx)
	b := make([]byte, n)
	_, err = io.ReadFull(in, b)
	if err != nil {
		return
	}
	in = bufiox.NewReaderBuffer(b)
	return MatchStream(p.r, in, ctx)
}

func (p *read) RetType() reflect.Type {

	return p.r.RetType()
}

func (p *read) SizeOf() int {

	return -1
}

// Read returns a matching unit that reads n(ctx) bytes and matches R.
//
func Read(n func(ctx *Context) int, r Ruler) Ruler {

	return &read{r: r, n: n}
}

// -----------------------------------------------------------------------------

type skip struct {
	n func(ctx *Context) int
}

func (p *skip) Match(in *bufio.Reader, ctx *Context) (v interface{}, err error) {

	n := p.n(ctx)
	v, err = in.Discard(n)
	return
}

func (p *skip) RetType() reflect.Type {

	return tyInt
}

func (p *skip) SizeOf() int {

	return -1
}

// Skip returns a matching unit that skips n(ctx) bytes.
//
func Skip(n func(ctx *Context) int) Ruler {

	return &skip{n: n}
}

// -----------------------------------------------------------------------------

type ifType struct {
	cond func(ctx *Context) bool
	r    Ruler
}

func (p *ifType) Match(in *bufio.Reader, ctx *Context) (v interface{}, err error) {

	if p.cond(ctx) {
		return p.r.Match(in, ctx)
	}
	return
}

func (p *ifType) RetType() reflect.Type {

	return TyInterface
}

func (p *ifType) SizeOf() int {

	return -1
}

// If returns a matching unit that if cond(ctx) then matches it with R.
//
func If(cond func(ctx *Context) bool, r Ruler) Ruler {

	return &ifType{r: r, cond: cond}
}

// -----------------------------------------------------------------------------

type eval struct {
	expr func(ctx *Context) interface{}
	r    Ruler
}

func (p *eval) Match(in *bufio.Reader, ctx *Context) (v interface{}, err error) {

	fclose := false
	val := p.expr(ctx)
	switch v := val.(type) {
	case []byte:
		in = bufiox.NewReaderBuffer(v)
	case io.Reader:
		in = bufio.NewReader(v)
		fclose = true
	default:
		panic("eval <expr> must return []byte or io.Reader")
	}
	v, err = MatchStream(p.r, in, ctx)
	if fclose {
		if v, ok := val.(io.Closer); ok {
			v.Close()
		}
	}
	return
}

func (p *eval) RetType() reflect.Type {

	return p.r.RetType()
}

func (p *eval) SizeOf() int {

	return -1
}

// Eval returns a matching unit that eval expr(ctx) and matches it with R.
//
func Eval(expr func(ctx *Context) interface{}, r Ruler) Ruler {

	return &eval{r: r, expr: expr}
}

// -----------------------------------------------------------------------------

type assert struct {
	expr func(ctx *Context) bool
	msg  string
}

func (p *assert) Match(in *bufio.Reader, ctx *Context) (v interface{}, err error) {

	if p.expr(ctx) {
		return
	}
	panic(p.msg)
}

func (p *assert) RetType() reflect.Type {

	return TyInterface
}

func (p *assert) SizeOf() int {

	return -1
}

// Assert returns a matching unit that assert expr(ctx).
//
func Assert(expr func(ctx *Context) bool, msg string) Ruler {

	return &assert{msg: msg, expr: expr}
}

// -----------------------------------------------------------------------------

// A TypeVar is typeinfo of a `Struct` member.
//
type TypeVar struct {
	Name string
	Elem Ruler
}

// Assign assigns TypeVar.Elem.
//
func (p *TypeVar) Assign(r Ruler) error {

	if p.Elem != nil {
		return ErrVarAssigned
	}
	p.Elem = r
	return nil
}

// Match is required by a matching unit. see Ruler interface.
//
func (p *TypeVar) Match(in *bufio.Reader, ctx *Context) (v interface{}, err error) {

	r := p.Elem
	if r == nil {
		return 0, ErrVarNotAssigned
	}
	return r.Match(in, ctx)
}

// RetType returns matching result type.
//
func (p *TypeVar) RetType() reflect.Type {

	return p.Elem.RetType()
}

// SizeOf is required by a matching unit. see Ruler interface.
//
func (p *TypeVar) SizeOf() int {

	return p.Elem.SizeOf()
}

// -----------------------------------------------------------------------------
