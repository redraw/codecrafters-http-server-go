package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
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
	file, err := os.Open(filepath)
	if err != nil {
		return handleNotFound(request)
	}
	if stat, err := file.Stat(); err == nil {
		response.Headers["Content-Length"] = fmt.Sprintf("%d", stat.Size())
	}
	response.StatusCode = 200
	response.Headers["Content-Type"] = "application/octet-stream"
	response.Body = file
	return response
}

func handlePostFile(request *Request) *Response {
	filename := request.Params[1]
	response := NewResponse()

	filepath := filepath.Join(rootDirectory, filename)
	file, err := os.Create(filepath)
	if err != nil {
		response.StatusCode = 500
		return response
	}
	defer file.Close()

	contentLength, _ := strconv.ParseInt(request.Headers["Content-Length"], 10, 64)
	_, err = io.CopyN(file, request.Body, contentLength)
	if err != nil {
		response.StatusCode = 500
		return response
	}
	response.StatusCode = 201
	return response
}

func handleFiles(request *Request) *Response {
	switch request.Method {
	case "GET":
		return handleGetFile(request)
	case "POST":
		return handlePostFile(request)
	}
	return handleNotFound(request)
}

func handleNotFound(_ *Request) *Response {
	response := NewResponse()
	response.StatusCode = 404
	return response
}

func handleFound(_ *Request) *Response {
	response := NewResponse()
	response.StatusCode = 200
	return response
}
