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
	Content *Content
	Cursor  *Cursor

	ScreenHeight int
	ScreenWidth  int

	FilePath string
	FileName string

	Mode EditorMode
}

func (editor *Editor) SaveContent() {
	string := string(editor.GetContent())
	err := os.WriteFile(editor.FilePath, []byte(string), 0644)
	if err != nil {
		panic(err)
	}
}

func InitializeEditor(path string, screenHeight int, screenWidth int) Editor {
	fileName := filepath.Base(path)

	cursor := Cursor{Index: 0, Row: 0, Col: 0}

	content := Content{}
	content.loadFromFile(path)

	editor := Editor{
		Content:      &content,
		Cursor:       &cursor,
		FilePath:     path,
		FileName:     fileName,
		Mode:         Normal,
		ScreenHeight: screenHeight,
		ScreenWidth:  screenWidth,
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

	newRow := editor.Cursor.Row + rowOffset
	newCol := editor.Cursor.Col + colOffset
	newIndex := -1

	if newRow < 0 {
		newRow = 0
	}

	if newCol < 0 || firstCol {
		newCol = 0
	}

	if newRow <= 0 && newCol <= 0 && !lastCol {
		newIndex = 0

		editor.Cursor.Row = newRow
		editor.Cursor.Col = newCol
		editor.Cursor.Index = newIndex

		return
	}

	index, row, col := 0, 0, 0

	for _, r := range editor.Content.calculateContent() {

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
		newIndex = editor.Content.Length
	}

	editor.Cursor.Row = newRow
	editor.Cursor.Col = newCol
	editor.Cursor.Index = newIndex
}

func (editor *Editor) InsertRune(r rune) {
	editor.Content.replace(
		[]rune{r},
		editor.Cursor.Index,
		editor.Cursor.Index,
	)
	if r == rune('\n') {
		editor.ShiftCursor(1, 0, true, false)
	} else {
		editor.ShiftCursor(0, 1, false, false)
	}
}

func (editor *Editor) Backspace() {
	delIdx := editor.Cursor.Index - 1
	if delIdx < 0 {
		return
	}

	r := editor.Content.calculateContent()[delIdx]

	if r == rune('\n') {
		editor.ShiftCursor(-1, 0, false, true)
	} else {
		editor.ShiftCursor(0, -1, false, false)
	}

	if editor.Cursor.Index >= 0 {
		editor.Content.replace([]rune{}, delIdx, delIdx+1)
	}
}

// maybe consume as list of lines for rendering since all in memory in anyways?
func (editor *Editor) GetContent() []rune {
	return editor.Content.calculateContent()
}

func (editor *Editor) GetStatusBar() []rune {
	row, col := editor.Cursor.Row, editor.Cursor.Col

	leftContent := []rune(" ")
	if editor.Mode == Normal {
		leftContent = append(leftContent, []rune("NORMAL")...)
	} else if editor.Mode == Insert {
		leftContent = append(leftContent, []rune("INSERT")...)
	}
	leftContent = append(leftContent, rune(' '), rune('|'), rune(' '))
	leftContent = append(leftContent, []rune(editor.FileName)...)

	rightContent := []rune(strconv.FormatInt(int64(row), 10))
	rightContent = append(rightContent, rune(':'))
	rightContent = append(rightContent, []rune(strconv.FormatInt(int64(col), 10))...)
	rightContent = append(rightContent, rune(' '))

	spaceBetween := editor.ScreenWidth - len(leftContent) - len(rightContent)

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
	editor.Mode = Normal
}

func (editor *Editor) ToInsert(after bool) {
	if after {
		editor.ShiftCursor(0, 1, false, false)
	}
	editor.Mode = Insert
}
