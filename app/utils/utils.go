package utils

import "time"

type Entry struct {
	Value      string
	ExpiryTime time.Time
}

func SlicesEqual[T comparable](a, b []T) bool {
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
