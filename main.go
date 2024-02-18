package main

import (
	"bufio"
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"unicode"

	"golang.org/x/term"
)

type editor struct {
	reader  *bufio.Reader
	termios *term.State
	buffer  string
	version string
	m_rows  int
	m_cols  int
	c_row   int
	c_col   int
}

func check(err error) {
	if err == nil {
		return
	}

	fmt.Print("\x1b[2J")
	fmt.Print("\x1b[H")

	panic(err)
}

func (ed *editor) read_key() rune {
	c, _, err := ed.reader.ReadRune()
	check(err)

	return c
}

func (ed *editor) process_keypress() {
	c := ed.read_key()

	switch c {
	case 17:
		fmt.Print("\x1b[2J")
		fmt.Print("\x1b[H")
		os.Exit(0)
	case 'w':
		fallthrough
	case 'a':
		fallthrough
	case 's':
		fallthrough
	case 'd':
		ed.move_cursor(c)
	case unicode.ReplacementChar:
		panic(errors.New("Invalid Character"))
	}
}

func (ed *editor) move_cursor(c rune) {
	switch c {
	case 'w':
		ed.c_row--
	case 'a':
		ed.c_col--
	case 's':
		ed.c_row++
	case 'd':
		ed.c_col++
	}
}

func (ed *editor) refresh_screen() {
	ed.buffer += "\x1b[?25l"
	ed.buffer += "\x1b[H"
	ed.buffer += "\x1b[%1;%1H"

	ed.draw_rows()

	cursor_position := fmt.Sprintf("\x1b[%d;%dH", ed.c_row+1, ed.c_col+1)
	ed.buffer += cursor_position

	ed.buffer += "\x1b[?25h"

	fmt.Print(ed.buffer)
}

func (ed *editor) draw_rows() {
	ed.buffer = ""
	for i := 0; i < ed.m_cols-1; i++ {

		if i < ed.m_cols-1 {
			ed.buffer += "\r\n"
		}

		if i == ed.m_rows/3 {
			welcome_message := fmt.Sprintf("Simple Editor -- v%s", ed.version)
			msg_len := int(math.Min(float64(ed.m_cols-1),
				float64(len(welcome_message))))

			padding := (ed.m_rows - msg_len) / 2
			if padding > 0 {
				ed.buffer += "~"
				padding--
			}

			for padding > 0 {
				ed.buffer += " "
				padding--
			}

			ed.buffer += welcome_message[:msg_len]
		} else {
			ed.buffer += "~"
		}

		ed.buffer += "\x1b[K"
	}
}

func (ed *editor) get_cursor_position() (uint16, uint16) {
	fmt.Print("\x1b[6n\r\n")

	rows_str := ""
	cols_str := ""
	found_sc := false
	for {
		c, _, err := ed.reader.ReadRune()
		check(err)

		if c == -1 || c == 'R' {
			break
		} else if c == ';' {
			found_sc = true
			continue
		}

		if found_sc {
			cols_str += string(c)
		} else {
			rows_str += string(c)
		}
	}

	var cols, rows int64
	var err error

	fmt.Println(len(rows_str))

	rows, err = strconv.ParseInt(rows_str[2:], 10, 16)
	check(err)

	cols, err = strconv.ParseInt(cols_str, 10, 16)
	check(err)

	return uint16(rows), uint16(cols)
}

func (ed *editor) get_win_size() {
	fmt.Print("\x1b[999C\x1b[999B")
	cols, rows := ed.get_cursor_position()

	ed.m_cols, ed.m_rows = int(cols), int(rows)
}

func (ed *editor) init() {
	ed.c_row = 0
	ed.c_col = 0

	ed.get_win_size()
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	ed := &editor{reader: reader, version: "0.0.1"}

	var err error

	ed.termios, err = term.MakeRaw(int(os.Stdin.Fd()))
	check(err)

	ed.init()

	defer term.Restore(int(os.Stdin.Fd()), ed.termios)

	for {
		ed.refresh_screen()
		ed.process_keypress()
	}
}
