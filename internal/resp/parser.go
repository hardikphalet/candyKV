package resp

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/hardikphalet/go-redis/internal/commands"
	"github.com/hardikphalet/go-redis/internal/commands/options"
	"github.com/hardikphalet/go-redis/internal/types"
)

var (
	ErrInvalidSyntax = errors.New("invalid RESP syntax")
)

type Parser struct {
	reader *bufio.Reader
}

func NewParser(reader *bufio.Reader) *Parser {
	return &Parser{reader: reader}
}

// Parse reads the RESP protocol input and returns a Command
func (p *Parser) Parse() (commands.Command, error) {
	firstByte, err := p.reader.ReadByte()
	if err != nil {
		return nil, err
	}

	switch firstByte {
	case '*':
		return p.parseArray()
	case '$':
		return nil, fmt.Errorf("bulk string must be part of an array")
	case '+':
		return nil, fmt.Errorf("simple string must be part of an array")
	case ':':
		return nil, fmt.Errorf("integer must be part of an array")
	case '-':
		return nil, fmt.Errorf("error must be part of an array")
	default:
		return nil, ErrInvalidSyntax
	}
}

// parseArray parses a RESP array
func (p *Parser) parseArray() (commands.Command, error) {
	length, err := p.readInteger()
	if err != nil {
		return nil, err
	}

	if length < 1 {
		return nil, fmt.Errorf("array length must be positive")
	}

	elements := make([]string, length)
	for i := 0; i < length; i++ {
		element, err := p.readBulkString()
		if err != nil {
			return nil, err
		}
		elements[i] = element
	}

	return p.createCommand(elements)
}

// readInteger reads a RESP integer
func (p *Parser) readInteger() (int, error) {
	line, err := p.readLine()
	if err != nil {
		return 0, err
	}

	n, err := strconv.Atoi(line)
	if err != nil {
		return 0, ErrInvalidSyntax
	}

	return n, nil
}

// readBulkString reads a RESP bulk string
func (p *Parser) readBulkString() (string, error) {
	b, err := p.reader.ReadByte()
	if err != nil {
		return "", err
	}
	if b != '$' {
		return "", ErrInvalidSyntax
	}

	length, err := p.readInteger()
	if err != nil {
		return "", err
	}

	if length < 0 {
		return "", nil // Null bulk string
	}

	data := make([]byte, length)
	_, err = io.ReadFull(p.reader, data)
	if err != nil {
		return "", err
	}

	if err := p.readCRLF(); err != nil {
		return "", err
	}

	return string(data), nil
}

// readLine reads a line ending in CRLF
func (p *Parser) readLine() (string, error) {
	line, err := p.reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	if len(line) < 2 || line[len(line)-2] != '\r' {
		return "", ErrInvalidSyntax
	}

	return line[:len(line)-2], nil
}

// readCRLF reads and verifies a CRLF sequence
func (p *Parser) readCRLF() error {
	cr, err := p.reader.ReadByte()
	if err != nil {
		return err
	}
	if cr != '\r' {
		return ErrInvalidSyntax
	}

	lf, err := p.reader.ReadByte()
	if err != nil {
		return err
	}
	if lf != '\n' {
		return ErrInvalidSyntax
	}

	return nil
}

// createCommand converts string array to a specific command
func (p *Parser) createCommand(args []string) (commands.Command, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("empty command")
	}

	cmd := strings.ToUpper(args[0])

	switch cmd {
	case "PING":
		return &commands.PingCommand{}, nil
	case "ECHO":
		return commands.NewEchoCommand(args)
	case "SET":
		if len(args) < 3 {
			return nil, fmt.Errorf("SET command requires at least 2 arguments")
		}

		opts := options.NewSetOptions()

		i := 3
		for i < len(args) {
			opt := strings.ToUpper(args[i])
			switch opt {
			case "NX", "XX", "GET":
				if err := opts.Set(opt); err != nil {
					return nil, fmt.Errorf("invalid option: %s", err)
				}
				i++
			case "EX", "PX", "EXAT", "PXAT", "KEEPTTL":
				if opt == "KEEPTTL" {
					if err := opts.SetExpiry(opt, 0); err != nil {
						return nil, fmt.Errorf("invalid option: %s", err)
					}
					i++
				} else {
					if i+1 >= len(args) {
						return nil, fmt.Errorf("option %s requires a value", opt)
					}
					val, err := strconv.ParseInt(args[i+1], 10, 64)
					if err != nil {
						return nil, fmt.Errorf("invalid value for option %s: %s", opt, err)
					}
					if err := opts.SetExpiry(opt, val); err != nil {
						return nil, fmt.Errorf("invalid option: %s", err)
					}
					i += 2
				}
			default:
				return nil, fmt.Errorf("unknown option: %s", opt)
			}
		}

		return commands.NewSetCommand(args[1], args[2], opts), nil

	case "GET":
		if len(args) != 2 {
			return nil, fmt.Errorf("GET command requires exactly 1 argument")
		}
		return &commands.GetCommand{
			Key: args[1],
		}, nil

	case "DEL":
		if len(args) < 2 {
			return nil, fmt.Errorf("DEL command requires at least 1 argument")
		}
		return &commands.DelCommand{
			Keys: args[1:],
		}, nil

	case "EXPIRE":
		if len(args) < 3 {
			return nil, fmt.Errorf("EXPIRE command requires at least 2 arguments")
		}
		ttl, err := strconv.Atoi(args[2])
		if err != nil {
			return nil, fmt.Errorf("invalid TTL value")
		}

		opts := options.NewExpireOptions()

		for i := 3; i < len(args); i++ {
			opt := strings.ToUpper(args[i])
			if err := opts.Set(opt); err != nil {
				return nil, fmt.Errorf("invalid option: %s", err)
			}
		}

		return &commands.ExpireCommand{
			Key:     args[1],
			TTL:     time.Duration(ttl) * time.Second,
			Options: opts,
		}, nil

	case "TTL":
		if len(args) != 2 {
			return nil, fmt.Errorf("TTL command requires exactly 1 argument")
		}
		return &commands.TtlCommand{
			Key: args[1],
		}, nil

	case "KEYS":
		if len(args) != 2 {
			return nil, fmt.Errorf("KEYS command requires exactly 1 argument")
		}
		return &commands.KeysCommand{
			Pattern: args[1],
		}, nil

	case "ZADD":
		if len(args) < 4 || (len(args)-2)%2 != 0 {
			return nil, fmt.Errorf("ZADD command requires at least one score-member pair")
		}

		opts := options.NewZAddOptions()

		i := 1
	optionLoop:
		for i < len(args) {
			opt := strings.ToUpper(args[i])
			switch opt {
			case "NX", "XX", "GT", "LT", "CH", "INCR":
				if err := opts.Set(opt); err != nil {
					return nil, fmt.Errorf("invalid option: %s", err)
				}
				i++
			default:
				break optionLoop
			}
		}

		i++

		members := make([]types.ScoreMember, 0, (len(args)-i)/2)
		for i < len(args) {
			score, err := strconv.ParseFloat(args[i], 64)
			if err != nil {
				return nil, fmt.Errorf("invalid score value: %s", args[i])
			}
			members = append(members, types.ScoreMember{
				Score:  score,
				Member: args[i+1],
			})
			i += 2
		}

		return &commands.ZAddCommand{
			Key:     args[1],
			Members: members,
			Options: opts,
		}, nil

	case "ZRANGE":
		if len(args) < 4 {
			return nil, fmt.Errorf("ZRANGE command requires at least 3 arguments")
		}

		opts := options.NewZRangeOptions()

		i := 4 // Start after key, start, stop
		for i < len(args) {
			opt := strings.ToUpper(args[i])
			switch opt {
			case "BYSCORE", "BYLEX":
				if err := opts.SetRangeType(opt); err != nil {
					return nil, fmt.Errorf("invalid range type: %s", err)
				}
				i++
			case "REV":
				opts.Rev = true
				i++
			case "WITHSCORES":
				opts.WithScores = true
				i++
			case "LIMIT":
				if i+2 >= len(args) {
					return nil, fmt.Errorf("LIMIT option requires offset and count")
				}
				offset, err := strconv.Atoi(args[i+1])
				if err != nil {
					return nil, fmt.Errorf("invalid LIMIT offset")
				}
				count, err := strconv.Atoi(args[i+2])
				if err != nil {
					return nil, fmt.Errorf("invalid LIMIT count")
				}
				if err := opts.SetLimit(offset, count); err != nil {
					return nil, fmt.Errorf("invalid LIMIT parameters: %s", err)
				}
				i += 3
			default:
				return nil, fmt.Errorf("unknown option: %s", opt)
			}
		}

		var start, stop interface{}
		var err error

		if opts.IsByScore() {
			start, err = strconv.ParseFloat(args[2], 64)
			if err != nil {
				return nil, fmt.Errorf("invalid score range start: %s", err)
			}
			stop, err = strconv.ParseFloat(args[3], 64)
			if err != nil {
				return nil, fmt.Errorf("invalid score range stop: %s", err)
			}
		} else if opts.IsByLex() {
			start = args[2]
			stop = args[3]
		} else {
			start, err = strconv.Atoi(args[2])
			if err != nil {
				return nil, fmt.Errorf("invalid start index")
			}
			stop, err = strconv.Atoi(args[3])
			if err != nil {
				return nil, fmt.Errorf("invalid stop index")
			}
		}

		return &commands.ZRangeCommand{
			Key:     args[1],
			Start:   start,
			Stop:    stop,
			Options: opts,
		}, nil

	case "COMMAND":
		return &commands.CommandCommand{}, nil

	default:
		return nil, fmt.Errorf("unknown command: %s", cmd)
	}
}
