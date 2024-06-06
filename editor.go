package main

import (
	"path/filepath"
	"strconv"
)

type Editor struct {
	content *Content
	cursor  *Cursor

	screen_height int
	screen_width  int

	// should these be rune lists?
	file_path string
	file_name string
}

func initialize_editor(path string) Editor {
	file_name := filepath.Base(path)

	cursor := Cursor{index: 0, row: 0, col: 0}

	content := Content{}
	content.load_from_file(path)

	editor := Editor{
		content:   &content,
		cursor:    &cursor,
		file_path: path,
		file_name: file_name,
	}

	return editor
}

func (editor *Editor) shift_cursor(
	row_offset int,
	col_offset int,
	first_col bool,
	last_col bool,
) {
	new_row := editor.cursor.row + row_offset
	new_col := editor.cursor.col + col_offset
	new_index := -1

	if new_row < 0 {
		new_row = 0
	}

	if new_col < 0 || first_col {
		new_col = 0
	}

	if new_row <= 0 && new_col <= 0 && !last_col {
		new_index = 0

		editor.cursor.row = new_row
		editor.cursor.col = new_col
		editor.cursor.index = new_index

		return
	}

	index, row, col := 0, 0, 0

	for _, r := range editor.content.calculate_content() {

		if new_row == row && r == '\n' && (new_col > col || last_col) {
			new_col = col
			last_col = false
		}

		if new_row == row && new_col == col && !last_col {
			new_index = index
			break
		}

		if r == '\n' {
			row = row + 1
			col = 0
		} else {
			col = col + 1
		}
		index += 1
	}

	if new_index == -1 {
		new_row = row
		new_col = col
		new_index = editor.content.length
	}

	editor.cursor.row = new_row
	editor.cursor.col = new_col
	editor.cursor.index = new_index
}

func (editor *Editor) get_cursor_index() int {
	return editor.cursor.index
}

func (editor *Editor) insert_rune(r rune) {
	editor.content.replace(
		[]rune{r},
		editor.get_cursor_index(),
		editor.get_cursor_index(),
	)
	if r == rune('\n') {
		editor.shift_cursor(1, 0, true, false)
	} else {
		editor.shift_cursor(0, 1, false, false)
	}
}

func (editor *Editor) backspace() {
	del_idx := editor.get_cursor_index() - 1
	if del_idx < 0 {
		return
	}

	r := editor.content.calculate_content()[del_idx]

	if r == rune('\n') {
		editor.shift_cursor(-1, 0, false, true)
	} else {
		editor.shift_cursor(0, -1, false, false)
	}

	if editor.get_cursor_index() >= 0 {
		editor.content.replace([]rune{}, del_idx, del_idx+1)
	}
}

// maybe consume as list of lines for rendering since all in memory in anyways?
func (editor *Editor) get_content() []rune {
	return editor.content.calculate_content()
}

func (editor *Editor) get_status_bar() []rune {
	row, col := editor.cursor.row, editor.cursor.col

	left_content := []rune(" ")
	left_content = append(left_content, []rune(editor.file_name)...)
	left_content = append(left_content, rune(' '))
	left_content = append(
		left_content,
		[]rune(strconv.FormatInt(int64(editor.content.num_pieces), 10))...,
	)

	right_content := []rune(strconv.FormatInt(int64(row), 10))
	right_content = append(right_content, rune(':'))
	right_content = append(right_content, []rune(strconv.FormatInt(int64(col), 10))...)
	right_content = append(right_content, rune(' '))

	space_between := editor.screen_width - len(left_content) - len(right_content)

	if space_between < 1 {
		return []rune{}
	}

	status_line := left_content
	for range space_between {
		status_line = append(status_line, rune(' '))
	}
	status_line = append(status_line, right_content...)

	return status_line
}
