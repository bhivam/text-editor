package backend

type Cursor struct {
	Index     int  `json:"index"`
	Row       int  `json:"row"`
	Col       int  `json:"col"`
	AtNewLine bool `json:"atNewLine"`
}
