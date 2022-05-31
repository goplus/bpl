package replay

import (
	"io"
	"io/ioutil"
	"net"

	"github.com/goplus/bpl/hex"
)

// -----------------------------------------------------------------------------

// Request replays sequent requests.
//
func Request(host string, w io.Writer, body io.Reader) (err error) {

	c1, err := net.Dial("tcp", host)
	if err != nil {
		return
	}
	c := c1.(*net.TCPConn)

	go func() {
		if w == nil {
			w = ioutil.Discard
		}
		io.Copy(w, c)
		c.CloseRead()
	}()

	_, err = io.Copy(c, body)
	c.CloseWrite()
	return
}

// HexRequest replays sequent requests from a hexdump file.
//
func HexRequest(host string, w io.Writer, in io.Reader, filter string) (err error) {

	f, err := hex.Reader(in, filter)
	if err != nil {
		return
	}
	defer f.Close()

	return Request(host, w, f)
}

// -----------------------------------------------------------------------------
