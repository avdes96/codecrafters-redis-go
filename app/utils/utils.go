package utils

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"strings"
	"time"
)

type Entry struct {
	Value      string
	ExpiryTime time.Time
}

type Command struct {
	CMD     string
	ARGS    []string
	ByteLen int
}

func (c *Command) GetCMD() string {
	return strings.ToLower(c.CMD)
}

func FormatAddress(a string) string {
	parts := strings.Split(a, " ")
	if len(parts) != 2 {
		return ""
	}
	return fmt.Sprintf("%s:%s", parts[0], parts[1])
}

const options string = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func RandomAlphanumericString(n int) string {
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

func WriteToConnection(conn net.Conn, b []byte) {
	if _, err := conn.Write(b); err != nil {
		log.Printf("Error writing to connection %s", err.Error())
	}
}
