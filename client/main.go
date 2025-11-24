package main

import (
	"encoding/json"
	"flag"
	"log"
	"net"
	"os"
	"strconv"

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

func printLineNum(
	screen tcell.Screen,
	row *int,
	col *int,
	numDigits int, lineNumStyle tcell.Style,
) {
	nums := []rune(strconv.FormatInt(int64(*row), 10))
	if len(nums) < numDigits {
		for i := 0; i < numDigits-len(nums); i += 1 {
			nums = append([]rune(" "), nums...)
		}
	}
	screen.SetContent(*col, *row, rune(' '), nil, lineNumStyle)
	*col += 1
	screen.SetContent(*col, *row, nums[0], nil, lineNumStyle)
	*col += 1
	screen.SetContent(*col, *row, nums[1], nil, lineNumStyle)
	*col += 1
	screen.SetContent(*col, *row, rune(' '), nil, lineNumStyle)
	*col += 1
}

func tcpFileEdit(remoteHost string, fileName string) {
	conn, err := net.Dial("tcp", remoteHost)
	if err != nil {
		return
	}
	defer conn.Close()

	screen, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("%+v", err)
	}

	if err := screen.Init(); err != nil {
		log.Fatalf("%+v", err)
	}

	initScreenWidth, initScreenHeight := screen.Size()

	initArgs := InitArgs{
		FileName:     fileName,
		ScreenWidth:  initScreenWidth,
		ScreenHeight: initScreenHeight,
	}

	enc := json.NewEncoder(conn)
	dec := json.NewDecoder(conn)

	err = enc.Encode(initArgs)
	if err != nil {
		return
	}

	defStyle := tcell.StyleDefault.
		Foreground(tcell.ColorReset.TrueColor()).
		Background(tcell.ColorReset.TrueColor())

	lineNumStyle := tcell.StyleDefault.
		Foreground(tcell.ColorDimGray.TrueColor()).
		Background(tcell.ColorReset.TrueColor())

	statusBarStyle := tcell.StyleDefault.
		Foreground(tcell.ColorBlack.TrueColor()).
		Background(tcell.ColorFloralWhite).
		Bold(true)

	screen.SetStyle(defStyle)
	screen.EnableMouse()
	screen.EnablePaste()
	screen.Clear()

	editor := backend.Editor{}

	go func() {
		for {
			dec.Decode(&editor)
			renderEditor(screen, editor, defStyle, lineNumStyle, statusBarStyle)
		}
	}()

	for {
		event := screen.PollEvent()

		editorEvent := EditorEvent{}
		switch event := event.(type) {
		case *tcell.EventKey:
			editorEvent.IsKey = true
			editorEvent.Key = event.Key()
			editorEvent.Rune = event.Rune()
		case *tcell.EventResize:
			editorEvent.IsKey = false
			editorEvent.Width, editorEvent.Height = event.Size()
		}

		err := enc.Encode(editorEvent)
		if err != nil {
			return
		}

		// update state based on new event
		switch event := event.(type) {
		case *tcell.EventKey:

			key := event.Key()

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
					editor.InsertRune(event.Rune())
				case tcell.KeyBackspace2:
					editor.Backspace()
				}
			case backend.Normal:
				switch key {
				case tcell.KeyRune:
					keyVal := event.Rune()
					switch keyVal {
					case rune('q'):
						conn.Close()
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

		case *tcell.EventResize:
			editor.ScreenWidth, editor.ScreenHeight = event.Size()
		}
	}
}

func localFileEdit(fileName string) {
	defStyle := tcell.StyleDefault.
		Foreground(tcell.ColorReset.TrueColor()).
		Background(tcell.ColorReset.TrueColor())

	lineNumStyle := tcell.StyleDefault.
		Foreground(tcell.ColorDimGray.TrueColor()).
		Background(tcell.ColorReset.TrueColor())

	statusBarStyle := tcell.StyleDefault.
		Foreground(tcell.ColorBlack.TrueColor()).
		Background(tcell.ColorFloralWhite).
		Bold(true)

	screen, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("%+v", err)
	}
	if err := screen.Init(); err != nil {
		log.Fatalf("%+v", err)
	}

	screen.SetStyle(defStyle)
	screen.EnableMouse()
	screen.EnablePaste()
	screen.Clear()

	initScreenHeight, initScreenWidth := screen.Size()
	editor := backend.InitializeEditor(fileName, initScreenHeight, initScreenWidth)

	quit := func() {
		maybePanic := recover()
		screen.Fini()
		if maybePanic != nil {
			panic(maybePanic)
		}

		// save content to file
		editor.SaveContent()
		os.Exit(0)
	}

	defer quit()

	for {
		renderEditor(
			screen,
			editor,
			defStyle,
			lineNumStyle,
			statusBarStyle,
		)

		// poll for new event
		event := screen.PollEvent()

		// update state based on new event
		switch event := event.(type) {
		case *tcell.EventKey:

			key := event.Key()

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
					editor.InsertRune(event.Rune())
				case tcell.KeyBackspace2:
					editor.Backspace()
				}
			case backend.Normal:
				switch key {
				case tcell.KeyRune:
					keyVal := event.Rune()
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

		case *tcell.EventResize:
			editor.ScreenWidth, editor.ScreenHeight = event.Size()
		}
	}
}

func renderEditor(
	screen tcell.Screen,
	editor backend.Editor,
	defStyle tcell.Style,
	lineNumStyle tcell.Style,
	statusBarStyle tcell.Style,
) {
	screen.Clear()

	maxRowDigits := 2
	i, row, col := 0, 0, 0
	editorContent := editor.GetContent()
	printLineNum(screen, &row, &col, 2, lineNumStyle)
	for i < editor.Content.Length && row < editor.ScreenHeight-1 {
		r := editorContent[i]
		if r == '\n' {
			row = row + 1
			col = 0

			printLineNum(screen, &row, &col, maxRowDigits, lineNumStyle)
		} else {
			screen.SetContent(col, row, r, nil, defStyle)
			col += 1
		}
		i += 1
	}

	statusBar := editor.GetStatusBar()
	row = editor.ScreenHeight - 1
	for col, r := range statusBar {
		screen.SetContent(col, row, r, nil, statusBarStyle)
	}

	screen.ShowCursor(maxRowDigits+2+editor.Cursor.Col, editor.Cursor.Row)

	// show new buffer
	screen.Show()
}

func main() {
	var fileName string

	isRemote := flag.Bool("R", false, "Specify remote host and port")

	flag.Parse()

	if *isRemote {
		if len(flag.Args()) < 2 {
			log.Fatal("Please specify a remote host and port after -R, followed by a file name")
			os.Exit(1)
		}
		remoteHost := flag.Arg(0) // remote host and port
		fileName = flag.Arg(1)    // file name

		log.Println("remoteHost: ", remoteHost)
		log.Println("fileName: ", fileName)

		tcpFileEdit(remoteHost, fileName)
		os.Exit(0)
	} else {
		if len(flag.Args()) < 1 {
			log.Fatal("Please specify a file name")
			os.Exit(1)
		}
		fileName = flag.Arg(0)

		localFileEdit(fileName)
		os.Exit(0)
	}
}
