package main

import (
	"bufio"
	"fmt"
	"io"
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
	Body    io.Reader
	Params  []string
	conn    net.Conn
}

type Response struct {
	Version    string
	StatusCode int
	Headers    Headers
	Body       io.Reader
}

func NewRequest(conn net.Conn) *Request {
	return &Request{
		Headers: make(Headers),
		conn:    conn,
	}
}

func (r *Request) Close() {
	r.conn.Close()
}

func (r *Request) Parse() (*Request, error) {
	reader := bufio.NewReader(r.conn)

	// Parse request line
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	parts := strings.Fields(line)
	if len(parts) != 3 {
		return nil, fmt.Errorf("malformed request line: %s", line)
	}

	r.Method = parts[0]
	r.Path = parts[1]
	r.Version = parts[2]

	// Parse headers
	for {
		line, err = reader.ReadString('\n')
		if err != nil {
			return nil, err
		}

		line = strings.TrimSpace(line)
		if line == "" {
			break
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("malformed header: %s", line)
		}

		key := parts[0]
		value := strings.TrimSpace(parts[1])
		r.Headers[key] = value
	}

	// Parse body
	r.Body = reader

	return r, nil
}

func (r *Request) Send(response *Response) error {
	// Write response line
	_, err := fmt.Fprintf(r.conn, "HTTP/1.1 %d %s\r\n", response.StatusCode, http.StatusText(response.StatusCode))
	if err != nil {
		return err
	}

	// Write headers
	for key, value := range response.Headers {
		_, err = fmt.Fprintf(r.conn, "%s: %s\r\n", key, value)
		if err != nil {
			return err
		}
	}

	fmt.Fprintf(r.conn, "\r\n")

	// Write body
	_, err = io.Copy(r.conn, response.Body)
	if err != nil {
		return err
	}

	return nil
}

func NewResponse() *Response {
	return &Response{
		Headers: make(Headers),
		Body:    strings.NewReader(""),
	}
}
