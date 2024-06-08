package main

import (
	"fmt"
	"log"
	"net"
	"regexp"
)

type Handler func(ResponseWriter, *Request)

type HttpServer struct {
	routes     map[string]Handler
	middleware []Middleware
	rootDir    string
}

func NewServer(rootDir string) *HttpServer {
	return &HttpServer{
		routes:  make(map[string]Handler),
		rootDir: rootDir,
	}
}

func (s *HttpServer) Listen(addr string) error {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	fmt.Println("Listening on", addr)

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatalf("Error accepting connection: %s", err)
		}

		go s.Handle(conn)
	}
}

func (s *HttpServer) Route(pattern string, handler Handler) {
	s.routes[pattern] = handler
}

func (s *HttpServer) Use(m Middleware) {
	s.middleware = append(s.middleware, m)
}

func (s *HttpServer) Handle(conn net.Conn) {
	defer conn.Close()

	request := NewRequest(conn)
	if err := request.Parse(); err != nil {
		log.Printf("Error parsing request: %s", err)
		return
	}

	request.ServerRoot = s.rootDir
	response := NewResponse(conn)
	defer response.Send()

	// Loop routes
	for path, handler := range s.routes {
		pattern := regexp.MustCompile(path)
		if pattern.MatchString(request.Path) {
			request.Params = pattern.FindStringSubmatch(request.Path)
			// Add middlewares
			for _, m := range s.middleware {
				handler = m.Handle(handler)
			}
			// Handle!
			handler(response, request)
			return
		}
	}

	// Default 404
	response.Status(404)
}
