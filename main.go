package main

import (
	"log"

	"github.com/gdamore/tcell/v2"
)

func main() {
	editor := initialize_editor("test.txt")
	content := editor.content

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

	cursor := editor.cursor

	for {
		// edit content based on new state
		screen.Clear()

    cursor_row, cursor_col := -1, -1

		row, col := 0, 0
		for i, r := range content.calculate_content() {
			if i == cursor.index {
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
				content.replace([]rune{'\n'},
					cursor.index,
					cursor.index,
				)
				cursor.index += 1
			} else if key == tcell.KeyRune {
				content.replace([]rune{event.Rune()},
					cursor.index,
					cursor.index,
				)
				cursor.index += 1
			}
		}
	}
}
