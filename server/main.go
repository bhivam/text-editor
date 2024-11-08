package main

import (
	"fmt"
	"io"
	"log"
	"net"
)

func handleConnection(conn net.Conn) {
	fmt.Println("Handling new connection from", conn.RemoteAddr())
	defer conn.Close()
	buf := [512]byte{}
	for {
		fmt.Fprintln(conn, "Hello from the server!")
		_, err := conn.Read(buf[:])
		if err == io.EOF {
			log.Printf("Connection closed by remote host: %v", conn.RemoteAddr())
			return
		}

		fmt.Printf("Received %v bytes from %v: %v\n", len(buf), conn.RemoteAddr(), string(buf[:]))

		if err != nil {
			log.Printf("Error reading from connection: %v", err)
			return
		}
	}
}

func main() {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		// handle error
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
		}
		go handleConnection(conn)
	}
}
