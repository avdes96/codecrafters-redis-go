package protocol

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/utils"
)

type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) Parse(b []byte, commandChan chan utils.Command) {
	defer close(commandChan)
	idx := 0
	var cmd string
	var args []string
	var err error
	for idx < len(b) {
		cmd, args, idx, err = p.ParseInputToCommandAndArgs(b, idx)
		if err != nil {
			log.Printf("Error parsing input: %s", err)
			continue
		}
		commandChan <- utils.Command{CMD: cmd, ARGS: args}
	}
}

func (p *Parser) ParseInputToCommandAndArgs(b []byte, startIdx int) (string, []string, int, error) {
	switch b[startIdx] {
	case '*':
		strs, idx, err := p.parseArrayBulkStrings(b, startIdx)
		if err != nil {
			return "", []string{}, -1, err
		}
		return strings.ToLower(strs[0]), strs[1:], idx, nil
	case '+':
		str, idx, err := p.parseSimpleString(b, startIdx)
		if err != nil {
			return "", []string{}, -1, err
		}
		return strings.ToLower(str), []string{}, idx, nil
	default:
		return "", []string{}, -1, fmt.Errorf("command does not start with valid char: %s", b)
	}
}

func (p *Parser) parseSimpleString(b []byte, startIdx int) (string, int, error) {
	if b[startIdx] != '+' {
		return "", -1, fmt.Errorf("simple string doesn't start with +: %s", b)
	}
	str := ""
	i := startIdx + 1
	for i < len(b) {
		if b[i] == '\n' && i > 0 && b[i-1] == '\r' {
			return str, i + 1, nil
		}
		str += string(b[i])
		i++
	}
	return "", -1, fmt.Errorf("simple string does not end with crlf: %s", b)
}

func (p *Parser) parseArrayBulkStrings(b []byte, startIdx int) ([]string, int, error) {
	if b[startIdx] != '*' {
		return []string{}, -1, fmt.Errorf("array of bulk strs doesn't start with *: %s", b)
	}
	i := startIdx + 1
	arraySize, i, err := p.parseIntFromByteArray(b, i)
	if err != nil {
		return []string{}, -1, err
	}
	arr := make([]string, arraySize)
	for arrIdx := range arraySize {
		s, newIdx, err := p.parseStringFromArray(b, i)
		if err != nil {
			return []string{}, -1, err
		}
		arr[arrIdx] = s
		i = newIdx
	}
	return arr, i, nil
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
