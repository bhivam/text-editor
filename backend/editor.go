package backend

import (
	"os"
	"path/filepath"
	"strconv"
)

type EditorMode int

const (
	Normal EditorMode = iota
	Insert            = iota
)

type Editor struct {
	content *Content
	cursor  *Cursor

	screenHeight int
	screenWidth  int

	filePath string
	fileName string

	mode EditorMode
}

func (editor *Editor) Length() int {
  return editor.content.length
}

func (editor *Editor) ScreenHeight() int {
  return editor.screenHeight;
}

func (editor *Editor) SetScreenHeight(height int) {
  editor.screenHeight = height;
}

func (editor *Editor) ScreenWidth() int {
  return editor.screenWidth;
}

func (editor *Editor) SetScreenWidth(width int) {
  editor.screenWidth = width;
}

func (editor *Editor) Mode() EditorMode {
  return editor.mode
}

func (editor *Editor) SaveContent() {
	string := string(editor.GetContent())
	err := os.WriteFile(editor.filePath, []byte(string), 0644)
	if err != nil {
		panic(err)
	}
}

func InitializeEditor(path string, screenHeight int, screenWidth int) Editor {
	fileName := filepath.Base(path)

	cursor := Cursor{index: 0, row: 0, col: 0}

	content := Content{}
	content.loadFromFile(path)

	editor := Editor{
		content:       &content,
		cursor:        &cursor,
		filePath:     path,
		fileName:     fileName,
		mode:          Normal,
		screenHeight: screenHeight,
		screenWidth:  screenWidth,
	}

	return editor
}

func (editor *Editor) ShiftCursor(
	rowOffset int,
	colOffset int,
	firstCol bool,
	lastCol bool,
) {
	// TODO cursor should not be able to go to last index for normal mode

	newRow := editor.cursor.row + rowOffset
	newCol := editor.cursor.col + colOffset
	newIndex := -1

	if newRow < 0 {
		newRow = 0
	}

	if newCol < 0 || firstCol {
		newCol = 0
	}

	if newRow <= 0 && newCol <= 0 && !lastCol {
		newIndex = 0

		editor.cursor.row = newRow
		editor.cursor.col = newCol
		editor.cursor.index = newIndex

		return
	}

	index, row, col := 0, 0, 0

	for _, r := range editor.content.calculateContent() {

		if newRow == row && r == '\n' && (newCol > col || lastCol) {
			newCol = col
			lastCol = false
		}

		if newRow == row && newCol == col && !lastCol {
			newIndex = index
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

	if newIndex == -1 {
		newRow = row
		newCol = col
		newIndex = editor.content.length
	}

	editor.cursor.row = newRow
	editor.cursor.col = newCol
	editor.cursor.index = newIndex
}

func (editor *Editor) CursorRow() int {
  return editor.cursor.row
}

func (editor *Editor) CursorCol() int {
  return editor.cursor.col
}

func (editor *Editor) CursorIndex() int {
	return editor.cursor.index
}

func (editor *Editor) InsertRune(r rune) {
	editor.content.replace(
		[]rune{r},
		editor.CursorIndex(),
		editor.CursorIndex(),
	)
	if r == rune('\n') {
		editor.ShiftCursor(1, 0, true, false)
	} else {
		editor.ShiftCursor(0, 1, false, false)
	}
}

func (editor *Editor) Backspace() {
	delIdx := editor.CursorIndex() - 1
	if delIdx < 0 {
		return
	}

	r := editor.content.calculateContent()[delIdx]

	if r == rune('\n') {
		editor.ShiftCursor(-1, 0, false, true)
	} else {
		editor.ShiftCursor(0, -1, false, false)
	}

	if editor.CursorIndex() >= 0 {
		editor.content.replace([]rune{}, delIdx, delIdx+1)
	}
}

// maybe consume as list of lines for rendering since all in memory in anyways?
func (editor *Editor) GetContent() []rune {
	return editor.content.calculateContent()
}

func (editor *Editor) GetStatusBar() []rune {
	row, col := editor.cursor.row, editor.cursor.col

	leftContent := []rune(" ")
	if editor.mode == Normal {
		leftContent = append(leftContent, []rune("NORMAL")...)
	} else if editor.mode == Insert {
		leftContent = append(leftContent, []rune("INSERT")...)
	}
	leftContent = append(leftContent, rune(' '), rune('|'), rune(' '))
	leftContent = append(leftContent, []rune(editor.fileName)...)

	rightContent := []rune(strconv.FormatInt(int64(row), 10))
	rightContent = append(rightContent, rune(':'))
	rightContent = append(rightContent, []rune(strconv.FormatInt(int64(col), 10))...)
	rightContent = append(rightContent, rune(' '))

	spaceBetween := editor.screenWidth - len(leftContent) - len(rightContent)

	if spaceBetween < 1 {
		return []rune{}
	}

	statusLine := leftContent
	for range spaceBetween {
		statusLine = append(statusLine, rune(' '))
	}
	statusLine = append(statusLine, rightContent...)

	return statusLine
}

func (editor *Editor) ToNormal() {
	editor.ShiftCursor(0, -1, false, false)
	editor.mode = Normal 
}

func (editor *Editor) ToInsert(after bool) {
	if after {
		editor.ShiftCursor(0, 1, false, false)
	}
	editor.mode = Insert 
}
