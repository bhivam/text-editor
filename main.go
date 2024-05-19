package main

import (
	"log"

	"github.com/gdamore/tcell/v2"
)

func main() {
	editor := initialize_editor("test.txt")

	def_style := tcell.StyleDefault.
		Foreground(tcell.ColorReset).
		Background(tcell.ColorReset)

	screen, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("%+v", err)
	}
	if err := screen.Init(); err != nil {
		log.Fatalf("%+v", err)
	}

	screen.SetStyle(def_style)
	screen.EnableMouse()
	screen.EnablePaste()
	screen.Clear()

	quit := func() {
		maybe_panic := recover()
		screen.Fini()
		if maybe_panic != nil {
			panic(maybe_panic)
		}
	}

	defer quit()

	for {
		// edit content based on new state
		screen.Clear()

		cursor_row, cursor_col := -1, -1

		row, col := 0, 0
		for i, r := range editor.get_content() {
			if i == editor.get_cursor_index() {
				cursor_row, cursor_col = row, col
			}

			if r == '\n' {
				row = row + 1
				col = 0
			} else {
				screen.SetContent(col, row, r, nil, def_style)
				col = col + 1
			}
		}

		screen.ShowCursor(cursor_col, cursor_row)

		// show new buffer
		screen.Show()

		// poll for new event
		event := screen.PollEvent()

		// update state based on new event
		switch event := event.(type) {
		case *tcell.EventKey:

			key := event.Key()

			if key == tcell.KeyEscape {
				return
			} else if key == tcell.KeyEnter {
        editor.insert_rune('\n')
			} else if key == tcell.KeyRight {
        editor.shift_cursor(1)
			} else if key == tcell.KeyLeft {
        editor.shift_cursor(-1)
			} else if key == tcell.KeyRune {
        editor.insert_rune(event.Rune())
			} else if key == tcell.KeyBackspace2 {
        editor.backspace()
			}
		}
	}
}
