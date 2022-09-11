// Package tcp-proxy creates a TCP reverse proxy routing traffic from src to dst
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
)

var src *string = flag.String("src", "localhost:8080", "local address to listen on")
var dst *string = flag.String("dst", "localhost:8081", "remove address to forward to")

// Proxy forwards TCP traffic from src to dst
type Proxy struct {
	src string
	dst string
	log log.Logger
}

func main() {
	// listen for new inbound connections
	// handle each each connection in a goroutine
	// need to consider that traffic comes in both directions
	flag.Parse()
	fmt.Printf("Forwarding TCP connections from: %s, to: %s\n", *src, *dst)
	// TODO: listen until ctrl+c is called, think about graceful shutdown
	p := Proxy{
		src: *src,
		dst: *dst,
		log: *log.Default(),
	}
	p.ServeTCP()
}

// todo add shutdown mechanism
// serve listens for inbound TCP connections, creating a goroutine for each to forward connections
func (p *Proxy) ServeTCP() error {
	listener, err := net.Listen("tcp", p.src)
	if err != nil {
		return fmt.Errorf("failed to listen %w", err)
	}
	for {
		// received new inbound connection
		srcConn, err := listener.Accept()
		if err != nil {
			p.log.Printf("error accepting connection %s\n", err)
			continue
		}

		go func() {
			defer srcConn.Close()
			dstConn, err := net.Dial("tcp", p.dst)
			if err != nil {
				p.log.Printf("failed to connect to dst %s\n", err)
				return
			}
			defer dstConn.Close()
			wg := sync.WaitGroup{}
			// if either errors close the proxe
			wg.Add(1)
			copyWatcher(&wg, srcConn, dstConn)
			copyWatcher(&wg, dstConn, srcConn)
			wg.Wait()
		}()
	}
}

// copyWatcher performs io.Copy() and decrements wg when it completes or returns an error
// todo: logging should be optional
// todo: each proxy connection should have a unique identifier
func copyWatcher(wg *sync.WaitGroup, src io.Reader, dst io.Writer) {
	io.Copy(os.Stdout, io.TeeReader(src, dst))
	wg.Done()
}
