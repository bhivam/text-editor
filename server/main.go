package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"path/filepath"

	"github.com/gdamore/tcell/v2"

	"github.com/bhivam/text-editor/backend"
)

type InitArgs struct {
	ScreenHeight int
	ScreenWidth  int
	FilePath     string
}

type EditorEvent struct {
	IsKey  bool
	Key    tcell.Key
	Rune   rune
	Width  int
	Height int
}

type IndividualEditorState struct {
	editorStatePub chan backend.Editor
	editor         backend.Editor
}

type FileEditSession struct {
	content      *backend.Content
	editorStates []IndividualEditorState
}

var fileEditSessions map[string]*FileEditSession = make(map[string]*FileEditSession)

func editorSubscribe(initArgs InitArgs) (*FileEditSession, IndividualEditorState) {
	newEditorStatePub := make(chan backend.Editor)

	if fileEditSession, ok := fileEditSessions[initArgs.FilePath]; ok {
		individualEditorState := IndividualEditorState{
			editorStatePub: newEditorStatePub,
			editor: backend.Editor{
				Content:      fileEditSession.content,
				Cursor:       &backend.Cursor{Index: 0, Row: 0, Col: 0},
				FilePath:     initArgs.FilePath,
				FileName:     filepath.Base(initArgs.FilePath),
				ScreenHeight: initArgs.ScreenHeight,
				ScreenWidth:  initArgs.ScreenWidth,
				Mode:         backend.Normal,
			},
		}

		fileEditSession.editorStates = append(fileEditSession.editorStates, individualEditorState)

		return fileEditSession, individualEditorState
	} else {
		var editorStates []IndividualEditorState

		log.Printf("Reading from %s %d", initArgs.FilePath, len(initArgs.FilePath))

		editor := backend.InitializeEditor(
			initArgs.FilePath,
			initArgs.ScreenHeight,
			initArgs.ScreenWidth,
		)

		individualEditorState := IndividualEditorState{
			editorStatePub: newEditorStatePub,
			editor:         editor,
		}

		editorStates = append(editorStates, individualEditorState)

		fileEditSessions[initArgs.FilePath] = &FileEditSession{
			content:      editor.Content,
			editorStates: editorStates,
		}

		fileEditSession := fileEditSessions[initArgs.FilePath]

		return fileEditSession, individualEditorState
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

	editorSharedState, thisClientEditorState := editorSubscribe(initArgs)

	editor := thisClientEditorState.editor

	go func() {
		for {
			newEditor, ok := <-thisClientEditorState.editorStatePub
			if !ok {
				return
			}

			editor.Content = newEditor.Content

			if err := enc.Encode(editor); err != nil {
				log.Printf("Error encoding to connection in goroutine: %v", err)
				return
			}
		}
	}()

	for {

		event := EditorEvent{}
		err := dec.Decode(&event)
		if err != nil {
			log.Printf("Error reading from connection: %v", err)
			return
		}

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

		for _, editorState := range editorSharedState.editorStates {
			if editorState.editorStatePub != thisClientEditorState.editorStatePub {
				editorState.editorStatePub <- editor
			}
		}

		if err := enc.Encode(editor); err != nil {
			log.Printf("Error encoding to connection: %v", err)
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
