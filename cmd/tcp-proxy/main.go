package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	proxy "github.com/Mat-Greaves/tcp-proxy"
)

var src *string = flag.String("src", "localhost:8080", "local address to listen on")
var dst *string = flag.String("dst", "localhost:8081", "remove address to forward to")
var debug *bool = flag.Bool("debug", false, "debug connections, logging connection details to stdout")

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
			p := proxy.New(*dst, *debug, log.Default())
			p.ServeTCP(conn)
		}()
	}
}
