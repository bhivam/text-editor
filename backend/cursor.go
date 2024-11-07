package backend

type Cursor struct {
	index     int
	row       int
	col       int
	atNewLine bool
}
