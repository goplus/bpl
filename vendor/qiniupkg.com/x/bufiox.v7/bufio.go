package bufiox

import (
	"bufio"
	"bytes"
	"io"
	"unsafe"
)

// -----------------------------------------------------------------------------

type nilReaderImpl int

func (p nilReaderImpl) Read(b []byte) (n int, err error) {
	return 0, io.EOF
}

var nilReader io.Reader = nilReaderImpl(0)

// -----------------------------------------------------------------------------

type reader struct {
	buf          []byte
	rd           io.Reader // reader provided by the client
	r, w         int       // buf read and write positions
	err          error
	lastByte     int
	lastRuneSize int
}

// NewReaderBuffer returns a new Reader who uses a extern buffer.
func NewReaderBuffer(buf []byte) *bufio.Reader {
	r := &reader{
		buf:          buf,
		rd:           nilReader,
		w:            len(buf),
		lastByte:     -1,
		lastRuneSize: -1,
	}
	b := new(bufio.Reader)
	*b = *(*bufio.Reader)(unsafe.Pointer(r))
	return b
}

// Buffer is reserved for internal use.
func Buffer(b *bufio.Reader) []byte {
	r := (*reader)(unsafe.Pointer(b))
	return r.buf
}

// IsReaderBuffer returns if this Reader instance is returned by NewReaderBuffer
func IsReaderBuffer(b *bufio.Reader) bool {
	r := (*reader)(unsafe.Pointer(b))
	return r.rd == nilReader
}

// ReadAll reads all data
func ReadAll(b *bufio.Reader) (ret []byte, err error) {
	r := (*reader)(unsafe.Pointer(b))
	if r.rd == nilReader {
		ret, r.buf = r.buf, nil
		return
	}
	var w bytes.Buffer
	_, err1 := b.WriteTo(&w)
	return w.Bytes(), err1
}

// -----------------------------------------------------------------------------
