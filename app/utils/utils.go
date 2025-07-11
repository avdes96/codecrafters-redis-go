package utils

import (
	"math/rand"
	"time"
)

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
	Role          role
	ReplicationId string
	Offset        int
}

const replicationIdLen int = 40

func NewReplicationInfo(r role) *ReplicationInfo {
	return &ReplicationInfo{
		Role:          r,
		ReplicationId: randomAlphanumericString(replicationIdLen),
		Offset:        0,
	}
}

const options string = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func randomAlphanumericString(n int) string {
	gen := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, n)
	for i := range n {
		b[i] = options[gen.Intn(len(options))]
	}
	return string(b)
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
