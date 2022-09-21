// Package tcp-proxy creates a TCP reverse proxy routing traffic from src to dst
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
)

var src *string = flag.String("src", "localhost:8080", "local address to listen on")
var dst *string = flag.String("dst", "localhost:8081", "remove address to forward to")
var debug *bool = flag.Bool("debug", false, "debug connections, logging connection details to stdout")

// Proxy forwards TCP traffic from src to dst
type Proxy struct {
	dst     string
	srcConn net.Conn
	dstConn net.Conn
	debug   bool
	done    chan (bool)
	log     *log.Logger
}

// TODO: move main to cmd directory
func main() {
	// listen for new inbound connections
	// handle each each connection in a goroutine
	// need to consider that traffic comes in both directions
	flag.Parse()
	fmt.Printf("Forwarding TCP connections from: %s, to: %s\n", *src, *dst)
	listener, err := net.Listen("tcp", *src)
	if err != nil {
		log.Fatalf("failed to listen %s", err)
	}
	for {
		// received new inbound connection
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("error accepting connection %s\n", err)
			continue
		}

		go func() {
			p := NewProxy(*dst, *debug, log.Default())
			p.ServeTCP(conn)
		}()
	}
}

func NewProxy(dst string, debug bool, log *log.Logger) Proxy {
	done := make(chan bool, 2)
	return Proxy{
		dst:   dst,
		debug: debug,
		done:  done,
		log:   log,
	}
}

// ServceTCP proxies a single connection copying data in bi-directionally
func (p *Proxy) ServeTCP(srcConn net.Conn) {
	p.srcConn = srcConn
	defer p.srcConn.Close()

	dstConn, err := net.Dial("tcp", *dst)
	if err != nil {
		log.Printf("failed to connect to dst %s\n", err)
		return
	}
	p.dstConn = dstConn
	defer dstConn.Close()

	go p.pipe(p.srcConn, p.dstConn)
	go p.pipe(p.dstConn, p.srcConn)
	<-p.done
}

// pipe copies from src to dst, decrements Proxies wg when an error
// is received to indicate connection has been closed. If debug is true
// pipe will log all data copied
func (p *Proxy) pipe(src net.Conn, dst net.Conn) {
	logger := io.Discard
	if p.debug {
		prefix := "---->"
		if src.LocalAddr() == p.dstConn.LocalAddr() {
			prefix = "<----"
		}
		logger = connLogger{
			prefix: prefix,
			logger: p.log,
		}
	}
	io.Copy(logger, io.TeeReader(src, dst))
	p.done <- true
}

type connLogger struct {
	prefix string
	logger *log.Logger
}

func (cl connLogger) Write(p []byte) (n int, err error) {
	cl.logger.Printf("%s %s", cl.prefix, string(p))
	return len(p), nil
}
