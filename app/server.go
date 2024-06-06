package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"path/filepath"
	"regexp"
)

const (
	PORT = 4221
)

type Handler func(ResponseWriter, *Request)

type Server struct {
	routes map[string]Handler
}

var rootDirectory string

func main() {
	flag.StringVar(&rootDirectory, "directory", "", "Directory to serve")
	flag.Parse()

	if rootDirectory == "" {
		rootDirectory, _ = filepath.Abs(".")
	}

	server := NewServer()
	server.Route(`^/echo/(.*)$`, handleEcho)
	server.Route(`^/user-agent$`, handleUserAgent)
	server.Route(`^/files/(.*)$`, handleFiles)
	server.Route(`^/$`, handleFound)

	if err := server.Listen(); err != nil {
		log.Fatalf("Error starting server: %s", err)
	}
}

func NewServer() *Server {
	return &Server{
		routes: make(map[string]Handler),
	}
}

func (s *Server) Listen() error {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", PORT))
	if err != nil {
		return err
	}
	fmt.Println("Listening on port:", PORT)

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatalf("Error accepting connection: %s", err)
		}

		request, err := NewRequest(conn).Parse()
		if err != nil {
			log.Fatalf("Error parsing request: %s", err)
		}

		fmt.Printf("Request: %+v\n", request)
		go s.Handle(request)
	}
}

func (s *Server) Route(pattern string, handler Handler) {
	s.routes[pattern] = handler
}

func (s *Server) Handle(request *Request) {
	writer := NewResponse(request.conn)
	defer writer.Close()

	for path, handler := range s.routes {
		pattern := regexp.MustCompile(path)
		if pattern.MatchString(request.Path) {
			request.Params = pattern.FindStringSubmatch(request.Path)
			handler(writer, request)
			return
		}
	}

	handleNotFound(writer, request)
}
