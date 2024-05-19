package main

type Editor struct {
	content *Content
	cursor  *Cursor
}

func initialize_editor(path string) Editor {
	cursor := Cursor{index: 0}
	content := Content{}

	content.load_from_file(path)

	editor := Editor{content: &content, cursor: &cursor}

	return editor
}

func (editor *Editor) shift_cursor(offset int) {
	new_index := editor.cursor.index + offset

	if new_index < 0 {
		new_index = 0
	}

	if new_index >= editor.content.length {
		new_index = editor.content.length - 1
	}

	editor.cursor.index = new_index
}

func (editor *Editor) get_content() []rune {
	return editor.content.calculate_content()
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
	editor.shift_cursor(1)
}

func (editor *Editor) backspace() {
  if editor.get_cursor_index() > 0 {
    editor.content.replace(
      []rune{},
      editor.get_cursor_index()-1,
      editor.get_cursor_index(),
    )
  }
  editor.shift_cursor(-1)
}
