package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func handleEcho(request *Request) *Response {
	match := request.Params[1]
	response := NewResponse()
	response.StatusCode = http.StatusOK
	response.Headers["Content-Type"] = "text/plain"
	response.Headers["Content-Length"] = fmt.Sprintf("%d", len(match))
	response.Body = strings.NewReader(match)
	return response
}

func handleUserAgent(request *Request) *Response {
	userAgent := request.Headers["User-Agent"]
	response := NewResponse()
	response.StatusCode = http.StatusOK
	response.Headers["Content-Type"] = "text/plain"
	response.Headers["Content-Length"] = fmt.Sprintf("%d", len(userAgent))
	response.Body = strings.NewReader(userAgent)
	return response
}

func handleGetFile(request *Request) *Response {
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
	response.Body = file
	return response
}

func handlePostFile(request *Request) *Response {
	filename := request.Params[1]
	filepath := filepath.Join(rootDirectory, filename)
	file, err := os.Create(filepath)
	if err != nil {
		response := NewResponse()
		response.StatusCode = http.StatusInternalServerError
		return response
	}
	defer file.Close()
	buf := make([]byte, 1024)
	n, err := request.Body.Read(buf)
	if err != nil && err != io.EOF {
		fmt.Println("Error reading request body:", err)
	}
	fmt.Println("Read", n, "bytes from request body")
	_, err = io.Copy(file, request.Body)
	if err != nil {
		response := NewResponse()
		response.StatusCode = http.StatusInternalServerError
		return response
	}
	response := NewResponse()
	response.StatusCode = http.StatusCreated
	return response
}

func handleFiles(request *Request) *Response {
	switch request.Method {
	case "GET":
		return handleGetFile(request)

	case "POST":
		return handlePostFile(request)
	}
	return handleDefault(request)
}

func handleDefault(request *Request) *Response {
	response := NewResponse()
	response.StatusCode = http.StatusNotFound
	return response
}
