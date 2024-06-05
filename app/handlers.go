package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func EchoHandler(request *Request) *Response {
	match := request.Params[1]
	response := NewResponse()
	response.StatusCode = http.StatusOK
	response.Headers["Content-Type"] = "text/plain"
	response.Headers["Content-Length"] = fmt.Sprintf("%d", len(match))
	response.Body = strings.NewReader(match)
	return response
}

func UserAgentHandler(request *Request) *Response {
	userAgent := request.Headers["User-Agent"]
	response := NewResponse()
	response.StatusCode = http.StatusOK
	response.Headers["Content-Type"] = "text/plain"
	response.Headers["Content-Length"] = fmt.Sprintf("%d", len(userAgent))
	response.Body = strings.NewReader(userAgent)
	return response
}

func FilesHandler(request *Request) *Response {
	filename := request.Params[1]
	filepath := filepath.Join(rootDirectory, filename)
	response := NewResponse()
	if _, err := os.Stat(filepath); err != nil {
		response.StatusCode = http.StatusNotFound
		return response
	}
	file, err := os.Open(filepath)
	if err != nil {
		response.StatusCode = http.StatusInternalServerError
		return response
	}
	if stat, _ := file.Stat(); !stat.IsDir() {
		response.Headers["Content-Length"] = fmt.Sprintf("%d", stat.Size())
	}
	response.StatusCode = http.StatusOK
	response.Headers["Content-Type"] = "application/octet-stream"

	return response
}
