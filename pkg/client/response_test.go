package client

import (
	"bufio"
	"strings"
	"testing"
)

func createParserWithInput(input string) *ResponseParser {
	reader := bufio.NewReader(strings.NewReader(input))
	return NewResponseParser(reader)
}

func TestSimpleString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "valid simple string",
			input:    "+OK\r\n",
			expected: "OK",
			wantErr:  false,
		},
		{
			name:     "simple string with spaces",
			input:    "+Hello World\r\n",
			expected: "Hello World",
			wantErr:  false,
		},
		{
			name:    "invalid format - no CRLF",
			input:   "+OK\n",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := createParserWithInput(tt.input)
			result, err := parser.ReadResponse()

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if str, ok := result.(string); !ok || str != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, str)
			}
		})
	}
}

func TestError(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError string
		wantErr       bool
	}{
		{
			name:          "standard error",
			input:         "-Error message\r\n",
			expectedError: "Error message",
			wantErr:       false,
		},
		{
			name:          "error with special characters",
			input:         "-ERR unknown command 'FOO'\r\n",
			expectedError: "ERR unknown command 'FOO'",
			wantErr:       false,
		},
		{
			name:    "invalid format - no CRLF",
			input:   "-Error\n",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := createParserWithInput(tt.input)
			result, err := parser.ReadResponse()

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if errResult, ok := result.(error); !ok {
				t.Error("expected error type result")
			} else if errResult.Error() != tt.expectedError {
				t.Errorf("expected error %q, got %q", tt.expectedError, errResult.Error())
			}
		})
	}
}

func TestInteger(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
		wantErr  bool
	}{
		{
			name:     "positive integer",
			input:    ":1000\r\n",
			expected: 1000,
			wantErr:  false,
		},
		{
			name:     "negative integer",
			input:    ":-123\r\n",
			expected: -123,
			wantErr:  false,
		},
		{
			name:     "zero",
			input:    ":0\r\n",
			expected: 0,
			wantErr:  false,
		},
		{
			name:    "invalid integer",
			input:   ":abc\r\n",
			wantErr: true,
		},
		{
			name:    "invalid format - no CRLF",
			input:   ":123\n",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := createParserWithInput(tt.input)
			result, err := parser.ReadResponse()

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if num, ok := result.(int); !ok || num != tt.expected {
				t.Errorf("expected %d, got %v", tt.expected, result)
			}
		})
	}
}

func TestBulkString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected interface{}
		wantErr  bool
	}{
		{
			name:     "normal bulk string",
			input:    "$5\r\nhello\r\n",
			expected: "hello",
			wantErr:  false,
		},
		{
			name:     "empty bulk string",
			input:    "$0\r\n\r\n",
			expected: "",
			wantErr:  false,
		},
		{
			name:     "null bulk string",
			input:    "$-1\r\n",
			expected: nil,
			wantErr:  false,
		},
		{
			name:    "invalid length",
			input:   "$abc\r\n",
			wantErr: true,
		},
		{
			name:    "length mismatch",
			input:   "$5\r\nhi\r\n",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := createParserWithInput(tt.input)
			result, err := parser.ReadResponse()

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestArray(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected interface{}
		wantErr  bool
	}{
		{
			name:     "simple array",
			input:    "*2\r\n$5\r\nhello\r\n$5\r\nworld\r\n",
			expected: []interface{}{"hello", "world"},
			wantErr:  false,
		},
		{
			name:     "empty array",
			input:    "*0\r\n",
			expected: []interface{}{},
			wantErr:  false,
		},
		{
			name:     "null array",
			input:    "*-1\r\n",
			expected: nil,
			wantErr:  false,
		},
		{
			name:     "nested array",
			input:    "*2\r\n*2\r\n+hello\r\n+world\r\n$5\r\nredis\r\n",
			expected: []interface{}{[]interface{}{"hello", "world"}, "redis"},
			wantErr:  false,
		},
		{
			name:     "mixed types array",
			input:    "*5\r\n:1\r\n:2\r\n:3\r\n:4\r\n$5\r\nhello\r\n",
			expected: []interface{}{1, 2, 3, 4, "hello"},
			wantErr:  false,
		},
		{
			name:    "invalid array length",
			input:   "*abc\r\n",
			wantErr: true,
		},
		{
			name:    "invalid array element",
			input:   "*1\r\n&invalid\r\n",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := createParserWithInput(tt.input)
			result, err := parser.ReadResponse()

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// For nil expected value
			if tt.expected == nil {
				if result != nil {
					t.Errorf("expected nil, got %v", result)
				}
				return
			}

			// Compare arrays
			expectedArr := tt.expected.([]interface{})
			resultArr, ok := result.([]interface{})
			if !ok {
				t.Errorf("expected array type, got %T", result)
				return
			}

			if len(expectedArr) != len(resultArr) {
				t.Errorf("expected array length %d, got %d", len(expectedArr), len(resultArr))
				return
			}

			for i := range expectedArr {
				if !deepEqual(expectedArr[i], resultArr[i]) {
					t.Errorf("at index %d: expected %v, got %v", i, expectedArr[i], resultArr[i])
				}
			}
		})
	}
}

// Helper function to compare nested arrays and other types
func deepEqual(expected, actual interface{}) bool {
	switch exp := expected.(type) {
	case []interface{}:
		act, ok := actual.([]interface{})
		if !ok || len(exp) != len(act) {
			return false
		}
		for i := range exp {
			if !deepEqual(exp[i], act[i]) {
				return false
			}
		}
		return true
	default:
		return expected == actual
	}
}

func TestUnknownType(t *testing.T) {
	parser := createParserWithInput("?invalid\r\n")
	_, err := parser.ReadResponse()
	if err == nil {
		t.Error("expected error for unknown type but got none")
	}
}

func TestReadLineErrors(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "EOF before CRLF",
			input:   "+OK",
			wantErr: true,
		},
		{
			name:    "missing CR",
			input:   "+OK\n",
			wantErr: true,
		},
		{
			name:    "missing LF",
			input:   "+OK\r",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := createParserWithInput(tt.input)
			_, err := parser.ReadResponse()
			if tt.wantErr && err == nil {
				t.Error("expected error but got none")
			}
		})
	}
}
