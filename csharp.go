package extract

import "io"

type ExtractorCSharp struct {
	annotation      bool //是否在注释内
	mutliAnnotation bool //是否在多行注释内
	isStr           bool
	isAbsoteStr     bool
	isFormat        bool

	lastByte byte
	findStr  []byte

	curIndex   int
	startIndex int
}

func (extractor *ExtractorCSharp) Extract(reader io.ReadSeeker, onFindStr OnFindStr) error {
	extractor.annotation = false
	extractor.mutliAnnotation = false
	extractor.isStr = false
	extractor.isAbsoteStr = false

	var err error

	cache := make([]byte, 1)
	_, err = reader.Read(cache)

	next := make([]byte, 1)

	extractor.curIndex = 1

	for err == nil {

		if extractor.annotation {
			// 在注释内
			if extractor.mutliAnnotation {
				if cache[0] == '/' && extractor.lastByte == '*' {
					// 多行注释结束
					extractor.endAnnotation()
				}
			} else {
				if cache[0] == '\n' {
					// 单行注释结束
					extractor.endAnnotation()
				}
			}

		} else if extractor.isStr {
			// 在字符串内
			if cache[0] != '"' {
				// 仍然在字符串内
			} else if extractor.isAbsoteStr {

				//要保证下一位不是 '"'
				_, nextErr := reader.Read(next)
				if nextErr == io.EOF {
					next[0] = 0
				} else if nextErr != nil {
					return nextErr
				}

				//回退一位
				_, nextErr = reader.Seek(-1, io.SeekCurrent)
				if nextErr != nil {
					return nextErr
				}

				if next[0] == '"' || extractor.lastByte == '"' && !isEvenEnd(extractor.findStr, '"') {
					// 仍然在字符串内
				} else {
					extractor.isStr = false
					extractor.isAbsoteStr = false
				}
			} else if extractor.isFormat {
				// 是format$ 语法糖，并且不在{}内，只检查{}成对出现，不检查语法
				n := 0
				for _, b := range extractor.findStr {
					if b == '{' {
						n += 1
					} else if b == '}' {
						n -= 1
					}
				}
				if n == 0 {
					extractor.isStr = false
					extractor.isFormat = false
				}

			} else {
				if extractor.lastByte == '\\' && !isEvenEnd(extractor.findStr, '\\') {
					// 仍然在字符串内
				} else {
					extractor.isStr = false
				}
			}

			if extractor.isStr {
				extractor.findStr = append(extractor.findStr, cache[0])
			} else if len(extractor.findStr) > 0 {
				res := make([]byte, len(extractor.findStr))
				for i, b := range extractor.findStr {
					res[i] = b
				}
				extractor.findStr = nil
				onFindStr(res, extractor.startIndex, extractor.curIndex-1)
			}

		} else { //不在注释、字符串中

			if cache[0] == '/' && extractor.lastByte == '/' {
				//单行注释开始
				extractor.annotation = true
			} else if cache[0] == '*' && extractor.lastByte == '/' {
				//多行注释开始
				extractor.annotation = true
				extractor.mutliAnnotation = true
			} else if cache[0] == '"' && extractor.lastByte == '@' {
				//多行字符串开始
				extractor.isStr = true
				extractor.isAbsoteStr = true
				extractor.startIndex = extractor.curIndex
				extractor.findStr = make([]byte, 0)
			} else if cache[0] == '"' && extractor.lastByte == '$' {
				// format 语法糖 字符串开始
				extractor.isStr = true
				extractor.isFormat = true
				extractor.startIndex = extractor.curIndex
				extractor.findStr = make([]byte, 0)
			} else if cache[0] == '"' {
				//单行字符串开始
				extractor.isStr = true
				extractor.startIndex = extractor.curIndex
				extractor.findStr = make([]byte, 0)
			}

		}

		extractor.lastByte = cache[0]
		_, err = reader.Read(cache)
		extractor.curIndex++
	}

	if err == io.EOF {
		return nil
	}

	return err
}

func (extractor *ExtractorCSharp) endAnnotation() {
	extractor.annotation, extractor.mutliAnnotation = false, false
}

// bs 结尾相邻 b 是偶数个
func isEvenEnd(bs []byte, b byte) bool {
	cnt := 0

	for i := len(bs) - 1; i >= 0; i-- {
		if bs[i] == b {
			cnt++
		} else {
			break
		}
	}

	return cnt%2 == 0
}
