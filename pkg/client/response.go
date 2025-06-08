package client

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
)

type ResponseParser struct {
	reader *bufio.Reader
}

func NewResponseParser(reader *bufio.Reader) *ResponseParser {
	return &ResponseParser{reader: reader}
}

func (p *ResponseParser) ReadResponse() (interface{}, error) {
	firstByte, err := p.reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read response type: %v", err)
	}

	switch firstByte {
	case '+': // Simple String
		line, err := p.readLine()
		if err != nil {
			return nil, err
		}
		return line, nil

	case '-': // Error
		line, err := p.readLine()
		if err != nil {
			return nil, err
		}
		return fmt.Errorf("%s", line), nil

	case ':': // Integer
		line, err := p.readLine()
		if err != nil {
			return nil, err
		}
		return p.parseInt(line)

	case '$': // Bulk String
		return p.readBulkString()

	case '*': // Array
		return p.readArray()

	default:
		return nil, fmt.Errorf("unknown response type: %c", firstByte)
	}
}

func (p *ResponseParser) readLine() (string, error) {
	line, err := p.reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	if len(line) < 2 || line[len(line)-2] != '\r' {
		return "", errors.New("invalid RESP syntax")
	}

	return line[:len(line)-2], nil
}

func (p *ResponseParser) parseInt(s string) (int, error) {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid integer: %v", err)
	}
	return n, nil
}

func (p *ResponseParser) readBulkString() (interface{}, error) {
	length, err := p.readLine()
	if err != nil {
		return nil, err
	}

	n, err := strconv.Atoi(length)
	if err != nil {
		return nil, fmt.Errorf("invalid bulk string length: %v", err)
	}

	if n < 0 {
		return nil, nil // Null bulk string
	}

	data := make([]byte, n+2) // +2 for CRLF
	_, err = io.ReadFull(p.reader, data)
	if err != nil {
		return nil, err
	}

	if data[n] != '\r' || data[n+1] != '\n' {
		return nil, errors.New("invalid RESP syntax")
	}

	return string(data[:n]), nil
}

func (p *ResponseParser) readArray() (interface{}, error) {
	length, err := p.readLine()
	if err != nil {
		return nil, err
	}

	n, err := strconv.Atoi(length)
	if err != nil {
		return nil, fmt.Errorf("invalid array length: %v", err)
	}

	if n < 0 {
		return nil, nil // Null array
	}

	array := make([]interface{}, n)
	for i := 0; i < n; i++ {
		element, err := p.ReadResponse()
		if err != nil {
			return nil, err
		}
		array[i] = element
	}

	return array, nil
}
