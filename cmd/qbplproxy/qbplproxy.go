package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"path"
	"reflect"
	"strings"

	bpl "github.com/qiniu/bpl/bpl.ext"
	qlang "github.com/xushiwei/qlang/spec"

	"github.com/qiniu/x/log"
)

// -----------------------------------------------------------------------------

// A Env is the environment of a callback.
//
type Env struct {
	Src       *net.TCPConn
	Dest      *net.TCPConn
	Direction string
	Conn      string
}

// A ReverseProxier is a reverse proxier server.
//
type ReverseProxier struct {
	Addr       string
	Backend    string
	OnResponse func(io.Reader, *Env) (err error)
	OnRequest  func(io.Reader, *Env) (err error)
	Listened   chan bool
}

// ListenAndServe listens on `Addr` and serves to proxy requests to `Backend`.
//
func (p *ReverseProxier) ListenAndServe() (err error) {

	addr := p.Addr
	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("ListenAndServe(qbplproxy) %s failed: %v\n", addr, err)
		return
	}
	if p.Listened != nil {
		p.Listened <- true
	}
	err = p.Serve(l)
	if err != nil {
		log.Fatalf("ListenAndServe(qbplproxy) %s failed: %v\n", addr, err)
	}
	return
}

func onNil(r io.Reader, env *Env) (err error) {

	_, err = io.Copy(ioutil.Discard, r)
	return
}

// Serve serves to proxy requests to `Backend`.
//
func (p *ReverseProxier) Serve(l net.Listener) (err error) {

	defer l.Close()

	backend, err := net.ResolveTCPAddr("tcp", p.Backend)
	if err != nil {
		return
	}

	onResponse := p.OnResponse
	if onResponse == nil {
		onResponse = onNil
	}

	onRequest := p.OnRequest
	if onRequest == nil {
		onRequest = onNil
	}

	for {
		c1, err1 := l.Accept()
		if err1 != nil {
			return err1
		}
		c := c1.(*net.TCPConn)
		go func() {
			c2, err2 := net.DialTCP("tcp", nil, backend)
			if err2 != nil {
				log.Error("qbplproxy: dial backend failed -", p.Backend, "error:", err2)
				c.Close()
				return
			}

			conn := c.RemoteAddr().String()
			go func() {
				r2 := io.TeeReader(c2, c)
				onResponse(r2, &Env{Src: c2, Dest: c, Direction: "RESP", Conn: conn})
				c.CloseWrite()
				c2.CloseRead()
			}()

			r := io.TeeReader(c, c2)
			err2 = onRequest(r, &Env{Src: c, Dest: c2, Direction: "REQ", Conn: conn})
			if err2 != nil {
				log.Info("qbplproxy (request):", err2, "type:", reflect.TypeOf(err2))
			}
			c.CloseRead()
			c2.CloseWrite()
		}()
	}
}

// -----------------------------------------------------------------------------

var (
	host     = flag.String("h", "", "listen host (listenIp:port).")
	backend  = flag.String("b", "", "backend host (backendIp:port).")
	filter   = flag.String("f", "", "filter condition. eg. -f 'flashVer=LNX 9,0,124,2' or -f 'reqMode=play' or -f 'dir=REQ|RESP'")
	protocol = flag.String("p", "", "protocol file in BPL syntax, default is guessed by <port>.")
	output   = flag.String("o", "", "output log file, default is stderr.")
	logmode  = flag.String("l", "", "log mode: short (default) or long.")
)

var (
	baseDir string // $HOME/.qbpl/formats/
)

func fileExists(file string) bool {

	_, err := os.Stat(file)
	return err == nil
}

func guessProtocol(host string) string {

	index := strings.Index(host, ":")
	if index >= 0 {
		proto := baseDir + host[index+1:] + ".bpl"
		if fileExists(proto) {
			return proto
		}
	}
	return ""
}

// qbplproxy -h <listenIp:port> -b <backendIp:port> [-p <protocol>.bpl -f <filter> -o <output>.log -l <logmode>]
//
func main() {

	flag.Parse()
	if *host == "" || *backend == "" {
		fmt.Fprintln(
			os.Stderr,
			"Usage: qbplproxy -h <listenIp:port> -b <backendIp:port> [-p <protocol>.bpl -f <filter> -o <output>.log -l <logmode>]")
		flag.PrintDefaults()
		return
	}
	bpl.SetDumpCode(os.Getenv("BPL_DUMPCODE"))
	qlang.DumpStack = true

	baseDir = os.Getenv("HOME") + "/.qbpl/formats/"
	if *protocol == "" {
		*protocol = guessProtocol(*host)
		if *protocol == "" {
			*protocol = guessProtocol(*backend)
			if *protocol == "" {
				log.Fatalln("I can't know protocol by listening port, please use -p <protocol>.")
			}
		}
	} else {
		if path.Ext(*protocol) == "" {
			*protocol = baseDir + *protocol + ".bpl"
		}
	}

	filterCond := make(map[string]interface{})
	if *filter != "" {
		m, err := url.ParseQuery(*filter)
		if err != nil {
			log.Fatalln("Error: invalid -f <filter> argument -", err)
		}
		for k, v := range m {
			filterCond[k] = v[0]
		}
	}

	logflags := bpl.Ldefault
	flong := (*logmode == "long")
	if flong {
		logflags = bpl.Llong
	}

	onBpl := onNil
	if *protocol != "nil" {
		if *output != "" {
			f, err := os.Create(*output)
			if err != nil {
				log.Fatalln("Create log file failed:", err)
			}
			defer f.Close()
			bpl.SetDumper(f, logflags)
		}
		ruler, err := bpl.NewFromFile(*protocol)
		if err != nil {
			log.Fatalln("bpl.NewFromFile failed:", err)
		}
		onBpl = func(r io.Reader, env *Env) (err error) {
			in := bufio.NewReader(r)
			ctx := bpl.NewContext()
			ctx.Globals.SetVar("BPL_FILTER", filterCond)
			ctx.Globals.SetVar("BPL_DIRECTION", env.Direction)
			if flong {
				ctx.Globals.SetVar("BPL_DUMP_PREFIX", "[CONN:"+env.Conn+"]["+env.Direction+"]")
			} else {
				ctx.Globals.SetVar("BPL_DUMP_PREFIX", "["+env.Direction+"]")
			}
			_, err = ruler.SafeMatch(in, ctx)
			if err != nil {
				log.Error("Match failed:", err)
			}
			in.WriteTo(ioutil.Discard)
			return
		}
	}
	log.Std = bpl.Dumper

	rp := &ReverseProxier{
		Addr:       *host,
		Backend:    *backend,
		OnRequest:  onBpl,
		OnResponse: onBpl,
	}
	rp.ListenAndServe()
}

// -----------------------------------------------------------------------------
