package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"regexp"
)

const (
	PORT = 4221
)

func main() {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", PORT))
	if err != nil {
		log.Fatalf("Failed to bind to port %d", PORT)
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

		go handle(request)
	}
}

func handle(request *Request) {
	defer request.conn.Close()

	echoPath := regexp.MustCompile(`^/echo/(.*)$`)
	userAgentPath := regexp.MustCompile(`^/user-agent$`)

	switch {
	case request.Path == "/":
		response := NewResponse()
		response.StatusCode = http.StatusOK
		request.Send(response)
	case echoPath.MatchString(request.Path):
		match := echoPath.FindStringSubmatch(request.Path)[1]
		response := NewResponse()
		response.StatusCode = http.StatusOK
		response.Headers["Content-Type"] = "text/plain"
		response.Body = match
		request.Send(response)
	case userAgentPath.MatchString(request.Path):
		response := NewResponse()
		response.StatusCode = http.StatusOK
		response.Headers["Content-Type"] = "text/plain"
		response.Body = request.Headers["User-Agent"]
		request.Send(response)
	default:
		response := NewResponse()
		response.StatusCode = http.StatusNotFound
		request.Send(response)
	}
}
