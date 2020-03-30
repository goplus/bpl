package mockd

import (
	"bufio"
	"io"
	"io/ioutil"
	"net"
	"os"

	"github.com/qiniu/bpl/hex"
	"github.com/qiniu/x/log"
)

// -----------------------------------------------------------------------------

// A Server is a mockd server.
//
type Server struct {
	Addr     string
	Listened chan bool
	Data     io.ReaderAt
	Fsize    int64
}

// ListenAndServe listens an address and serves requests.
//
func (p *Server) ListenAndServe() (err error) {

	addr := p.Addr
	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("ListenAndServe(tcp proxy) %s failed: %v\n", addr, err)
		return
	}
	if p.Listened != nil {
		p.Listened <- true
	}
	err = p.Serve(l)
	if err != nil {
		log.Fatalf("ListenAndServe(mockd) %s failed: %v\n", addr, err)
	}
	return
}

// Serve serves requests.
//
func (p *Server) Serve(l net.Listener) (err error) {

	defer l.Close()

	for {
		c1, err1 := l.Accept()
		if err1 != nil {
			return err1
		}
		c := c1.(*net.TCPConn)
		go p.handle(c)
	}
}

func (p *Server) handle(c *net.TCPConn) {

	go func() {
		io.Copy(ioutil.Discard, c)
		c.CloseRead()
	}()

	r := io.NewSectionReader(p.Data, 0, p.Fsize)
	in := bufio.NewReader(r)
	hex.Undump(c, in, "[RESP]")
	c.CloseWrite()
}

// -----------------------------------------------------------------------------

// ListenAndServe listens an address and serves requests.
//
func ListenAndServe(addr string, file string) (err error) {

	f, err := os.Open(file)
	if err != nil {
		log.Fatalln("Open failed:", err)
		return
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		log.Fatalln("Stat failed:", err)
		return
	}

	server := &Server{
		Addr:  addr,
		Data:  f,
		Fsize: fi.Size(),
	}
	return server.ListenAndServe()
}

// -----------------------------------------------------------------------------
