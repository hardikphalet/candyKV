# Go Redis Server Implementation

[![Go Tests](https://github.com/hardikphalet/candyKV/actions/workflows/go.yml/badge.svg)](https://github.com/hardikphalet/candyKV/actions/workflows/go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/hardikphalet/go-redis)](https://goreportcard.com/report/github.com/hardikphalet/go-redis)

A lightweight Redis-compatible server implementation in Go that supports core Redis functionality. This project implements the Redis RESP (REdis Serialization Protocol) protocol and provides a subset of Redis commands.

## Features

### Core Features
- TCP server implementation with concurrent client handling
- Full RESP (Redis Serialization Protocol) protocol support
- In-memory key-value store
- Command parser and executor
- Graceful shutdown support
- Client connection management

### Supported Commands

#### Basic Operations
- `PING` - Test server connectivity
- `ECHO <message>` - Echo back the given message
- `SET <key> <value> [options]` - Set key to hold string value with optional parameters
  - Options: `NX` (only set if key doesn't exist)
  - Options: `XX` (only set if key exists)
  - Options: `GET` (return old value)
  - Options: `EX <seconds>` (set expiry in seconds)
  - Options: `PX <milliseconds>` (set expiry in milliseconds)
  - Options: `EXAT <timestamp>` (set expiry at Unix timestamp in seconds)
  - Options: `PXAT <timestamp>` (set expiry at Unix timestamp in milliseconds)
  - Options: `KEEPTTL` (retain the TTL associated with the key)
- `GET <key>` - Get the value of a key
- `DEL <key> [key ...]` - Delete one or more keys
- `EXPIRE <key> <seconds> [options]` - Set a key's time to live in seconds
- `TTL <key>` - Get the time to live for a key in seconds
- `KEYS <pattern>` - Find all keys matching the given pattern

#### Sorted Sets
- `ZADD <key> [options] <score> <member> [<score> <member> ...]` - Add members to a sorted set
  - Options: `NX` (only add new elements)
  - Options: `XX` (only update existing elements)
  - Options: `GT` (only update existing elements if new score is greater than current)
  - Options: `LT` (only update existing elements if new score is less than current)
  - Options: `CH` (modify the return value to be the numbers of changed elements)
  - Options: `INCR` (when specified, ZADD acts like ZINCRBY)
- `ZRANGE <key> <start> <stop>` - Return a range of members in a sorted set

### Client Implementation
- Redis-compatible client implementation in Go
- Support for all implemented commands
- Automatic connection handling and reconnection
- Command-line interface (CLI) with interactive mode
- Error handling and response parsing

## Getting Started

### Prerequisites
- Go 1.x or higher

### Installation
```bash
git clone https://github.com/yourusername/go-redis.git
cd go-redis
go mod download
```

### Running the Server
```bash
go run cmd/server/main.go
```
The server will start listening on the default Redis port (6379).

### Using the CLI Client
```bash
go run cmd/client/main.go
```

## Project Structure
```
.
├── cmd/
│   ├── client/         # Client CLI implementation
│   └── server/         # Server implementation
├── internal/
│   ├── commands/       # Command implementations
│   ├── resp/          # RESP protocol implementation
│   ├── server/        # Server core functionality
│   ├── store/         # In-memory store implementation
│   └── types/         # Common types and interfaces
├── pkg/
│   └── client/        # Client library implementation
└── docs/              # Documentation
```

## Implementation Details

### RESP Protocol
The server implements the RESP (REdis Serialization Protocol) with support for all data types:
- Simple Strings ("+")
- Errors ("-")
- Integers (":")
- Bulk Strings ("$")
- Arrays ("*")

### Server Architecture
- Non-blocking I/O with goroutines for handling multiple clients
- Thread-safe in-memory store implementation
- Command pattern for easy addition of new commands
- Graceful shutdown with connection draining
- Comprehensive error handling

### Testing
The project includes extensive unit tests covering:
- Server functionality
- Command parsing and execution
- RESP protocol implementation
- Client library functionality
- Connection handling
- Concurrent operations

## Contributing
Contributions are welcome! Please feel free to submit a Pull Request.

## License
This project is licensed under the MIT License - see the LICENSE file for details.
