// Package proxy creates a TCP reverse proxy routing traffic from src to dst
package proxy

import (
	"io"
	"log"
	"net"
)

// Proxy forwards TCP traffic from src to dst
type Proxy struct {
	dst     string
	srcConn net.Conn
	dstConn net.Conn
	debug   bool
	done    chan (bool)
	log     *log.Logger
}

func New(dst string, debug bool, log *log.Logger) Proxy {
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

	dstConn, err := net.Dial("tcp", p.dst)
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
