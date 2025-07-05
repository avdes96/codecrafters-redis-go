package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/command"
)

type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) ParseInputToCommand(b []byte) (command.Command, error) {
	switch b[0] {
	case '*':
		strs, err := p.parseArrayBulkStrings(b)
		if err != nil {
			return command.Command{CMD: "", ARGS: []string{}}, err
		}
		return command.Command{CMD: strings.ToLower(strs[0]), ARGS: strs[1:]}, nil
	case '+':
		str, err := p.parseSimpleString(b)
		if err != nil {
			return command.Command{CMD: "", ARGS: []string{}}, err
		}
		return command.Command{CMD: strings.ToLower(str), ARGS: []string{}}, err
	default:
		return command.Command{CMD: "", ARGS: []string{}}, fmt.Errorf("command does not start with valid char: %s", b)
	}
}

func (p *Parser) parseSimpleString(b []byte) (string, error) {
	if b[0] != '+' {
		return "", fmt.Errorf("simple string doesn't start with +: %s", b)
	}
	if !(len(b) >= 3 && b[len(b)-2] == '\r' && b[len(b)-1] == '\n') {
		return "", fmt.Errorf("simple string does not end with crlf: %s", b)
	}
	return string(b[1 : len(b)-2]), nil
}

func (p *Parser) parseArrayBulkStrings(b []byte) ([]string, error) {
	if b[0] != '*' {
		return []string{}, fmt.Errorf("array of bulk strs doesn't start with *: %s", b)
	}
	i := 1
	arraySize, i, err := p.parseIntFromByteArray(b, i)
	if err != nil {
		return []string{}, err
	}
	arr := make([]string, arraySize)
	arrIdx := 0
	for i < len(b) {
		s, newIdx, err := p.parseStringFromArray(b, i)
		if err != nil {
			return []string{}, err
		}
		arr[arrIdx] = s
		arrIdx++
		i = newIdx
	}
	fmt.Println(arr, len(arr))
	return arr, nil
}

func (p *Parser) parseStringFromArray(b []byte, idx int) (string, int, error) {
	if b[idx] != '$' {
		return "", -1, fmt.Errorf("string does not start with $: %s", b)
	}
	strLen, idx, err := p.parseIntFromByteArray(b, idx+1)
	if err != nil {
		return "", -1, err
	}
	str := ""
	for range strLen {
		str += string(b[idx])
		idx++
	}
	if (idx+1 >= len(b)) || b[idx] != '\r' || b[idx+1] != '\n' {
		return "", -1, fmt.Errorf("string not followed by crlf %s", b)
	}
	return str, idx + 2, nil
}

func (p *Parser) parseIntFromByteArray(b []byte, idx int) (int, int, error) {
	intString := ""
	for {
		if idx >= len(b) {
			return -1, -1, fmt.Errorf("int not followed by crlf %s", b)
		}
		if !(b[idx] >= '0' && b[idx] <= '9') {
			break
		}
		intString += string(b[idx])
		idx++
	}
	parsedInt, err := strconv.Atoi(intString)
	if err != nil {
		return -1, -1, err
	}
	if (idx+1 >= len(b)) || b[idx] != '\r' || b[idx+1] != '\n' {
		return -1, -1, fmt.Errorf("int not followed by crlf %s", b)
	}
	return parsedInt, idx + 2, nil
}
