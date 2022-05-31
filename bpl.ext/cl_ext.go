package bpl

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/goplus/bpl"
	"github.com/qiniu/text/tpl/interpreter.util"
	"github.com/xushiwei/qlang/exec"
)

// -----------------------------------------------------------------------------

func (p *Compiler) eval(ctx *bpl.Context, start, end int) interface{} {

	vars, hasDom := ctx.Dom().(map[string]interface{})
	if vars == nil {
		vars = make(map[string]interface{})
	}
	code := &p.code
	stk := ctx.Stack
	parent := exec.NewSimpleContext(ctx.Globals.Impl, nil, nil, nil)
	ectx := exec.NewSimpleContext(vars, stk, code, parent)
	code.Exec(start, end, stk, ectx)
	if !hasDom && len(vars) > 0 { // update dom
		ctx.SetDom(vars)
	}
	v, _ := stk.Pop()
	return v
}

// -----------------------------------------------------------------------------

type exprBlock struct {
	start int
	end   int
}

func (p *Compiler) istart() {

	p.idxStart = p.code.Len()
}

func (p *Compiler) iend() {

	end := p.code.Len()
	p.gstk.Push(&exprBlock{start: p.idxStart, end: end})
}

func (p *Compiler) popExpr() *exprBlock {

	if v, ok := p.gstk.Pop(); ok {
		if e, ok := v.(*exprBlock); ok {
			return e
		}
	}
	panic("no index expression")
}

func (p *Compiler) array() {

	e := p.popExpr()
	stk := p.stk
	i := len(stk) - 1
	n := func(ctx *bpl.Context) int {
		v := p.eval(ctx.Parent, e.start, e.end)
		return toInt(v, "index isn't an integer expression")
	}
	stk[i] = bpl.Dynarray(stk[i].(bpl.Ruler), n)
}

func (p *Compiler) array0() {

	stk := p.stk
	i := len(stk) - 1
	stk[i] = bpl.Array0(stk[i].(bpl.Ruler))
}

func (p *Compiler) array1() {

	stk := p.stk
	i := len(stk) - 1
	stk[i] = bpl.Array1(stk[i].(bpl.Ruler))
}

// -----------------------------------------------------------------------------

func (p *Compiler) casei(v int) {

	p.gstk.Push(v)
}

func (p *Compiler) cases(lit string) {

	v, err := strconv.Unquote(lit)
	if err != nil {
		panic("invalid string `" + lit + "`: " + err.Error())
	}
	p.gstk.Push(v)
}

func (p *Compiler) source(v interface{}) {

	p.gstk.Push(v)
}

func (p *Compiler) popRule() bpl.Ruler {

	n := len(p.stk) - 1
	r := p.stk[n].(bpl.Ruler)
	p.stk = p.stk[:n]
	return r
}

func (p *Compiler) popRules(m int) []bpl.Ruler {

	if m == 0 {
		return nil
	}
	stk := p.stk
	n := len(stk)
	rulers := clone(stk[n-m:])
	p.stk = stk[:n-m]
	return rulers
}

func sourceOf(engine interpreter.Engine, src interface{}) string {

	b := engine.Source(src)
	return strings.Trim(string(b), " \t\r\n")
}

func (p *Compiler) fnCase(engine interpreter.Engine) {

	var defaultR bpl.Ruler
	if p.popArity() != 0 {
		defaultR = p.popRule()
	}

	arity := p.popArity()

	stk := p.stk
	n := len(stk)
	caseRs := clone(stk[n-arity:])
	caseExprAndSources := p.gstk.PopNArgs(arity << 1)
	e := p.popExpr()
	srcSw, _ := p.gstk.Pop()
	r := func(ctx *bpl.Context) (bpl.Ruler, error) {
		v := p.eval(ctx, e.start, e.end)
		for i := 0; i < len(caseExprAndSources); i += 2 {
			expr := caseExprAndSources[i]
			if eq(v, expr) {
				if SetCaseType {
					key := sourceOf(engine, srcSw)
					val := sourceOf(engine, caseExprAndSources[i+1])
					ctx.SetVar(key+".kind", val)
				}
				return caseRs[i>>1], nil
			}
		}
		if defaultR != nil {
			return defaultR, nil
		}
		return nil, fmt.Errorf("case `%s(=%v)` is not found", sourceOf(engine, srcSw), v)
	}
	stk[n-arity] = bpl.Dyntype(r)
	p.stk = stk[:n-arity+1]
}

// -----------------------------------------------------------------------------

func (p *Compiler) fnIf() {

	var elseR bpl.Ruler
	if p.popArity() != 0 {
		elseR = p.popRule()
	} else {
		elseR = bpl.Nil
	}

	arity := p.popArity() + 1

	stk := p.stk
	n := len(stk)
	bodyRs := clone(stk[n-arity:])
	condExprs := p.gstk.PopNArgs(arity)

	arityOptimized := 0
	for i := 0; i < arity; i++ {
		e := condExprs[i].(*exprBlock)
		if e.end-e.start == 1 {
			if v, ok := p.code.CheckConst(e.start); ok {
				if toBool(v, "condition isn't a boolean expression") { // true
					arityOptimized = i
					elseR = bodyRs[i]
					break
				}
				continue // false
			}
		}
		if arityOptimized != i {
			bodyRs[arityOptimized] = bodyRs[i]
			condExprs[arityOptimized] = condExprs[i]
		}
		arityOptimized++
	}

	dynR := elseR
	if arityOptimized > 0 {
		r := func(ctx *bpl.Context) (bpl.Ruler, error) {
			for i := 0; i < arityOptimized; i++ {
				e := condExprs[i].(*exprBlock)
				v := p.eval(ctx, e.start, e.end)
				if toBool(v, "condition isn't a boolean expression") {
					return bodyRs[i], nil
				}
			}
			return elseR, nil
		}
		dynR = bpl.Dyntype(r)
	}
	stk[n-arity] = dynR
	p.stk = stk[:n-arity+1]
}

// -----------------------------------------------------------------------------

func (p *Compiler) fnEval() {

	e := p.popExpr()
	stk := p.stk
	i := len(stk) - 1
	expr := func(ctx *bpl.Context) interface{} {
		return p.eval(ctx, e.start, e.end)
	}
	stk[i] = bpl.Eval(expr, stk[i].(bpl.Ruler))
}

// -----------------------------------------------------------------------------

func (p *Compiler) fnDo() {

	e := p.popExpr()
	fn := func(ctx *bpl.Context) error {
		p.eval(ctx, e.start, e.end)
		return nil
	}
	p.stk = append(p.stk, bpl.Do(fn))
}

// -----------------------------------------------------------------------------

func (p *Compiler) fnLet() {

	e := p.popExpr()
	arity := p.popArity()
	stk := p.stk
	n := len(stk) - arity
	if arity == 1 {
		name := stk[n].(string)
		fn := func(ctx *bpl.Context) error {
			v := p.eval(ctx, e.start, e.end)
			ctx.LetVar(name, v)
			return nil
		}
		stk[n] = bpl.Do(fn)
	} else {
		names := cloneNames(stk[n:])
		fn := func(ctx *bpl.Context) error {
			v := p.eval(ctx, e.start, e.end)
			multiAssignFromSlice(names, v, ctx)
			return nil
		}
		stk[n] = bpl.Do(fn)
		p.stk = stk[:n+1]
	}
}

func multiAssignFromSlice(names []string, val interface{}, ctx *bpl.Context) {

	v := reflect.ValueOf(val)
	if v.Kind() != reflect.Slice {
		panic("expression of multi assignment must be a slice")
	}

	n := v.Len()
	arity := len(names)
	if arity != n {
		panic(fmt.Errorf("multi assignment error: require %d variables, but we got %d", n, arity))
	}

	for i, name := range names {
		ctx.LetVar(name, v.Index(i).Interface())
	}
}

// -----------------------------------------------------------------------------

func (p *Compiler) fnGlobal() {

	e := p.popExpr()
	stk := p.stk
	i := len(stk) - 1
	name := stk[i].(string)
	fn := func(ctx *bpl.Context) error {
		v := p.eval(ctx, e.start, e.end)
		ctx.Globals.SetVar(name, v)
		return nil
	}
	stk[i] = bpl.Do(fn)
}

// -----------------------------------------------------------------------------

func (p *Compiler) fnAssert(src interface{}) {

	e := p.popExpr()
	expr := func(ctx *bpl.Context) bool {
		v := p.eval(ctx, e.start, e.end)
		return toBool(v, "assert condition isn't a boolean expression")
	}
	msg := sourceOf(p.ipt, src)
	p.stk = append(p.stk, bpl.Assert(expr, msg))
}

func (p *Compiler) fnFatal(src interface{}) {

	e := p.popExpr()
	r := func(ctx *bpl.Context) (bpl.Ruler, error) {
		val := p.eval(ctx, e.start, e.end)
		if v, ok := val.(string); ok {
			panic("fatal: " + v)
		}
		panic("fatal <expr> must return a string")
	}
	p.stk = append(p.stk, bpl.Dyntype(r))
	//p.stk = append(p.stk, bpl.And(dump(0), bpl.Dyntype(r)))
}

func (p *Compiler) fnDump() {

	p.stk = append(p.stk, bpl.Ruler(dump(0)))
}

// -----------------------------------------------------------------------------

func (p *Compiler) fnRead() {

	e := p.popExpr()
	stk := p.stk
	i := len(stk) - 1
	n := func(ctx *bpl.Context) int {
		v := p.eval(ctx, e.start, e.end)
		return toInt(v, "read bytes isn't an integer expression")
	}
	stk[i] = bpl.Read(n, stk[i].(bpl.Ruler))
}

func (p *Compiler) fnSkip() {

	e := p.popExpr()
	n := func(ctx *bpl.Context) int {
		v := p.eval(ctx, e.start, e.end)
		return toInt(v, "skip bytes isn't an integer expression")
	}
	p.stk = append(p.stk, bpl.Skip(n))
}

// -----------------------------------------------------------------------------

func (p *Compiler) fnReturn() {

	e := p.popExpr()
	fnRet := func(ctx *bpl.Context) (v interface{}, err error) {
		v = p.eval(ctx, e.start, e.end)
		return
	}
	p.stk = append(p.stk, bpl.Return(fnRet))
}

// -----------------------------------------------------------------------------

func (p *Compiler) member(name string) {

	stk := p.stk
	i := len(stk) - 1
	stk[i] = &bpl.Member{Name: name, Type: stk[i].(bpl.Ruler)}
}

func (p *Compiler) gostruct() {

	m := p.popArity()
	rulers := p.popRules(m)
	p.stk = append(p.stk, bpl.Struct(rulers))
}

// -----------------------------------------------------------------------------
