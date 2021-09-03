package extract

import (
	"fmt"
	"os"
	"testing"
)

func Test_FindInCSharp(t *testing.T) {
	path := "./test/test.cs"
	f, err := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
	if err != nil {
		t.Error(err)
	}
	defer f.Close()

	e := new(ExtractorCSharp)
	e.Extract(f, func(str []byte, startIndex, endIndex int, ext ...interface{}) bool {
		fmt.Println(string(str))
		return true
	})
}
