package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	PORT = 4221
)

func main() {
	directory := flag.String("directory", ".", "Directory to serve")
	flag.Parse()

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

		go handle(request, *directory)
	}
}

func handle(request *Request, directory string) {
	defer request.Close()

	echoPath := regexp.MustCompile(`^/echo/(.*)$`)
	userAgentPath := regexp.MustCompile(`^/user-agent$`)
	filesPath := regexp.MustCompile(`^/files/(.*)$`)

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
		response.Headers["Content-Length"] = fmt.Sprintf("%d", len(match))
		response.Body = strings.NewReader(match)
		request.Send(response)
	case userAgentPath.MatchString(request.Path):
		userAgent := request.Headers["User-Agent"]
		response := NewResponse()
		response.StatusCode = http.StatusOK
		response.Headers["Content-Type"] = "text/plain"
		response.Headers["Content-Length"] = fmt.Sprintf("%d", len(userAgent))
		response.Body = strings.NewReader(userAgent)
		request.Send(response)
	case filesPath.MatchString(request.Path):
		match := filesPath.FindStringSubmatch(request.Path)[1]
		filepath := filepath.Join(directory, match)
		response := NewResponse()
		if _, err := os.Stat(filepath); err != nil {
			response.StatusCode = http.StatusNotFound
			request.Send(response)
			return
		}
		file, err := os.Open(filepath)
		if err != nil {
			response.StatusCode = http.StatusInternalServerError
			request.Send(response)
			return
		}
		if stat, _ := file.Stat(); !stat.IsDir() {
			response.Headers["Content-Length"] = fmt.Sprintf("%d", stat.Size())
		}
		response.StatusCode = http.StatusOK
		response.Headers["Content-Type"] = "application/octet-stream"
		response.Body = file
		request.Send(response)
	default:
		response := NewResponse()
		response.StatusCode = http.StatusNotFound
		request.Send(response)
	}
}
