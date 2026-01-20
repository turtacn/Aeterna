package main

import (
	"net"
	"fmt"
)

func main() {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Listening on %s\n", ln.Addr().String())
	ln.Close()
}
