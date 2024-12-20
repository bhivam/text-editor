package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/bhivam/text-editor/backend"
)

func handleConnection(conn net.Conn) {
	fmt.Println("Handling new connection from", conn.RemoteAddr())

	/*
	   1. Read file name, screen height/width from client. delimited by '|' and ends with ';'
	   2. Read file or create file.
	   3. Initialize data structures accordingly
	   4. Serialize and send over obejcts to client
	   5. recieve key, update, serialize, send loop
	*/

	// read file name from client
	buf := [512]byte{}
	editor_arg_str := ""
  editor_arg_str_len := 0
	for {
		_, err := conn.Read(buf[:])
		if err == io.EOF {
			log.Printf("Connection closed by remote host: %v", conn.RemoteAddr())
			return
		}

		if err != nil {
			log.Printf("Error reading from connection: %v", err)
			return
		}

		fmt.Printf("Received %v bytes from %v: %v\n", len(buf), conn.RemoteAddr(), string(buf[:]))

		for i, b := range buf {
			if b == ';' {
        fmt.Println("Found ;")
				editor_arg_str += string(buf[:i])
        editor_arg_str_len += i
				break
			}
		}

    if buf[editor_arg_str_len] == ';' {
      break
    }

		editor_arg_str += string(buf[:])
    editor_arg_str_len += len(buf)
	}
  
	editor_args := strings.Split(editor_arg_str, "|")
	file_name := editor_args[0]
	screen_height, err := strconv.Atoi(editor_args[1])
	if err != nil {
		log.Fatalf("Error parsing screen height: %v", err)
	}
	screen_width, err := strconv.Atoi(editor_args[2])
	if err != nil {
		log.Fatalf("Error parsing screen width: %v", err)
	}

	fmt.Println("File name:", file_name)
  fmt.Println("Screen height:", screen_height)
  fmt.Println("Screen width:", screen_width)

	// initialize data structures accordingly
  editor := backend.InitializeEditor(file_name, screen_height, screen_width)

	// serialize and send over objects to client
  json, err := json.Marshal(editor)
  if err != nil {
    log.Fatalf("Error marshalling editor: %v", err)
  }

  fmt.Println(string(json))
  conn.Close()
  return
	// recieve key, update, serialize, send loop

	defer conn.Close()
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
