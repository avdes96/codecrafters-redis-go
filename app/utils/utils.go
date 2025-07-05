package utils

import "strconv"

func ToBulkString(s string) []byte {
	ret := []byte{}
	ret = append(ret, '$')
	ret = append(ret, []byte(strconv.Itoa(len(s)))...)
	ret = appendCrlf(ret)
	for _, c := range s {
		ret = append(ret, byte(c))
	}
	ret = appendCrlf(ret)
	return ret
}

func appendCrlf(b []byte) []byte {
	b = append(b, '\r')
	b = append(b, '\n')
	return b
}
