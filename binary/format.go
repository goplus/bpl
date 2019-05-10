package binary

import (
	"bufio"
	"bytes"
	"fmt"
	"reflect"
	"unsafe"
)

// -----------------------------------------------------------------------------

func writeCString(w *bufio.Writer, v string) (err error) {

	_, err = w.WriteString(v)
	if err != nil {
		return
	}
	return w.WriteByte(0)
}

func writeStruct(w *bufio.Writer, v reflect.Value) (err error) {

	n := v.NumField()
	for i := 0; i < n; i++ {
		fv := v.Field(i)
		err = WriteValue(w, fv)
		if err != nil {
			return
		}
	}
	return
}

// WriteValue serializes data into a writer.
//
func WriteValue(w *bufio.Writer, v reflect.Value) (err error) {

	var val uint64

retry:
	kind := v.Kind()
	switch {
	case kind == reflect.Struct:
		return writeStruct(w, v)
	case kind >= reflect.Int8 && kind <= reflect.Int64:
		val = uint64(v.Int())
		kind -= reflect.Int8
	case kind >= reflect.Uint8 && kind <= reflect.Uint64:
		val = v.Uint()
		kind -= reflect.Uint8
	case kind == reflect.String:
		return writeCString(w, v.String())
	case kind == reflect.Float64:
		*(*float64)(unsafe.Pointer(&val)) = v.Float()
		kind = 3
	case kind == reflect.Float32:
		*(*float32)(unsafe.Pointer(&val)) = float32(v.Float())
		kind = 2
	case kind == reflect.Ptr:
		v = v.Elem()
		goto retry
	default:
		return fmt.Errorf("bpl/binary.Write - unsupported type: %v", v.Type())
	}
	b := (*[8]byte)(unsafe.Pointer(&val))
	_, err = w.Write(b[:1<<kind])
	return
}

// Write serializes data into a writer.
//
func Write(w *bufio.Writer, v interface{}) (err error) {

	return WriteValue(w, reflect.ValueOf(v))
}

// Marshal returns serialization result of a value.
//
func Marshal(v interface{}) (b []byte, err error) {

	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	err = Write(w, v)
	if err != nil {
		return
	}
	w.Flush()
	return buf.Bytes(), nil
}

// -----------------------------------------------------------------------------
