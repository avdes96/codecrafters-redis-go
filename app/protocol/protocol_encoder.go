package protocol

import (
	"fmt"
	"strconv"
)

func ToSimpleString(str string) []byte {
	return []byte(fmt.Sprintf("+%s\r\n", str))
}

func ToError(str string) []byte {
	return []byte(fmt.Sprintf("-%s\r\n", str))
}

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

func ToRespInt(i int) []byte {
	respInt := ":" + strconv.Itoa(i) + crlf
	return []byte(respInt)
}

func appendCrlf(b []byte) []byte {
	b = append(b, '\r')
	b = append(b, '\n')
	return b
}

func OkResp() []byte {
	return ToSimpleString("OK")
}

func NullBulkString() []byte {
	return []byte("$-1\r\n")
}

func CommandAndArgsToBulkString(cmd string, args []string) []byte {
	s := []string{cmd}
	s = append(s, args...)
	return ToArrayBulkStrings(s)
}
