package main

import (
	"compress/gzip"
	"log"
)

type Middleware interface {
	Handle(Handler) Handler
}

type LoggingMiddleware struct{}

type GzipMiddleware struct{}

type gzipResponseWriter struct {
	writer *gzip.Writer
	ResponseWriter
}

func (l *LoggingMiddleware) Handle(h Handler) Handler {
	return func(w ResponseWriter, r *Request) {
		log.Printf("%s - %s %s %s", r.conn.RemoteAddr(), r.Method, r.Version, r.Path)
		h(w, r)
	}
}

func (g *gzipResponseWriter) Write(b []byte) (int, error) {
	return g.writer.Write(b)
}

func (g *GzipMiddleware) Handle(h Handler) Handler {
	return func(w ResponseWriter, r *Request) {
		if !r.AcceptsEncoding("gzip") {
			h(w, r)
			return
		}
		gw := gzip.NewWriter(w)
		defer gw.Close()

		w.Header().Set("Content-Encoding", "gzip")
		gzipResponse := &gzipResponseWriter{gw, w}

		h(gzipResponse, r)
	}
}
