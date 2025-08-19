package main

import (
	"compress/gzip"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type gzipResponseWriter struct {
	*gzip.Writer
	http.ResponseWriter
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func (w *gzipResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

func (w *gzipResponseWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
}

func gzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Content-Encoding", "gzip")
		gzw := gzip.NewWriter(w)
		defer func() { _ = gzw.Close() }()

		next.ServeHTTP(&gzipResponseWriter{gzw, w}, r)
	})
}

func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		slog.Info("request",
			"remote_addr", r.RemoteAddr,
			"method", r.Method,
			"url", r.URL.String(),
			"duration", time.Since(start).String(),
		)
	})
}
