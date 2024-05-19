package main

import (
	"os"
)

type PieceType int

const (
	original PieceType = iota
	add                = iota
)

type Piece struct {
	start  int
	length int
	kind   PieceType
	next   *Piece
}

type Content struct {
	original []rune
	add      []rune
	root     *Piece
	length   int
}

func (content *Content) load_from_file(path string) {
	raw_file_content, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	file_content := []rune{}

	for _, r := range string(raw_file_content) {
		file_content = append(file_content, r)
	}

	content.original = file_content
	content.add = []rune{}

	original_piece := Piece{
		start:  0,
		length: len(file_content),
		kind:   original,
		next:   nil,
	}

  content.length = original_piece.length

	content.root = &original_piece
}

// TODO add a bunch of error cases, should return error
func (content *Content) replace(r []rune, start int, end int) {
	/*
	   start, end inclusive

	   start == length, means append
	   start < length, means insert

	   assert start <= end
	   assert start <= length of all pieces
	   assert end <= length of all pieces

	   r can just be empty, thats a deletion

	   insert and delete instead of direct use of replace
	*/

	new_piece := Piece{
		start:  len(content.add),
		length: len(r),
		kind:   add,
		next:   nil,
	}

	content.add = append(content.add, r...)

	// CASE 1: no content
	if content.root == nil {
		// assert start == 0 and end == 0
		content.root = &new_piece
		content.length = new_piece.length
		return
	}

	// CASE 2: prepend
	if start == 0 && end == 0 {
		new_piece.next = content.root
		content.root = &new_piece
		content.length = content.length + new_piece.length
		return
	}

	if start == content.length && end == content.length {
		piece := content.root
		for {
			if piece.next == nil {
				break
			}
			piece = piece.next
		}

		piece.next = &new_piece
		content.length = content.length + new_piece.length
		return
	}

	// CASE 4: inserting
	if start == end {
		piece_start := 0
		for piece := content.root; piece != nil; piece = piece.next {
			piece_end := piece_start + piece.length
			
      if piece_end == start {
        temp := piece.next
        piece.next = &new_piece
        new_piece.next = temp
        content.length = content.length + new_piece.length
        return
      } else if piece_start < start && start < piece_end {
        pl := Piece{
          start: piece.start,
          length: start - piece_start,
          kind: piece.kind,
          next: nil,
        }
        pr := Piece{
          start: pl.start + pl.length,
          length: piece.length - pl.length,
          kind: piece.kind,
          next: nil,
        }

        temp := piece.next
        piece.start = pl.start
        piece.length = pl.length
        piece.next = &new_piece
        new_piece.next = &pr
        pr.next=temp

        content.length = content.length + new_piece.length
        return
      }

			piece_start = piece_end
		}
	}

	// Case 5: replacing
	// delete stuff first
	piece_start := 0
	for piece := content.root; piece != nil; piece = piece.next {
		piece_end := piece_start + piece.length
		if piece_start <= start && start < piece_end {
		}
		if start < piece_start && end >= piece_end {
		}
		if piece_start <= end && piece_end > end {
		}

		piece_start = piece_end
	}
}

func (content *Content) calculate_content() []rune {
	final_string := []rune{}

	for piece := content.root; piece != nil; piece = piece.next {
		var piece_string []rune = []rune{}

		// TODO Put this as a function of piece
		start := piece.start
		end := piece.start + piece.length
		if piece.kind == add {
			piece_string = content.add[start:end]
		} else if piece.kind == original {
			piece_string = content.original[start:end]
		}

		final_string = append(final_string, piece_string...)
	}

	return final_string
}