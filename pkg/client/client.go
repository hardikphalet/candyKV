package client

import (
	"bufio"
	"fmt"
	"net"

	"github.com/hardikphalet/go-redis/internal/resp"
)

type Client struct {
	conn           net.Conn
	respWriter     *resp.Writer
	responseParser *ResponseParser
}

func NewClient(address string) (*Client, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis server: %v", err)
	}

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	return &Client{
		conn:           conn,
		respWriter:     resp.NewWriter(writer),
		responseParser: NewResponseParser(reader),
	}, nil
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Client) Send(command string, args ...string) error {
	cmdArray := make([]string, 0, len(args)+1)
	cmdArray = append(cmdArray, command)
	cmdArray = append(cmdArray, args...)

	return c.respWriter.WriteArray(cmdArray)
}

func (c *Client) Receive() (interface{}, error) {
	return c.responseParser.ReadResponse()
}
