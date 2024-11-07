package main

import (
	"log"
	"os"
	"strconv"

	"github.com/gdamore/tcell/v2"

	"github.com/bhivam/text-editor/backend"
)

func printLineNum(
	screen tcell.Screen,
	row *int,
	col *int,
	numDigits int,
	lineNumStyle tcell.Style,
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

func main() {
	args := os.Args

	if len(args) != 2 {
		log.Fatalf("Usage: %s <filename>", args[0])
	}

	filename := args[1]

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
	editor := backend.InitializeEditor(filename, initScreenHeight, initScreenWidth)

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
		// edit content based on new state
		screen.Clear()

		// render content
		maxRowDigits := 2
		i, row, col := 0, 0, 0
		editorContent := editor.GetContent()
		printLineNum(screen, &row, &col, 2, lineNumStyle)
		for i < editor.Length() && row < editor.ScreenHeight()-1 {
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
		row = editor.ScreenHeight() - 1
		for col, r := range statusBar {
			screen.SetContent(col, row, r, nil, statusBarStyle)
		}

		screen.ShowCursor(maxRowDigits+2+editor.CursorCol(), editor.CursorRow())

		// show new buffer
		screen.Show()

		// poll for new event
		event := screen.PollEvent()

		// update state based on new event
		switch event := event.(type) {
		case *tcell.EventKey:

			key := event.Key()

			if editor.Mode() == backend.Insert {
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
					editor.InsertRune(event.Rune())
				} else if key == tcell.KeyBackspace2 {
					editor.Backspace()
				}
			} else if editor.Mode() == backend.Normal {
				if key == tcell.KeyRune {
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

		case *tcell.EventResize:
			w, h := event.Size()
			editor.SetScreenWidth(w)
			editor.SetScreenHeight(h)
		}
	}
}
