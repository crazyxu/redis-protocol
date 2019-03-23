package parse

import (
	"fmt"
)

const (
	//\n
	LF = 10
	//\r
	CR = 13
	//$, bulk reply
	BR = 36
	//*, multi bulk reply
	MBR = 42
	//+ status reply
	SR = 43
	//- error reply
	ER = 45
	//:, integer reply
	IR = 58
	//0
	N0 = 48
	//9
	N9 = 57
)

func parse(s string) (result interface{}, err error) {
	bs := []byte(s)
	if len(bs) == 0 {
		return nil, fmt.Errorf("empty string")
	}

	//是否是multi bulk reply
	multi := bs[0] == MBR
	blucks := make([]interface{}, 0)

	//待读取的bulk reply条数
	multiBulkNum := 0

	//下两个字符必须为CRLF
	needCRLF := false

	defer func() {
		if err != nil {
			return
		}
		//最后也需要CRLF结束
		if needCRLF {
			err = fmt.Errorf("expect CRLF in the end")
		}
		if multiBulkNum != 0 {
			err = fmt.Errorf("%d bulk reply not meet", multiBulkNum)
		}
	}()

	length := len(bs)
	for i := 0; i < length; i++ {
		if needCRLF {
			if i+1 < length && bs[i] == CR && bs[i+1] == LF {
				i += 1
				needCRLF = false
				continue
			} else {
				err = fmt.Errorf("expect CRLF")
				return
			}
		}
		needCRLF = true

		switch b := bs[i]; b {
		case BR, MBR, IR:
			//接下来读取整数
			m := 0
			//可能为-1
			if i+2 < length && bs[i+1] == ER && bs[i+2] == N0+1 {
				m = -1
				i += 2
			} else {
				for j := i + 1; j < len(bs); j++ {
					n := bs[j]
					if n >= N0 && n <= N9 {
						m = m*10 + int(n-N0)
						i += 1
					} else {
						break
					}
				}
			}

			if b == MBR {
				//批回复
				multiBulkNum = m
			} else if b == BR {
				//单条回复，如果m为-1，相当于nil；否则在一个CRLF后，读取后续m个字符,
				var v interface{}
				if m == 0 {
					err = fmt.Errorf("not support $0")
					return
				} else if m == -1 {
					v = nil
				} else if length > i+m+2 {
					if bs[i+1] != CR || bs[i+2] != LF {
						err = fmt.Errorf("next of $ should be CRLF")
						return
					}
					v = string(bs[i+3 : i+m+3])
					i = i + m + 2
				} else {
					err = fmt.Errorf("length not enough")
					return
				}
				//批回复模式
				if multi {
					if multiBulkNum == 0 {
						err = fmt.Errorf("not support *0")
						return
					}
					blucks = append(blucks, v)
					multiBulkNum -= 1
				} else {
					return v, nil
				}

			} else {
				//回复整形
				if multi {
					if multiBulkNum == 0 {
						err = fmt.Errorf("*n not match")
						return
					}
					blucks = append(blucks, m)
					multiBulkNum -= 1
				} else {
					return m, nil
				}
			}
		case SR, ER:
			//如果是批回复模式，并且回复数量为0了
			if multi && multiBulkNum == 0 {
				err = fmt.Errorf("*n not match")
				return
			}

			//读取字符直到CRLF
			s := make([]byte, 0)
			for j := i + 1; j < length; j++ {
				if bs[j] == CR && j+1 < length && bs[j+1] == LF {
					break
				} else {
					s = append(s, bs[j])
				}
			}

			if multi {
				blucks = append(blucks, string(s))
				multiBulkNum -= 1
				i += len(s)
			} else {
				return string(s), nil
			}
		default:
			err = fmt.Errorf("unexpected char in position %d", i)
			return
		}
	}

	return blucks, nil
}
