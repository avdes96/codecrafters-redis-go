package utils

import "time"

type Entry struct {
	Value      string
	ExpiryTime time.Time
}

type role int

const (
	MASTER role = iota
	REPLICA
)

func (r role) String() string {
	// "slave" included as this version of codecrafters does not use updated renaming of "slave" to "replica"
	return [...]string{"master", "slave"}[r]
}

type ReplicationInfo struct {
	Role role
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
