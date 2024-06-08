package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
)

func handleEcho(w ResponseWriter, request *Request) {
	match := request.Params[1]
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(w, match)
}

func handleUserAgent(w ResponseWriter, request *Request) {
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(w, request.Header["User-Agent"])
}

func handleGetFile(w ResponseWriter, request *Request) {
	filename := request.Params[1]
	filepath := filepath.Join(request.ServerRoot, filename)
	file, err := os.Open(filepath)
	if err != nil {
		w.Status(404)
		return
	}
	defer file.Close()

	w.Header().Set("Content-Type", "application/octet-stream")
	io.Copy(w, file)
}

func handlePostFile(w ResponseWriter, request *Request) {
	filename := request.Params[1]
	filepath := filepath.Join(request.ServerRoot, filename)
	file, err := os.Create(filepath)
	if err != nil {
		w.Status(500)
		return
	}
	defer file.Close()

	n, err := strconv.ParseInt(request.Header["Content-Length"], 10, 64)
	if err != nil {
		w.Status(400)
		fmt.Fprint(w, "Invalid Content-Length")
		return
	}

	_, err = io.CopyN(file, request.Body, n)
	if err != nil {
		w.Status(500)
		return
	}

	w.Status(201)
}

func handleFiles(w ResponseWriter, r *Request) {
	switch r.Method {
	case "GET":
		handleGetFile(w, r)
	case "POST":
		handlePostFile(w, r)
	default:
		w.Status(404)
	}
}

func handleFound(w ResponseWriter, _ *Request) {
	w.Status(200)
}
