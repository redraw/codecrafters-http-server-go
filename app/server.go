package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net"
	"net/http"
	"regexp"
	"strings"
)

const (
	PORT = 4221
)

type Headers map[string]string

type Request struct {
	Method  string
	Version string
	Path    string
	Headers Headers
	Body    string
}

type Response struct {
	Version    string
	StatusCode int
	Headers    Headers
	Body       string
}

func NewRequest() *Request {
	return &Request{
		Headers: make(Headers),
	}
}

func (r *Request) Parse(conn net.Conn) (*Request, error) {
	scanner := bufio.NewScanner(conn)
	scanner.Split(splitCRLF)

	// Read request line
	if scanner.Scan() {
		requestLine := scanner.Text()
		parts := strings.Split(requestLine, " ")
		r.Method, r.Path, r.Version = parts[0], parts[1], parts[2]
	}

	// Read headers
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			break
		}
		parts := strings.Split(line, ":")
		key, value := parts[0], strings.Join(parts[1:], "")
		r.Headers[key] = strings.Trim(value, " ")
	}

	return r, nil
}

func NewResponse() *Response {
	return &Response{
		Headers: make(Headers),
	}
}

func (r *Response) Send(conn net.Conn) error {
	// Write response line
	_, err := fmt.Fprintf(conn, "HTTP/1.1 %d %s\r\n", r.StatusCode, http.StatusText(r.StatusCode))
	if err != nil {
		return err
	}

	// Write headers
	r.Headers["Content-Length"] = fmt.Sprintf("%d", len(r.Body))
	for key, value := range r.Headers {
		_, err = fmt.Fprintf(conn, "%s: %s\r\n", key, value)
		if err != nil {
			return err
		}
	}

	// Write body
	_, err = fmt.Fprintf(conn, "\r\n%s", r.Body)
	if err != nil {
		return err
	}

	return nil
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

	request, err := NewRequest().Parse(conn)
	if err != nil {
		log.Fatalf("Error parsing request: %s", err)
	}
	fmt.Printf("Request: %+v\n", request)

	echoPath := regexp.MustCompile(`^/echo/(.*)$`)
	userAgentPath := regexp.MustCompile(`^/user-agent$`)

	switch {
	case request.Path == "/":
		response := NewResponse()
		response.StatusCode = http.StatusOK
		response.Headers["Content-Type"] = "text/plain"
		response.Body = "Hello, World!"
		response.Send(conn)
	case echoPath.MatchString(request.Path):
		match := echoPath.FindStringSubmatch(request.Path)[1]
		response := NewResponse()
		response.StatusCode = http.StatusOK
		response.Headers["Content-Type"] = "text/plain"
		response.Body = match
		response.Send(conn)
	case userAgentPath.MatchString(request.Path):
		response := NewResponse()
		response.StatusCode = http.StatusOK
		response.Headers["Content-Type"] = "text/plain"
		response.Body = request.Headers["User-Agent"]
		response.Send(conn)
	default:
		response := NewResponse()
		response.StatusCode = http.StatusNotFound
		response.Send(conn)
	}
}
