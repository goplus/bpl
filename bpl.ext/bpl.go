package bpl

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"sort"
	"strings"

	"github.com/qiniu/bpl"
	"qlang.io/qlang.spec.v1"

	"qiniupkg.com/text/tpl.v1/interpreter"
	"qiniupkg.com/x/bufiox.v7"
	"qiniupkg.com/x/log.v7"
)

// -----------------------------------------------------------------------------

type filterWriter bytes.Buffer

func (p *filterWriter) Write(b []byte) (n int, err error) {

	p1 := (*bytes.Buffer)(p)
	n, err = p1.Write(b)
	b = p1.Bytes()
	b = b[len(b)-n:]
	for i, c := range b {
		if c == '`' {
			b[i] = '.'
		}
	}
	return
}

// -----------------------------------------------------------------------------

var (
	// Ldefault is default flag for `Dumper`.
	Ldefault = log.Llevel

	// Llong is long log mode for `Dumper`.
	Llong = log.Llevel | log.LstdFlags

	// Dumper is used for dumping log informations.
	Dumper = log.New(os.Stdout, "", Ldefault)

	// SetCaseType controls to set `_type` into matching result or not.
	SetCaseType = true
)

// SetDumper sets the dumper instance for dumping log informations.
//
func SetDumper(w io.Writer, flags ...int) {

	flag := Ldefault
	if len(flags) > 0 {
		flag = flags[0]
	}
	Dumper = log.New(w, "", flag)
}

func writePrefix(b *bytes.Buffer, lvl int) {

	for i := 0; i < lvl; i++ {
		b.WriteString("  ")
	}
}

// DumpDom dumps a dom tree.
//
func DumpDom(b *bytes.Buffer, dom interface{}, lvl int) {

	dumpDomValue(b, reflect.ValueOf(dom), lvl)
}

type stringSlice []reflect.Value

func (p stringSlice) Len() int           { return len(p) }
func (p stringSlice) Less(i, j int) bool { return p[i].String() < p[j].String() }
func (p stringSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

var typeBytes = reflect.TypeOf([]byte(nil))

func dumpDomValue(b *bytes.Buffer, dom reflect.Value, lvl int) {

retry:
	switch dom.Kind() {
	case reflect.Slice:
		if dom.Type() == typeBytes {
			b.WriteByte('\n')
			d := hex.Dumper((*filterWriter)(b))
			d.Write(dom.Bytes())
			d.Close()
			return
		}
		b.WriteByte('[')
		n := dom.Len()
		for i := 0; i < n; i++ {
			b.WriteByte('\n')
			writePrefix(b, lvl+1)
			dumpDomValue(b, dom.Index(i), lvl+1)
			b.WriteByte(',')
		}
		b.WriteByte('\n')
		writePrefix(b, lvl)
		b.WriteByte(']')
	case reflect.Map:
		b.WriteByte('{')
		keys := dom.MapKeys()
		fstring := dom.Type().Key().Kind() == reflect.String
		if fstring {
			n := 0
			for _, key := range keys {
				if strings.HasPrefix(key.String(), "_") {
					continue
				}
				keys[n] = key
				n++
			}
			keys = keys[:n]
			sort.Sort(stringSlice(keys))
		}
		for _, key := range keys {
			item := dom.MapIndex(key)
			b.WriteByte('\n')
			writePrefix(b, lvl+1)
			if fstring {
				b.WriteString(key.String())
			} else {
				dumpDomValue(b, key, lvl+1)
			}
			b.WriteString(": ")
			dumpDomValue(b, item, lvl+1)
		}
		b.WriteByte('\n')
		writePrefix(b, lvl)
		b.WriteByte('}')
	case reflect.Interface, reflect.Ptr:
		if dom.IsNil() {
			b.WriteString("<nil>")
			return
		}
		dom = dom.Elem()
		goto retry
	default:
		ret, _ := json.Marshal(dom.Interface())
		b.Write(ret)
	}
}

// -----------------------------------------------------------------------------

type dump int

func (p dump) Match(in *bufio.Reader, ctx *bpl.Context) (v interface{}, err error) {

	dom := ctx.Dom()
	if dom == qlang.Undefined {
		return
	}

	var b bytes.Buffer
	if prefix, ok := ctx.Globals.Var("BPL_DUMP_PREFIX"); ok {
		b.WriteString(prefix.(string))
	}
	b.WriteByte('\n')
	DumpDom(&b, dom, 0)
	Dumper.Info(b.String())
	return
}

func (p dump) RetType() reflect.Type {

	return bpl.TyInterface
}

func (p dump) SizeOf() int {

	return 0
}

// -----------------------------------------------------------------------------

// A Ruler is a matching unit.
//
type Ruler struct {
	Impl bpl.Ruler
}

// Match matches input stream `in`, and returns matching result.
//
func (p Ruler) Match(in *bufio.Reader, ctx *bpl.Context) (v interface{}, err error) {

	return bpl.MatchStream(p.Impl, in, ctx)
}

// SafeMatch matches input stream `in`, and returns matching result.
//
func (p Ruler) SafeMatch(in *bufio.Reader, ctx *bpl.Context) (v interface{}, err error) {

	defer func() {
		if e := recover(); e != nil {
			switch val := e.(type) {
			case string:
				err = errors.New(val)
			case error:
				err = val
			case int:
				v = val
			default:
				panic(e)
			}
		}
	}()

	return bpl.MatchStream(p.Impl, in, ctx)
}

// MatchStream matches input stream `r`, and returns matching result.
//
func (p Ruler) MatchStream(r io.Reader) (v interface{}, err error) {

	in := bufio.NewReader(r)
	ctx := bpl.NewContext()
	return p.SafeMatch(in, ctx)
}

// MatchBuffer matches input buffer `b`, and returns matching result.
//
func (p Ruler) MatchBuffer(b []byte) (v interface{}, err error) {

	in := bufiox.NewReaderBuffer(b)
	ctx := bpl.NewContext()
	return p.SafeMatch(in, ctx)
}

// -----------------------------------------------------------------------------

// New compiles bpl source code and returns the corresponding matching unit.
//
func New(code []byte, fname string) (r Ruler, err error) {

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

	p := newCompiler()
	engine, err := interpreter.New(p, interpreter.InsertSemis)
	if err != nil {
		return
	}

	p.ipt = engine
	err = engine.MatchExactly(code, fname)
	if err != nil {
		return
	}

	if DumpCode != 0 {
		p.code.Dump()
	}
	return p.Ret()
}

// NewFromString compiles bpl source code and returns the corresponding matching unit.
//
func NewFromString(code string, fname string) (r Ruler, err error) {

	return New([]byte(code), fname)
}

// NewFromFile compiles bpl source file and returns the corresponding matching unit.
//
func NewFromFile(fname string) (r Ruler, err error) {

	b, err := ioutil.ReadFile(fname)
	if err != nil {
		return
	}
	return New(b, fname)
}

// NewContext returns a new matching Context.
//
func NewContext() *bpl.Context {

	return bpl.NewContext()
}

// -----------------------------------------------------------------------------

// SetDumpCode sets dump code mode:
//	"1" - dump code with rem instruction.
//	"2" - dump code without rem instruction.
//  else - don't dump code.
//
func SetDumpCode(dumpCode string) {

	switch dumpCode {
	case "true", "1":
		DumpCode = 1
	case "2":
		DumpCode = 2
	default:
		DumpCode = 0
	}
}

// -----------------------------------------------------------------------------
