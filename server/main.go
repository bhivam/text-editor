package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"path/filepath"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/google/uuid"

	"github.com/bhivam/text-editor/backend"
)

type InitArgs struct {
	ScreenHeight int
	ScreenWidth  int
	FilePath     string
}

type EditorEvent struct {
	DispatchTime int64
	IsKey        bool
	IsExit       bool
	Key          tcell.Key
	Rune         rune
	Width        int
	Height       int
}

type IndividualEditorState struct {
	editorStatePub chan backend.Editor
	editor         backend.Editor
}

type FileEditSession struct {
	content      *backend.Content
	editorStates map[string]*IndividualEditorState
	mu           sync.RWMutex
}

var fileEditSessions map[string]*FileEditSession = make(map[string]*FileEditSession)
var sessionsMu sync.RWMutex

func editorSubscribe(initArgs InitArgs) (*FileEditSession, string, *IndividualEditorState) {
	clientID := uuid.New().String()
	newEditorStatePub := make(chan backend.Editor, 10) // buffered channel

	sessionsMu.Lock()
	defer sessionsMu.Unlock()

	if fileEditSession, ok := fileEditSessions[initArgs.FilePath]; ok {
		individualEditorState := &IndividualEditorState{
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

		fileEditSession.mu.Lock()
		fileEditSession.editorStates[clientID] = individualEditorState
		fileEditSession.mu.Unlock()

		log.Printf("Client %s subscribed to %s", clientID, initArgs.FilePath)
		return fileEditSession, clientID, individualEditorState
	} else {
		log.Printf("Reading from %s %d", initArgs.FilePath, len(initArgs.FilePath))

		editor := backend.InitializeEditor(
			initArgs.FilePath,
			initArgs.ScreenHeight,
			initArgs.ScreenWidth,
		)

		individualEditorState := &IndividualEditorState{
			editorStatePub: newEditorStatePub,
			editor:         editor,
		}

		fileEditSession := &FileEditSession{
			content:      editor.Content,
			editorStates: make(map[string]*IndividualEditorState),
		}
		fileEditSession.editorStates[clientID] = individualEditorState

		fileEditSessions[initArgs.FilePath] = fileEditSession

		log.Printf("Client %s subscribed to %s (new session)", clientID, initArgs.FilePath)
		return fileEditSession, clientID, individualEditorState
	}
}

func editorUnsubscribe(filePath string, clientID string) {
	sessionsMu.RLock()
	defer sessionsMu.RUnlock()

	fileEditSession, ok := fileEditSessions[filePath]

	if !ok {
		return
	}

	if state, exists := fileEditSession.editorStates[clientID]; exists {
		close(state.editorStatePub)
		delete(fileEditSession.editorStates, clientID)
	}

	log.Printf("Client %s unsubscribed from %s", clientID, filePath)
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

	fileEditSession, clientID, thisClientEditorState := editorSubscribe(initArgs)
	defer editorUnsubscribe(initArgs.FilePath, clientID)

	editor := thisClientEditorState.editor

	go func() {
		for {
			newEditor, ok := <-thisClientEditorState.editorStatePub
			if !ok {
				return
			}

			editor.Content = newEditor.Content

			if err := enc.Encode(editor); err != nil {
				log.Printf("Client %s: Error encoding in goroutine: %v", clientID, err)
				return
			}
		}
	}()

	for {
		event := EditorEvent{}
		err := dec.Decode(&event)
		if err != nil {
			log.Printf("Client %s: Error reading from connection: %v", clientID, err)
			return
		}

		// Handle exit message
		if event.IsExit {
			log.Printf("Client %s: Received exit message", clientID)
			return
		}

		if event.IsKey {
			key := event.Key

			if key == tcell.KeyRune {
				fmt.Printf("Client %s sent '%c'\n", clientID, event.Rune)
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

		// TODO prefer event reception time over arbitrary race
		fileEditSession.mu.RLock()
		for id, editorState := range fileEditSession.editorStates {
			if id != clientID {
				select {
				case editorState.editorStatePub <- editor:
				default:
					log.Printf("Client %s: Channel full, skipping update", id)
				}
			}
		}
		fileEditSession.mu.RUnlock()

		if err := enc.Encode(editor); err != nil {
			log.Printf("Client %s: Error encoding to connection: %v", clientID, err)
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
