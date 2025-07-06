package protocol

import (
	"strings"
	"testing"

	"github.com/codecrafters-io/redis-starter-go/app/utils"
)

func TestParser(t *testing.T) {
	tests := []struct {
		input        []byte
		expectedCmd  string
		expectedArgs []string
		errorStr     string
	}{
		{[]byte("*2\r\n$4\r\nECHO\r\n$3\r\nhey\r\n"), "echo", []string{"hey"}, ""},
		{[]byte("*2\n$4\r\nECHO\r\n$3\r\nhey\r\n"), "", []string{}, "int not followed by crlf"},
		{[]byte("*2\r\n$4\r\nECHO\n$3\r\nhey\r\n"), "", []string{}, "string not followed by crlf"},
		{[]byte("*2\r\n4\r\nECHO\r\n3\r\nhey\r\n"), "", []string{}, "string does not start with $"},
		{[]byte("2\r\n$4\r\nECHO\r\n3\r\nhey\r\n"), "", []string{}, "command does not start with valid char"},
		{[]byte("+PING\r\n"), "ping", []string{}, ""},
		{[]byte("+PING"), "", []string{}, "simple string does not end with crlf"},
	}

	p := NewParser()

	for _, tt := range tests {
		func() {
			gotCmd, gotArgs, err := p.ParseInputToCommandAndArgs(tt.input)
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
			if gotCmd != tt.expectedCmd {
				t.Errorf("got %s; want %s", gotCmd, tt.expectedCmd)
				return
			}

			if !utils.SlicesEqual(gotArgs, tt.expectedArgs) {
				t.Errorf("got %s; want %s", gotArgs, tt.expectedArgs)
			}

		}()
	}
}
