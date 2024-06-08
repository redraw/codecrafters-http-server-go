package main

import (
	"flag"
	"log"
	"path/filepath"
)

func main() {
	directory := flag.String("directory", ".", "Directory to serve")
	addr := flag.String("addr", ":4221", "Address to listen on")
	flag.Parse()

	rootDir, err := filepath.Abs(*directory)
	if err != nil {
		log.Fatalf("Error getting absolute path: %s", err)
	}

	app := NewApp(rootDir)

	if err := app.Listen(*addr); err != nil {
		log.Fatalf("Error starting server: %s", err)
	}
}

func NewApp(rootDir string) *HttpServer {
	s := NewServer(rootDir)

	s.Route(`^/echo/(.*)$`, handleEcho)
	s.Route(`^/user-agent$`, handleUserAgent)
	s.Route(`^/files/(.*)$`, handleFiles)
	s.Route(`^/$`, handleFound)

	s.Use(&LoggingMiddleware{})
	s.Use(&GzipMiddleware{})

	return s
}
