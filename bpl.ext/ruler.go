package bpl

import (
	"fmt"

	"github.com/qiniu/bpl"
	"github.com/xushiwei/qlang/exec"
)

// -----------------------------------------------------------------------------

func clone(rs []interface{}) []bpl.Ruler {

	dest := make([]bpl.Ruler, len(rs))
	for i, v := range rs {
		dest[i] = v.(bpl.Ruler)
	}
	return dest
}

func cloneNames(names []interface{}) []string {

	dest := make([]string, len(names))
	for i, v := range names {
		dest[i] = v.(string)
	}
	return dest
}

func (p *Compiler) and(m int) {

	if m == 1 {
		return
	}
	stk := p.stk
	n := len(stk)
	stk[n-m] = bpl.And(clone(stk[n-m:])...)
	p.stk = stk[:n-m+1]
}

func (p *Compiler) seq(m int) {

	stk := p.stk
	n := len(stk)
	stk[n-m] = bpl.Seq(clone(stk[n-m:])...)
	p.stk = stk[:n-m+1]
}

func (p *Compiler) variable(name string) {

	p.stk = append(p.stk, name)
}

func (p *Compiler) ruleOf(name string) (r bpl.Ruler, ok bool) {

	r, ok = p.rulers[name]
	if !ok {
		if r, ok = p.vars[name]; !ok {
			if r, ok = builtins[name]; ok {
				p.rulers[name] = r
			}
		}
	}
	return
}

func (p *Compiler) sizeof(name string) {

	r, ok := p.ruleOf(name)
	if !ok {
		panic(fmt.Errorf("sizeof error: type `%v` not found", name))
	}
	n := r.SizeOf()
	if n < 0 {
		panic(fmt.Errorf("sizeof error: type `%v` isn't a fixed size type", name))
	}
	p.code.Block(exec.Push(n))
}

func (p *Compiler) ident(name string) {

	r, ok := p.ruleOf(name)
	if !ok {
		v := &bpl.TypeVar{Name: name}
		p.vars[name] = v
		r = v
	}
	p.stk = append(p.stk, r)
}

func (p *Compiler) assign(name string) {

	a := p.stk[0].(bpl.Ruler)
	if v, ok := p.vars[name]; ok {
		if err := v.Assign(a); err != nil {
			panic(err)
		}
	} else if _, ok := p.rulers[name]; ok {
		panic("ruler already exists: " + name)
	} else {
		p.rulers[name] = a
	}
	p.stk = p.stk[:0]
}

func (p *Compiler) repeat0() {

	stk := p.stk
	i := len(stk) - 1
	stk[i] = bpl.Repeat0(stk[i].(bpl.Ruler))
}

func (p *Compiler) repeat1() {

	stk := p.stk
	i := len(stk) - 1
	stk[i] = bpl.Repeat1(stk[i].(bpl.Ruler))
}

func (p *Compiler) repeat01() {

	stk := p.stk
	i := len(stk) - 1
	stk[i] = bpl.Repeat01(stk[i].(bpl.Ruler))
}

func (p *Compiler) xline(src interface{}) {

	f := p.ipt.FileLine(src)
	stk := p.stk
	i := len(stk) - 1
	stk[i] = bpl.FileLine(f.File, f.Line, stk[i].(bpl.Ruler))
}

// -----------------------------------------------------------------------------
