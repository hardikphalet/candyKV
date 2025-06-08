package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/hardikphalet/go-redis/pkg/client"
)

func main() {

	redisClient, err := client.NewClient("localhost:6379")
	if err != nil {
		log.Fatalf("Failed to create Redis client: %v", err)
	}
	defer redisClient.Close()

	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("go-redis-cli")
	fmt.Println("Type 'help' for help, 'exit' to quit")

	for {
		fmt.Print("127.0.0.1:6379> ")
		if !scanner.Scan() {
			break
		}

		input := scanner.Text()
		if input == "" {
			continue
		}

		switch strings.ToLower(input) {
		case "exit", "quit":
			return
		case "help":
			printHelp()
			continue
		}

		args := parseCommand(input)
		if len(args) == 0 {
			continue
		}

		for retries := 0; retries < 3; retries++ {
			err := redisClient.Send(args[0], args[1:]...)
			if err != nil {
				if isConnectionError(err) && retries < 2 {
					redisClient.Close()
					redisClient, err = client.NewClient("localhost:6379")
					if err != nil {
						fmt.Printf("(error) Failed to reconnect: %v\n", err)
						continue
					}
					time.Sleep(100 * time.Millisecond)
					continue
				}
				fmt.Printf("(error) %v\n", err)
				break
			}

			response, err := redisClient.Receive()
			if err != nil {
				if isConnectionError(err) && retries < 2 {
					redisClient.Close()
					redisClient, err = client.NewClient("localhost:6379")
					if err != nil {
						fmt.Printf("(error) Failed to reconnect: %v\n", err)
						continue
					}
					time.Sleep(100 * time.Millisecond)
					continue
				}
				fmt.Printf("(error) %v\n", err)
				break
			}

			printResponse(response)
			break
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
	}
}

func isConnectionError(err error) bool {
	errStr := err.Error()
	return strings.Contains(errStr, "broken pipe") ||
		strings.Contains(errStr, "connection reset") ||
		strings.Contains(errStr, "EOF") ||
		strings.Contains(errStr, "connection refused")
}

func parseCommand(input string) []string {
	var args []string
	var currentArg strings.Builder
	inQuotes := false
	escaped := false

	for _, char := range input {
		if escaped {
			currentArg.WriteRune(char)
			escaped = false
			continue
		}

		switch char {
		case '\\':
			escaped = true
		case '"':
			inQuotes = !inQuotes
		case ' ', '\t':
			if inQuotes {
				currentArg.WriteRune(char)
			} else if currentArg.Len() > 0 {
				args = append(args, currentArg.String())
				currentArg.Reset()
			}
		default:
			currentArg.WriteRune(char)
		}
	}

	if currentArg.Len() > 0 {
		args = append(args, currentArg.String())
	}

	return args
}

func printResponse(response interface{}) {
	switch v := response.(type) {
	case nil:
		fmt.Println("(nil)")
	case string:
		fmt.Printf("\"%s\"\n", v)
	case []interface{}:
		for i, item := range v {
			fmt.Printf("%d) %v\n", i+1, item)
		}
	case int:
		fmt.Printf("(integer) %d\n", v)
	case error:
		fmt.Printf("(error) %s\n", v)
	default:
		fmt.Printf("%v\n", v)
	}
}

func printHelp() {
	fmt.Println(`Available commands:
  SET key value                    Set key to hold string value
  GET key                         Get the value of key
  DEL key [key ...]              Delete one or more keys
  EXPIRE key seconds              Set a key's time to live in seconds
  TTL key                         Get the time to live for a key in seconds
  KEYS pattern                    Find all keys matching the given pattern
  ZADD key score member          Add member with score to a sorted set
  ZRANGE key start stop          Return a range of members in a sorted set
  
Special commands:
  help                           Show this help
  exit                           Exit the CLI
  quit                           Exit the CLI`)
}
