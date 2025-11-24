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

type EditorSharedState struct {
	editor          backend.Editor
	editorStatePubs []chan backend.Editor
}

var editors map[string]EditorSharedState = make(map[string]EditorSharedState)

func editorSubscribe(initArgs InitArgs) (*EditorSharedState, chan backend.Editor) {
	newEditorStatePub := make(chan backend.Editor)

	if editorSharedState, ok := editors[initArgs.FileName]; ok {
		editorSharedState.editorStatePubs = append(editorSharedState.editorStatePubs, newEditorStatePub)

		return &editorSharedState, newEditorStatePub
	} else {
		var editorStatePubs []chan backend.Editor
		editorStatePubs = append(editorStatePubs, newEditorStatePub)

		editors[initArgs.FileName] = EditorSharedState{
			editor: backend.InitializeEditor(
				initArgs.FileName,
				initArgs.ScreenHeight,
				initArgs.ScreenWidth,
			),
			editorStatePubs: editorStatePubs,
		}

		editorSharedState := editors[initArgs.FileName]

		return &editorSharedState, newEditorStatePub
	}
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

	editorSharedState, thisEditorStatePub := editorSubscribe(initArgs)

	editor := editorSharedState.editor
	editorStatePubs := editorSharedState.editorStatePubs

	go func() {
		for {
			editorState := <-thisEditorStatePub
			err = enc.Encode(editorState)
		}
	}()

	for {
		for _, editorStatePub := range editorStatePubs {
			editorStatePub <- editor
		}

		event := EditorEvent{}
		err := dec.Decode(&event)

		if event.IsKey {
			key := event.Key

			if key == tcell.KeyRune {
				fmt.Printf("CLIENT SENT '%c'\n", event.Rune)
			}

			switch editor.Mode {
			case backend.Insert:
				switch key {
				case tcell.KeyEscape:
					editor.ToNormal()
				case tcell.KeyEnter:
					editor.InsertRune('\n')
				case tcell.KeyRight:
					editor.ShiftCursor(0, 1, false, false)
				case tcell.KeyLeft:
					editor.ShiftCursor(0, -1, false, false)
				case tcell.KeyUp:
					editor.ShiftCursor(-1, 0, false, false)
				case tcell.KeyDown:
					editor.ShiftCursor(1, 0, false, false)
				case tcell.KeyRune:
					editor.InsertRune(event.Rune)
				case tcell.KeyBackspace2:
					editor.Backspace()
				}
			case backend.Normal:
				switch key {
				case tcell.KeyRune:
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
				case tcell.KeyRight:
					editor.ShiftCursor(0, 1, false, false)
				case tcell.KeyLeft:
					editor.ShiftCursor(0, -1, false, false)
				case tcell.KeyUp:
					editor.ShiftCursor(-1, 0, false, false)
				case tcell.KeyDown:
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
