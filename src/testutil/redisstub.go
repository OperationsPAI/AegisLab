package testutil

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"testing"
)

func StartRedisStub(tb testing.TB) (string, func()) {
	tb.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		tb.Fatalf("listen redis stub: %v", err)
	}

	done := make(chan struct{})
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				select {
				case <-done:
					return
				default:
					return
				}
			}

			go handleRedisStubConn(conn)
		}
	}()

	cleanup := func() {
		close(done)
		_ = ln.Close()
	}

	return ln.Addr().String(), cleanup
}

func handleRedisStubConn(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	for {
		cmd, err := readRESPArray(reader)
		if err != nil {
			if err == io.EOF {
				return
			}
			_, _ = writer.WriteString("-ERR invalid request\r\n")
			_ = writer.Flush()
			return
		}
		if len(cmd) == 0 {
			continue
		}

		switch strings.ToUpper(cmd[0]) {
		case "PING":
			_, _ = writer.WriteString("+PONG\r\n")
		case "HELLO":
			_, _ = writer.WriteString("%7\r\n+server\r\n+redis\r\n+version\r\n+7.0.0\r\n+proto\r\n:3\r\n+id\r\n:1\r\n+mode\r\n+standalone\r\n+role\r\n+master\r\n+modules\r\n*0\r\n")
		case "CLIENT", "AUTH", "SELECT", "QUIT":
			_, _ = writer.WriteString("+OK\r\n")
		case "COMMAND":
			_, _ = writer.WriteString("*0\r\n")
		case "LPUSH", "HSET", "ZADD":
			_, _ = writer.WriteString(":1\r\n")
		default:
			_, _ = writer.WriteString("+OK\r\n")
		}

		if err := writer.Flush(); err != nil {
			return
		}
	}
}

func readRESPArray(reader *bufio.Reader) ([]string, error) {
	prefix, err := reader.ReadByte()
	if err != nil {
		return nil, err
	}
	if prefix != '*' {
		return nil, fmt.Errorf("unexpected prefix %q", prefix)
	}

	countLine, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	count, err := strconv.Atoi(strings.TrimSpace(countLine))
	if err != nil {
		return nil, err
	}

	items := make([]string, 0, count)
	for i := 0; i < count; i++ {
		bulkPrefix, err := reader.ReadByte()
		if err != nil {
			return nil, err
		}
		if bulkPrefix != '$' {
			return nil, fmt.Errorf("unexpected bulk prefix %q", bulkPrefix)
		}

		sizeLine, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		size, err := strconv.Atoi(strings.TrimSpace(sizeLine))
		if err != nil {
			return nil, err
		}

		buf := make([]byte, size+2)
		if _, err := io.ReadFull(reader, buf); err != nil {
			return nil, err
		}
		items = append(items, string(buf[:size]))
	}

	return items, nil
}
