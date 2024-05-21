package main

type Editor struct {
	content *Content
	cursor  *Cursor
}

func initialize_editor(path string) Editor {
	cursor := Cursor{index: 0, row: 0, col: 0}
	content := Content{}

	content.load_from_file(path)

	editor := Editor{content: &content, cursor: &cursor}

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

	if new_row <= 0 && new_col <= 0 && !last_col{
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
		col = max(col-1, 0)
		new_row = row
		new_col = col
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

	if editor.get_cursor_index() > 0 {
		editor.content.replace([]rune{}, del_idx, del_idx+1)
	}
}
