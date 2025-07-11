package protocol

import (
	"strconv"
)

func ToArrayBulkStrings(strs []string) []byte {
	ret := []byte{}
	ret = append(ret, '*')
	ret = append(ret, []byte(strconv.Itoa(len(strs)))...)
	ret = appendCrlf(ret)
	for _, s := range strs {
		ret = append(ret, ToBulkString(s)...)
	}
	return ret
}

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

func OkResp() []byte {
	return []byte("+OK\r\n")
}

func NullBulkString() []byte {
	return []byte("$-1\r\n")
}
