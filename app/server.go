package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"regexp"
)

const (
	PORT = 4221
)

type Handler func(*Request) *Response

type Server struct {
	routes map[string]Handler
}

var rootDirectory = "."

func main() {
	flag.StringVar(&rootDirectory, "directory", ".", "Directory to serve")
	flag.Parse()

	server := NewServer()
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

	// echoPath := regexp.MustCompile(`^/echo/(.*)$`)
	// userAgentPath := regexp.MustCompile(`^/user-agent$`)
	// filesPath := regexp.MustCompile(`^/files/(.*)$`)

	s.AddRoute(`^/echo/(.*)$`, EchoHandler)
	s.AddRoute(`^/user-agent$`, UserAgentHandler)

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatalf("Error accepting connection: %s", err)
		}

		request, err := NewRequest(conn).Parse()
		if err != nil {
			log.Fatalf("Error parsing request: %s", err)
		}

		go s.Handle(request)
	}
}

func (s *Server) AddRoute(pattern string, handler Handler) {
	s.routes[pattern] = handler
}

func (s *Server) Handle(request *Request) {
	defer request.Close()

	for path, handler := range s.routes {
		pattern := regexp.MustCompile(path)
		if pattern.MatchString(request.Path) {
			request.Params = pattern.FindStringSubmatch(request.Path)
			fmt.Printf("Request: %+v\n", request)
			response := handler(request)
			request.Send(response)
			return
		}
	}

	// Fallback to 404
	response := NewResponse()
	response.StatusCode = http.StatusNotFound
	request.Send(response)
}
