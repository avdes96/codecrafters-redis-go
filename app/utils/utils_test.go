package utils

import (
	"strconv"
	"testing"
)

func TestToArrayBulkStrings(t *testing.T) {
	tests := []struct {
		input    []string
		expected []byte
	}{
		{[]string{"one", "two", "three"}, []byte("*3\r\n$3\r\none\r\n$3\r\ntwo\r\n$5\r\nthree\r\n")},
	}

	for _, tt := range tests {
		got := ToArrayBulkStrings(tt.input)
		if !sliceEqual(got, tt.expected) {
			t.Errorf("Expected: %s, but got %s", strconv.Quote(string(tt.expected)), strconv.Quote(string(got)))
		}
	}

}

func TestTBulkString(t *testing.T) {
	tests := []struct {
		input    string
		expected []byte
	}{
		{"hello", []byte("$5\r\nhello\r\n")},
		{"onetwothreefour", []byte("$15\r\nonetwothreefour\r\n")},
	}

	for _, tt := range tests {
		got := ToBulkString(tt.input)
		if !sliceEqual(got, tt.expected) {
			t.Errorf("Expected: %s, but got %s", strconv.Quote(string(tt.expected)), strconv.Quote(string(got)))
		}
	}

}

func sliceEqual(a []byte, b []byte) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
