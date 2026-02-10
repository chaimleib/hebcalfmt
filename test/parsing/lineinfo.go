package parsing

type LineInfo struct {
	Line     []byte
	FileName string
	Number   int
}

func (li LineInfo) Position(col int) Position {
	return Position{
		Line:       li.Line,
		FileName:   li.FileName,
		LineNumber: li.Number,
		Col:        col,
	}
}
