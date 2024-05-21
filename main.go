package main

import (
	"log"
	"strconv"

	"github.com/gdamore/tcell/v2"
)

func print_line_num(
	screen tcell.Screen,
	row *int,
	col *int,
	num_digits int,
	line_num_style tcell.Style,
) {
	nums := []rune(strconv.FormatInt(int64(*row), 10))
	if len(nums) < num_digits {
		for i := 0; i < num_digits-len(nums); i += 1 {
			nums = append([]rune(" "), nums...)
		}
	}
	screen.SetContent(*col, *row, rune(' '), nil, line_num_style)
	*col += 1
	screen.SetContent(*col, *row, nums[0], nil, line_num_style)
	*col += 1
	screen.SetContent(*col, *row, nums[1], nil, line_num_style)
	*col += 1
	screen.SetContent(*col, *row, rune(' '), nil, line_num_style)
	*col += 1
}

func main() {
	editor := initialize_editor("test.txt")

	def_style := tcell.StyleDefault.
		Foreground(tcell.ColorReset).
		Background(tcell.ColorReset)

	line_num_style := tcell.StyleDefault.
		Foreground(tcell.ColorDimGray.TrueColor()).
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

	editor.screen_width, editor.screen_height = screen.Size()

	for {
		// edit content based on new state
		screen.Clear()

		max_row_digits := 2

		i, row, col := 0, 0, 0
		editor_content := editor.get_content()
		print_line_num(screen, &row, &col, 2, line_num_style)

		for i < editor.content.length {
			r := editor_content[i]
			if r == '\n' {
				row = row + 1
				col = 0

				print_line_num(screen, &row, &col, max_row_digits, line_num_style)
			} else {
				screen.SetContent(col, row, r, nil, def_style)
				col += 1
			}
			i += 1
		}

		screen.ShowCursor(max_row_digits+2+editor.cursor.col, editor.cursor.row)

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
				editor.shift_cursor(0, 1, false, false)
			} else if key == tcell.KeyLeft {
				editor.shift_cursor(0, -1, false, false)
			} else if key == tcell.KeyUp {
				editor.shift_cursor(-1, 0, false, false)
			} else if key == tcell.KeyDown {
				editor.shift_cursor(1, 0, false, false)
			} else if key == tcell.KeyRune {
				editor.insert_rune(event.Rune())
			} else if key == tcell.KeyBackspace2 {
				editor.backspace()
			}

		case *tcell.EventResize:
			editor.screen_width, editor.screen_height = event.Size()
		}
	}
}
