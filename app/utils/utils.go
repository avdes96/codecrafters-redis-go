package utils

import (
	"fmt"
	"math/rand"
	"strings"
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
	MasterAddress string
}

const replicationIdLen int = 40

func NewReplicationInfo(masterAddress string) *ReplicationInfo {
	role := REPLICA
	formattedAddress := formatAddress(masterAddress)
	if formattedAddress == "" {
		role = MASTER
	}
	return &ReplicationInfo{
		Role:          role,
		ReplicationId: randomAlphanumericString(replicationIdLen),
		Offset:        0,
		MasterAddress: formattedAddress,
	}
}

func formatAddress(a string) string {
	parts := strings.Split(a, " ")
	if len(parts) != 2 {
		return ""
	}
	return fmt.Sprintf("%s:%s", parts[0], parts[1])
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
