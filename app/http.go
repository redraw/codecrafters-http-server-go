package main

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"strings"
)

type Headers map[string]string

type Request struct {
	Method  string
	Version string
	Path    string
	Headers Headers
	Body    string
	conn    net.Conn
}

type Response struct {
	Version    string
	StatusCode int
	Headers    Headers
	Body       string
}

func NewRequest(conn net.Conn) *Request {
	return &Request{
		Headers: make(Headers),
		conn:    conn,
	}
}

func (r *Request) Parse() (*Request, error) {
	scanner := bufio.NewScanner(r.conn)
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

func (r *Request) Send(response *Response) error {
	// Write response line
	_, err := fmt.Fprintf(r.conn, "HTTP/1.1 %d %s\r\n", response.StatusCode, http.StatusText(response.StatusCode))
	if err != nil {
		return err
	}

	// Write headers
	response.Headers["Content-Length"] = fmt.Sprintf("%d", len(response.Body))
	for key, value := range response.Headers {
		_, err = fmt.Fprintf(r.conn, "%s: %s\r\n", key, value)
		if err != nil {
			return err
		}
	}

	// Write body
	_, err = fmt.Fprintf(r.conn, "\r\n%s", response.Body)
	if err != nil {
		return err
	}

	return nil
}

func NewResponse() *Response {
	return &Response{
		Headers: make(Headers),
	}
}
