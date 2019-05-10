package bson

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"io"
	"reflect"

	"gopkg.in/mgo.v2/bson"
	"github.com/qiniu/bpl"
)

// -----------------------------------------------------------------------------

func peekInt32(in *bufio.Reader) (v int32, err error) {

	t, err := in.Peek(4)
	if err != nil {
		return
	}
	v = int32(binary.LittleEndian.Uint32(t))
	return
}

// -----------------------------------------------------------------------------

// A Document represents a bson document.
//
type Document struct {
	data  []byte
	cache map[string]interface{}
}

// MarshalJSON is required by json.Marshal.
//
func (p *Document) MarshalJSON() (b []byte, err error) {

	vars, err := p.Vars()
	if err != nil {
		return
	}
	return json.Marshal(vars)
}

// Vars returns all variables of this document.
//
func (p *Document) Vars() (vars map[string]interface{}, err error) {

	if p.cache == nil {
		err = bson.Unmarshal(p.data, &vars)
		if err != nil {
			return
		}
		p.cache = vars
		return
	}
	return p.cache, nil
}

// -----------------------------------------------------------------------------

type typeImpl int

var tyDocument = reflect.TypeOf((*Document)(nil))

func (p typeImpl) Match(in *bufio.Reader, ctx *bpl.Context) (v interface{}, err error) {

	n, err := peekInt32(in)
	if err != nil {
		return
	}
	data := make([]byte, n)
	_, err = io.ReadFull(in, data)
	if err != nil {
		return
	}
	return &Document{data: data}, nil
}

func (p typeImpl) RetType() reflect.Type {

	return tyDocument
}

func (p typeImpl) SizeOf() int {

	return -1
}

// Type is a matching unit that matches a bson document.
//
var Type bpl.Ruler = typeImpl(0)

// -----------------------------------------------------------------------------
