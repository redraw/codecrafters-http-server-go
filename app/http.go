package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
)

type Header map[string]string

type Request struct {
	Method  string
	Version string
	Path    string
	Header  Header
	Body    io.Reader
	Params  []string
	conn    net.Conn
}

type ResponseWriter interface {
	Header() Header
	WriteHeader(int) error
	Write([]byte) (int, error)
}

type Response struct {
	statusCode int
	header     Header
	headerSent bool
	conn       net.Conn
}

func NewRequest(conn net.Conn) *Request {
	return &Request{
		Header: make(Header),
		conn:   conn,
	}
}

func NewResponse(conn net.Conn) *Response {
	return &Response{
		header: make(Header),
		conn:   conn,
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
		r.Header[key] = value
	}

	// Parse body
	r.Body = reader

	return r, nil
}

func (r *Request) AcceptsEncoding(encoding string) bool {
	for _, item := range strings.Split(r.Header["Accept-Encoding"], ",") {
		if value := strings.TrimSpace(item); encoding == value {
			return true
		}
	}
	return false
}

func (r *Response) WriteHeader(statusCode int) error {
	if r.headerSent {
		return nil
	}
	r.statusCode = statusCode

	// Write response line
	_, err := fmt.Fprintf(r.conn, "HTTP/1.1 %d %s\r\n", r.statusCode, http.StatusText(r.statusCode))
	if err != nil {
		return err
	}

	// Write headers
	for key, value := range r.header {
		_, err = fmt.Fprintf(r.conn, "%s: %s\r\n", key, value)
		if err != nil {
			return err
		}
	}

	fmt.Fprintf(r.conn, "\r\n")

	return nil
}

func (r *Response) Write(data []byte) (int, error) {
	if !r.headerSent {
		if err := r.WriteHeader(200); err != nil {
			return 0, err
		}
		r.headerSent = true
	}

	return r.conn.Write(data)
}

func (h Header) Set(key string, value string) {
	h[key] = value
}

func (h Header) Get(key string) (string, bool) {
	value, ok := h[key]
	return value, ok
}

func (r *Response) Header() Header {
	return r.header
}

func (r *Response) Close() {
	r.conn.Close()
}
