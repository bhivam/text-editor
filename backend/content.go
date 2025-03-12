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
	Start  int
	Length int
	Kind   PieceType
	Next   *Piece
}

type Content struct {
	Original    []rune
	Add         []rune
	ContentRoot *Piece
	Length      int
	lastEdit    int64
	NumPieces   int
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

	content.Original = fileContent
	content.Add = []rune{}

	originalPiece := Piece{
		Start:  0,
		Length: len(fileContent) - 1,
		Kind:   original,
		Next:   nil,
	}

	content.Length = originalPiece.Length
	content.ContentRoot = &originalPiece
	content.NumPieces = 1
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
		Start:  len(content.Add),
		Length: len(r),
		Kind:   add,
		Next:   nil,
	}

	if len(r) > 0 {
		content.Add = append(content.Add, r...)
	}

	// CASE 1: no content
	if content.ContentRoot == nil {
		// assert start == 0 and end == 0
		content.ContentRoot = &newPiece
		content.Length = newPiece.Length
		return
	}

	// CASE 2: prepend
	if start == 0 && end == 0 {
		newPiece.Next = content.ContentRoot
		content.ContentRoot = &newPiece
		content.Length = content.Length + newPiece.Length
		return
	}

	// CASE 3: prepend
	if start == content.Length && end == content.Length {
		piece := content.ContentRoot
		for {
			if piece.Next == nil {
				break
			}
			piece = piece.Next
		}

		piece.Next = &newPiece
		content.Length = content.Length + newPiece.Length
		return
	}

	// CASE 4: inserting
	if start == end {
		pieceStart := 0
		for piece := content.ContentRoot; piece != nil; piece = piece.Next {
			pieceEnd := pieceStart + piece.Length

			if coalesce && piece.Kind == add &&
				piece.Start+piece.Length == newPiece.Start &&
				pieceEnd == start {

				piece.Length += newPiece.Length
				content.Length += newPiece.Length
				return
			}

			if pieceEnd == start {
				temp := piece.Next
				piece.Next = &newPiece
				newPiece.Next = temp
				content.Length = content.Length + newPiece.Length
				return
			} else if pieceStart < start && start < pieceEnd {
				pl := Piece{
					Start:  piece.Start,
					Length: start - pieceStart,
					Kind:   piece.Kind,
					Next:   nil,
				}
				pr := Piece{
					Start:  pl.Start + pl.Length,
					Length: piece.Length - pl.Length,
					Kind:   piece.Kind,
					Next:   nil,
				}

				temp := piece.Next
				piece.Start = pl.Start
				piece.Length = pl.Length
				piece.Next = &newPiece
				newPiece.Next = &pr
				pr.Next = temp

				content.Length = content.Length + newPiece.Length
				return
			}

			pieceStart = pieceEnd
		}
	}

	// Case 5: replacing
	// delete stuff first, recursion to insert
	pieceStart := 0
	var prev *Piece = nil
	for piece := content.ContentRoot; piece != nil; piece = piece.Next {

		pieceEnd := pieceStart + piece.Length

		if pieceStart < start && end < pieceEnd {
			// fmt.Println("going 1")
			content.Length -= piece.Length

			// set up right piece
			pr := &Piece{
				Start:  piece.Start + end - pieceStart,
				Length: pieceEnd - end,
				Kind:   piece.Kind,
				Next:   piece.Next,
			}

			// set up left piece
			piece.Length = start - pieceStart
			piece.Next = pr

			content.Length += piece.Length + pr.Length
			break
		}

		if pieceStart >= start && end >= pieceEnd {
			// fmt.Println("going 2")
			// fmt.Printf("%d %d %d %d\n", start, end, piece_start, piece_end)
			content.Length -= piece.Length
			if prev == nil {
				content.ContentRoot = content.ContentRoot.Next
			} else {
				prev.Next = piece.Next
			}
			piece = content.ContentRoot
			prev = nil

			if pieceEnd == end {
				break
			}

			continue

		} else if pieceStart >= start && end > pieceStart && end < pieceEnd {
			// fmt.Println("going 3")
			content.Length -= piece.Length
			piece.Start = piece.Start + end - start
			piece.Length = pieceEnd - end

			content.Length += piece.Length
			if pieceStart == start {
				break
			}

			pieceEnd = pieceStart + piece.Length
		} else if end >= pieceEnd && start > pieceStart && start < pieceEnd {
			// fmt.Println("going 4")
			content.Length -= piece.Length
			piece.Length = start - pieceStart

			content.Length += piece.Length
			if pieceEnd == end {
				break
			}

			pieceEnd = pieceStart + piece.Length
		}

		pieceStart = pieceEnd
		prev = piece
	}

	if len(r) > 0 {
		content.replace(r, start, start)
	}
}

func (content *Content) printPieces() {
	for piece := content.ContentRoot; piece != nil; piece = piece.Next {
		fmt.Println(piece)
	}
}

func (content *Content) calculateContent() []rune {
	finalString := []rune{}
	numPieces := 0

	for piece := content.ContentRoot; piece != nil; piece = piece.Next {
		var pieceString []rune = []rune{}

		// TODO Put this as a function of piece
		start := piece.Start
		end := piece.Start + piece.Length
		if piece.Kind == add {
			pieceString = content.Add[start:end]
		} else if piece.Kind == original {
			pieceString = content.Original[start:end]
		}

		finalString = append(finalString, pieceString...)
		numPieces += 1
	}
	content.NumPieces = numPieces

	return finalString
}
