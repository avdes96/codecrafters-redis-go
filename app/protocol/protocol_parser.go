package protocol

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/utils"
)

type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
}

const crlf string = "\r\n"

func (p *Parser) Parse(r io.Reader, commandChan chan utils.Command, replicaResponseChan chan string) {
	defer func() {
		close(commandChan)
		close(replicaResponseChan)
	}()
	reader := bufio.NewReader(r)
	for {
		prefix, err := reader.Peek(1)
		if err == io.EOF {
			return
		} else if err != nil {
			log.Printf("peek error: %v", err)
			return
		}

		var cmd utils.Command
		var str string
		switch prefix[0] {
		case '*':
			cmd, err = p.parseArrayBulkStrings(reader)
		case '+':
			str, numBytes, err := p.parseSimpleString(reader)
			if err != nil {
				log.Printf("Error in parsing command: %s", err)
			}
			if isReplicaResp(str) {
				replicaResponseChan <- str
				continue
			}
			commandChan <- utils.Command{CMD: str, ByteLen: numBytes}

		case '$':
			str, _, err = p.parseBulkString(reader)
			if err != nil {
				log.Printf("Error in parsing command: %s", err)
				return
			}
			if isRdbFile(str) {
				replicaResponseChan <- "rdbFileReceived"
			}
			continue
		default:
			err = fmt.Errorf("command does not start with valid char: %b", prefix[0])
			if _, err := reader.Discard(1); err != nil {
				return
			}
		}
		if err != nil {
			log.Printf("Error in parsing command: %s", err)
			return
		}
		commandChan <- cmd
	}
}

func isReplicaResp(str string) bool {
	respStart := strings.Split(str, " ")[0]
	lower := strings.ToLower(respStart)
	resps := map[string]struct{}{
		"pong":       struct{}{},
		"ok":         struct{}{},
		"fullresync": struct{}{},
	}
	if _, ok := resps[lower]; ok {
		return true
	}
	return false
}

func isRdbFile(str string) bool {
	return strings.HasPrefix(str, "REDIS")
}

func (p *Parser) parseSimpleString(r *bufio.Reader) (string, int, error) {
	line, err := r.ReadString('\n')
	if err != nil {
		return "", -1, err
	}
	if line[0] != '+' {
		return "", 1, fmt.Errorf("simple string doesn't start with +: %s", line)
	}
	return strings.TrimSuffix(line[1:], crlf), len(line), nil
}

func (p *Parser) parseArrayBulkStrings(r *bufio.Reader) (utils.Command, error) {
	numBytes := 0
	header, err := r.ReadString('\n')
	if err != nil {
		return utils.Command{}, err
	}
	if header[0] != '*' {
		return utils.Command{}, fmt.Errorf("array of bulk strs doesn't start with *: %s", header)
	}
	numBytes += len(header)
	arraySize, err := strconv.Atoi(strings.TrimSuffix(header[1:], crlf))
	if err != nil {
		return utils.Command{}, err
	}
	arr := make([]string, arraySize)
	for arrIdx := 0; arrIdx < arraySize; arrIdx++ {
		str, strBytes, err := p.parseBulkString(r)
		numBytes += strBytes
		if err != nil {
			return utils.Command{}, err
		}
		arr[arrIdx] = str
	}
	return utils.Command{CMD: arr[0], ARGS: arr[1:], ByteLen: numBytes}, nil
}

func (p *Parser) parseBulkString(r *bufio.Reader) (string, int, error) {
	numBytes := 0
	stringHeader, err := r.ReadString('\n')
	if err != nil {
		return "", -1, err
	}
	if stringHeader[0] != '$' {
		return "", -1, fmt.Errorf("str doesn't start with *: %s", stringHeader)
	}
	numBytes += len(stringHeader)
	strLen, err := strconv.Atoi(strings.TrimSuffix(stringHeader[1:], crlf))
	if err != nil {
		return "", -1, err
	}
	strBuffer := make([]byte, strLen)
	if _, err := io.ReadFull(r, strBuffer); err != nil {
		return "", -1, err
	}
	numBytes += strLen

	str := string(strBuffer)
	buf, err := r.Peek(2)
	if err != nil {
		return "", -1, fmt.Errorf("unable to peek: %w", err)
	}

	if buf[0] == '\r' && buf[1] == '\n' {
		numBytes += 2
		if _, err := r.Discard(2); err != nil {
			return "", -1, err
		}
	}
	return str, numBytes, nil
}
