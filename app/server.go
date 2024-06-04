package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net"
	"strings"
)

const (
	PORT = 4221
)

type Request struct {
	Method  string
	Version string
	Path    string
	Headers map[string]string
	Body    string
}

// splitCRLF is a split function for a Scanner that splits on '\r\n'.
func splitCRLF(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	// Find the index of '\r\n'.
	if i := bytes.Index(data, []byte("\r\n")); i >= 0 {
		// We have a full '\r\n'-terminated line.
		return i + 2, data[0:i], nil
	}

	// If at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}

	// Request more data.
	return 0, nil, nil
}

func parseRequest(conn net.Conn) *Request {
	var method, path, version string

	scanner := bufio.NewScanner(conn)
	scanner.Split(splitCRLF)

	// Read request line
	if scanner.Scan() {
		requestLine := scanner.Text()
		parts := strings.Split(requestLine, " ")
		method, path, version = parts[0], parts[1], parts[2]
	}

	// Read headers
	headers := make(map[string]string)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			break
		}
		parts := strings.Split(line, ":")
		key, value := parts[0], strings.Join(parts[1:], "")
		headers[key] = strings.Trim(value, " ")
	}

	return &Request{
		Method:  method,
		Version: version,
		Path:    path,
		Headers: headers,
	}
}

func main() {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", PORT))
	if err != nil {
		log.Fatalf("Failed to bind to port %d", PORT)
	}

	fmt.Println("Listening on port:", PORT)
	conn, err := l.Accept()
	if err != nil {
		log.Fatalf("Error accepting connection: %s", err)
	}
	defer conn.Close()

	request := parseRequest(conn)
	fmt.Printf("Request: %+v\n", request)

	if request.Path == "/" {
		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	} else {
		conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	}
}
