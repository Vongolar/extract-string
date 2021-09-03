package extract

import "io"

type OnFindStr func(str []byte, startIndex int, endIndex int, ext ...interface{}) bool

type Extractor interface {
	Extract(w io.WriteSeeker, onFindStr OnFindStr)
}
