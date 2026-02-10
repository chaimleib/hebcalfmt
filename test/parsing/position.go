package parsing

type Position struct {
	Line       []byte
	FileName   string
	LineNumber int
	Col        int
}

func (p Position) LineInfo() LineInfo {
	return LineInfo{
		Line:     p.Line,
		FileName: p.FileName,
		Number:   p.LineNumber,
	}
}
