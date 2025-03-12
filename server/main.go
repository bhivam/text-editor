package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"

	"github.com/gdamore/tcell/v2"

	"github.com/bhivam/text-editor/backend"
)

type InitArgs struct {
	ScreenHeight int
	ScreenWidth  int
	FileName     string
}

type EditorEvent struct {
	IsKey  bool
	Key    tcell.Key
	Rune   rune
	Width  int
	Height int
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	enc := json.NewEncoder(conn)
	dec := json.NewDecoder(conn)

	fmt.Println("Handling new connection from", conn.RemoteAddr())

	initArgs := InitArgs{}
	err := dec.Decode(&initArgs)
	if err != nil {
		log.Println("Initialization failed!")
		return
	}

	editor := backend.InitializeEditor(
		initArgs.FileName,
		initArgs.ScreenHeight,
		initArgs.ScreenWidth,
	)

	for {
		err = enc.Encode(editor)

		event := EditorEvent{}
		err := dec.Decode(&event)

		if event.IsKey {
			key := event.Key

      if key == tcell.KeyRune {
        fmt.Printf("CLIENT SENT '%c'\n", event.Rune)
      }

			if editor.Mode == backend.Insert {
				if key == tcell.KeyEscape {
					editor.ToNormal()
				} else if key == tcell.KeyEnter {
					editor.InsertRune('\n')
				} else if key == tcell.KeyRight {
					editor.ShiftCursor(0, 1, false, false)
				} else if key == tcell.KeyLeft {
					editor.ShiftCursor(0, -1, false, false)
				} else if key == tcell.KeyUp {
					editor.ShiftCursor(-1, 0, false, false)
				} else if key == tcell.KeyDown {
					editor.ShiftCursor(1, 0, false, false)
				} else if key == tcell.KeyRune {
					editor.InsertRune(event.Rune)
				} else if key == tcell.KeyBackspace2 {
					editor.Backspace()
				}
			} else if editor.Mode == backend.Normal {
				if key == tcell.KeyRune {
					keyVal := event.Rune
					switch keyVal {
					case rune('q'):
						return

						// switch mode
					case rune('a'):
						editor.ToInsert(true)
					case rune('i'):
						editor.ToInsert(false)

						// basic movement keys
					case rune('j'):
						editor.ShiftCursor(1, 0, false, false)
					case rune('k'):
						editor.ShiftCursor(-1, 0, false, false)
					case rune('h'):
						editor.ShiftCursor(0, -1, false, false)
					case rune('l'):
						editor.ShiftCursor(0, 1, false, false)
					}
				} else if key == tcell.KeyRight {
					editor.ShiftCursor(0, 1, false, false)
				} else if key == tcell.KeyLeft {
					editor.ShiftCursor(0, -1, false, false)
				} else if key == tcell.KeyUp {
					editor.ShiftCursor(-1, 0, false, false)
				} else if key == tcell.KeyDown {
					editor.ShiftCursor(1, 0, false, false)
				}
			}
		} else {
			editor.ScreenWidth, editor.ScreenHeight = event.Width, event.Height
		}

		if err != nil {
			log.Printf("Error reading from connection: %v", err)
			return
		}
	}
}

func main() {
	ln, err := net.Listen("tcp", ":8081")
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
