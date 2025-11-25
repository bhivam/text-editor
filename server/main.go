package main

import (
	"encoding/json"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/google/uuid"
	"log"
	"net"
	"path/filepath"
	"sync"

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

type ClientEditorEvent struct {
	clientID string
	event    EditorEvent
}

type IndividualEditorState struct {
	editor *backend.Editor
	enc    *json.Encoder
}

type FileEditSession struct {
	content       *backend.Content
	editorStates  map[string]*IndividualEditorState
	clientEventCh chan ClientEditorEvent
	mu            sync.RWMutex
}

var fileEditSessions map[string]*FileEditSession = make(map[string]*FileEditSession)
var sessionsMu sync.RWMutex

func processClientEvents(fileEditSession *FileEditSession) {
	for {
		clientEvent := <-fileEditSession.clientEventCh

		currClientID := clientEvent.clientID
		event := clientEvent.event

		fileEditSession.mu.RLock()

		editor := fileEditSession.editorStates[currClientID].editor

		if event.IsKey {
			key := event.Key

			if key == tcell.KeyRune {
				fmt.Printf("Client %s sent '%c'\n", currClientID, event.Rune)
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

					case rune('a'):
						editor.ToInsert(true)
					case rune('i'):
						editor.ToInsert(false)

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

		for _, editorState := range fileEditSession.editorStates {
			if err := editorState.enc.Encode(editorState.editor); err != nil {
				log.Printf("Error encoding in goroutine: %v", err)
				return
			}
			log.Printf("Sent new editor state")
		}
		fileEditSession.mu.RUnlock()

		log.Printf("Client Event Received: %+v", clientEvent)
	}
}

func editorSubscribe(initArgs InitArgs, enc *json.Encoder) (*FileEditSession, string, *IndividualEditorState) {
	clientID := uuid.New().String()

	sessionsMu.Lock()
	defer sessionsMu.Unlock()

	if fileEditSession, ok := fileEditSessions[initArgs.FilePath]; ok {
		individualEditorState := &IndividualEditorState{
			editor: &backend.Editor{
				Content:      fileEditSession.content,
				Cursor:       &backend.Cursor{Index: 0, Row: 0, Col: 0},
				FilePath:     initArgs.FilePath,
				FileName:     filepath.Base(initArgs.FilePath),
				ScreenHeight: initArgs.ScreenHeight,
				ScreenWidth:  initArgs.ScreenWidth,
				Mode:         backend.Normal,
			},
			enc: enc,
		}

		fileEditSession.mu.Lock()
		fileEditSession.editorStates[clientID] = individualEditorState
		fileEditSession.mu.Unlock()

		log.Printf("Client %s subscribed to %s", clientID, initArgs.FilePath)
		return fileEditSession, clientID, individualEditorState
	} else {
		editor := backend.InitializeEditor(
			initArgs.FilePath,
			initArgs.ScreenHeight,
			initArgs.ScreenWidth,
		)

		individualEditorState := &IndividualEditorState{
			editor: &editor,
			enc:    enc,
		}

		fileEditSession := &FileEditSession{
			content:       editor.Content,
			editorStates:  make(map[string]*IndividualEditorState),
			clientEventCh: make(chan ClientEditorEvent, 10),
		}
		fileEditSession.editorStates[clientID] = individualEditorState

		fileEditSessions[initArgs.FilePath] = fileEditSession

		log.Printf("Client %s subscribed to %s (new session)", clientID, initArgs.FilePath)

		go processClientEvents(fileEditSession)

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
	fileEditSession.mu.RLock()
	delete(fileEditSession.editorStates, clientID)
	fileEditSession.mu.RUnlock()

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

	fileEditSession, currClientID, _ := editorSubscribe(initArgs, enc)
	defer editorUnsubscribe(initArgs.FilePath, currClientID)

	for {
		event := EditorEvent{}
		err := dec.Decode(&event)
		if err != nil {
			log.Printf("Client %s: Error reading from connection: %v", currClientID, err)
			return
		}

		if event.IsExit {
			log.Printf("Client %s: Received exit message", currClientID)
			return
		}

		fileEditSession.clientEventCh <- ClientEditorEvent{
			clientID: currClientID,
			event:    event,
		}
	}
}

func main() {
	ln, err := net.Listen("tcp", ":8081")
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatalf("Failed to accept connection: %v", err)
		}
		go handleConnection(conn)
	}
}
