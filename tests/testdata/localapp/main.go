// Command localapp is a test helper. It listens on the port given as its first
// argument and blocks, so the ports CLI sees a binary built inside a project
// directory that serves a port. The name says nothing about a language, which is
// what a compiled app such as a Roc or Go build looks like.
package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: localapp <port>")
		os.Exit(2)
	}

	listener, err := net.Listen("tcp", "127.0.0.1:"+os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to listen: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = listener.Close() }()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
}
