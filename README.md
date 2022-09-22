# tcp-proxy

A TCP proxy.

## Install

```
$ go get -v github.com/Mat-Greaves/tcp-proxy/cmd/tcp-proxy
```

## Usage

```
$  tcp-proxy --help
Usage of tcp-proxy:
  -debug
    	debug connections, logging connection details to stdout
  -dst string
    	remove address to forward to (default "localhost:8081")
  -src string
    	local address to listen on (default "localhost:8080")
```

## TODO

- [] add logging for inbound connections
- [] add record errors and log errors on connection closed
