package backend

import (
	"fmt"
	"os"
	"time"
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
	original    []rune
	add         []rune
	contentRoot *Piece
	length      int
	lastEdit    int64
	numPieces   int
}

func (content *Content) loadFromFile(path string) {
	content.lastEdit = -1

	rawFileContent, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	fileContent := []rune{}

	for _, r := range string(rawFileContent) {
		fileContent = append(fileContent, r)
	}

	content.original = fileContent
	content.add = []rune{}

	originalPiece := Piece{
		start:  0,
		length: len(fileContent) - 1,
		kind:   original,
		next:   nil,
	}

	content.length = originalPiece.length
	content.contentRoot = &originalPiece
	content.numPieces = 1
}

func (content *Content) undo() {
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
	currentTime := time.Now().UnixMilli()
	coalesce := currentTime-content.lastEdit < 1000
	content.lastEdit = currentTime

	newPiece := Piece{
		start:  len(content.add),
		length: len(r),
		kind:   add,
		next:   nil,
	}

	if len(r) > 0 {
		content.add = append(content.add, r...)
	}

	// CASE 1: no content
	if content.contentRoot == nil {
		// assert start == 0 and end == 0
		content.contentRoot = &newPiece
		content.length = newPiece.length
		return
	}

	// CASE 2: prepend
	if start == 0 && end == 0 {
		newPiece.next = content.contentRoot
		content.contentRoot = &newPiece
		content.length = content.length + newPiece.length
		return
	}

	// CASE 3: prepend
	if start == content.length && end == content.length {
		piece := content.contentRoot
		for {
			if piece.next == nil {
				break
			}
			piece = piece.next
		}

		piece.next = &newPiece
		content.length = content.length + newPiece.length
		return
	}

	// CASE 4: inserting
	if start == end {
		pieceStart := 0
		for piece := content.contentRoot; piece != nil; piece = piece.next {
			pieceEnd := pieceStart + piece.length

			if coalesce && piece.kind == add &&
				piece.start+piece.length == newPiece.start &&
				pieceEnd == start {

				piece.length += newPiece.length
				content.length += newPiece.length
				return
			}

			if pieceEnd == start {
				temp := piece.next
				piece.next = &newPiece
				newPiece.next = temp
				content.length = content.length + newPiece.length
				return
			} else if pieceStart < start && start < pieceEnd {
				pl := Piece{
					start:  piece.start,
					length: start - pieceStart,
					kind:   piece.kind,
					next:   nil,
				}
				pr := Piece{
					start:  pl.start + pl.length,
					length: piece.length - pl.length,
					kind:   piece.kind,
					next:   nil,
				}

				temp := piece.next
				piece.start = pl.start
				piece.length = pl.length
				piece.next = &newPiece
				newPiece.next = &pr
				pr.next = temp

				content.length = content.length + newPiece.length
				return
			}

			pieceStart = pieceEnd
		}
	}

	// Case 5: replacing
	// delete stuff first, recursion to insert
	pieceStart := 0
	var prev *Piece = nil
	for piece := content.contentRoot; piece != nil; piece = piece.next {

		pieceEnd := pieceStart + piece.length

		if pieceStart < start && end < pieceEnd {
			// fmt.Println("going 1")
			content.length -= piece.length

			// set up right piece
			pr := &Piece{
				start:  piece.start + end - pieceStart,
				length: pieceEnd - end,
				kind:   piece.kind,
				next:   piece.next,
			}

			// set up left piece
			piece.length = start - pieceStart
			piece.next = pr

			content.length += piece.length + pr.length
			break
		}

		if pieceStart >= start && end >= pieceEnd {
			// fmt.Println("going 2")
			// fmt.Printf("%d %d %d %d\n", start, end, piece_start, piece_end)
			content.length -= piece.length
			if prev == nil {
				content.contentRoot = content.contentRoot.next
			} else {
				prev.next = piece.next
			}
			piece = content.contentRoot
			prev = nil

			if pieceEnd == end {
				break
			}

			continue

		} else if pieceStart >= start && end > pieceStart && end < pieceEnd {
			// fmt.Println("going 3")
			content.length -= piece.length
			piece.start = piece.start + end - start
			piece.length = pieceEnd - end

			content.length += piece.length
			if pieceStart == start {
				break
			}

			pieceEnd = pieceStart + piece.length
		} else if end >= pieceEnd && start > pieceStart && start < pieceEnd {
			// fmt.Println("going 4")
			content.length -= piece.length
			piece.length = start - pieceStart

			content.length += piece.length
			if pieceEnd == end {
				break
			}

			pieceEnd = pieceStart + piece.length
		}

		pieceStart = pieceEnd
		prev = piece
	}

	if len(r) > 0 {
		content.replace(r, start, start)
	}
}

func (content *Content) printPieces() {
	for piece := content.contentRoot; piece != nil; piece = piece.next {
		fmt.Println(piece)
	}
}

func (content *Content) calculateContent() []rune {
	finalString := []rune{}
	numPieces := 0

	for piece := content.contentRoot; piece != nil; piece = piece.next {
		var pieceString []rune = []rune{}

		// TODO Put this as a function of piece
		start := piece.start
		end := piece.start + piece.length
		if piece.kind == add {
			pieceString = content.add[start:end]
		} else if piece.kind == original {
			pieceString = content.original[start:end]
		}

		finalString = append(finalString, pieceString...)
		numPieces += 1
	}
	content.numPieces = numPieces

	return finalString
}
