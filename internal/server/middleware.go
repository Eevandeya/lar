package server

import (
	"log/slog"
	"net/http"
	"time"
)

type ResponseWriter struct {
	writer http.ResponseWriter
	status int
}

func (w *ResponseWriter) Header() http.Header {
	return w.writer.Header()
}

func (w *ResponseWriter) Write(b []byte) (int, error) {
	return w.writer.Write(b)
}

func (w *ResponseWriter) WriteHeader(statusCode int) {
	w.status = statusCode
	w.writer.WriteHeader(statusCode)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := ResponseWriter{
			writer: w,
			status: http.StatusOK,
		}

		start := time.Now()

		next.ServeHTTP(&rw, r)

		slog.Info("request completed",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rw.status,
			"duration", time.Since(start))
	})
}
