package parser

import (
	"strings"
	"testing"

	"github.com/codecrafters-io/redis-starter-go/app/command"
)

func TestParser(t *testing.T) {
	tests := []struct {
		input       []byte
		expectedCmd command.Command
		errorStr    string
	}{
		{[]byte("*2\r\n$4\r\nECHO\r\n$3\r\nhey\r\n"), command.Command{CMD: "echo", ARGS: []string{"hey"}}, ""},
		{[]byte("*2\n$4\r\nECHO\r\n$3\r\nhey\r\n"), command.Command{CMD: "", ARGS: []string{}}, "int not followed by crlf"},
		{[]byte("*2\r\n$4\r\nECHO\n$3\r\nhey\r\n"), command.Command{CMD: "", ARGS: []string{}}, "string not followed by crlf"},
		{[]byte("*2\r\n4\r\nECHO\r\n3\r\nhey\r\n"), command.Command{CMD: "", ARGS: []string{}}, "string does not start with $"},
		{[]byte("2\r\n$4\r\nECHO\r\n3\r\nhey\r\n"), command.Command{CMD: "", ARGS: []string{}}, "command does not start with valid char"},
		{[]byte("+PING\r\n"), command.Command{CMD: "ping", ARGS: []string{}}, ""},
		{[]byte("+PING"), command.Command{CMD: "", ARGS: []string{}}, "simple string does not end with crlf"},
	}

	p := NewParser()

	for _, tt := range tests {
		func() {
			got, err := p.ParseInputToCommand(tt.input)
			if tt.errorStr == "" && err != nil {
				t.Errorf("Error during parsing: %s", err)
				return
			}
			if tt.errorStr != "" && err == nil {
				t.Errorf("Expected error: %s, but got none", tt.errorStr)
				return
			}
			if tt.errorStr != "" {
				if !strings.Contains(err.Error(), tt.errorStr) {
					t.Errorf("unexpected error; got %s; want %s", err.Error(), tt.errorStr)
				}
				return
			}
			if !commandEqual(got, tt.expectedCmd) {
				t.Errorf("got %s; want %s", got, tt.expectedCmd)
				return
			}
		}()
	}
}

func commandEqual(got command.Command, expected command.Command) bool {
	if got.CMD != expected.CMD {
		return false
	}
	if len(got.ARGS) != len(expected.ARGS) {
		return false
	}
	for i := range got.ARGS {
		if got.ARGS[i] != expected.ARGS[i] {
			return false
		}
	}
	return true
}
