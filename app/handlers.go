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
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(match)))
	w.WriteStatus(200)
	fmt.Fprint(w, match)
}

func handleUserAgent(w ResponseWriter, request *Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(request.Headers["User-Agent"])))
	w.WriteStatus(200)
	fmt.Fprint(w, request.Headers["User-Agent"])
}

func handleGetFile(w ResponseWriter, request *Request) {
	filename := request.Params[1]
	filepath := filepath.Join(rootDirectory, filename)
	file, err := os.Open(filepath)
	if err != nil {
		handleNotFound(w, request)
		return
	}
	defer file.Close()

	stat, _ := file.Stat()
	w.Header().Set("Content-Length", fmt.Sprintf("%d", stat.Size()))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteStatus(200)
	io.Copy(w, file)
}

func handlePostFile(w ResponseWriter, request *Request) {
	filename := request.Params[1]
	filepath := filepath.Join(rootDirectory, filename)
	file, err := os.Create(filepath)
	if err != nil {
		w.WriteStatus(500)
		return
	}
	defer file.Close()

	contentLength, _ := strconv.ParseInt(request.Headers["Content-Length"], 10, 64)
	_, err = io.CopyN(file, request.Body, contentLength)
	if err != nil {
		w.WriteStatus(500)
		return
	}
	w.WriteStatus(201)
	w.Header().Set("Location", "/files/"+filename)
	w.WriteHeader()
}

func handleFiles(w ResponseWriter, r *Request) {
	switch r.Method {
	case "GET":
		handleGetFile(w, r)
	case "POST":
		handlePostFile(w, r)
	default:
		handleNotFound(w, r)
	}
}

func handleNotFound(w ResponseWriter, _ *Request) {
	w.WriteStatus(404)
	fmt.Fprint(w, "Not Found")
}

func handleFound(w ResponseWriter, _ *Request) {
	w.WriteStatus(200)
	fmt.Fprint(w, "OK")
}
