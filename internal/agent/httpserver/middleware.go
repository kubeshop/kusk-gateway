package httpserver

import (
	"net/http"
	"runtime/debug"
	"time"

	"go.uber.org/zap"
)

func LoggerMiddleware(l *zap.Logger, next http.Handler) http.Handler {

	fn := func(w http.ResponseWriter, r *http.Request) {
		// Recovery in case of panic
		defer func() {
			if err := recover(); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				l.Error("Panic",
					zap.String("path", r.URL.EscapedPath()),
					zap.Any("error", err),
					zap.ByteString("trace", debug.Stack()),
				)
			}
		}()
		startTime := time.Now()
		wrapped := wrapResponseWriter(w)
		next.ServeHTTP(wrapped, r)
		// Read the response and log the results
		l.Info("Served",
			zap.String("path", r.URL.EscapedPath()),
			zap.Duration("duration", time.Since(startTime)),
			zap.Int("size", wrapped.size),
			zap.Int("status", wrapped.status),
		)

	}
	return http.HandlerFunc(fn)
}

// responseWriter is a minimal wrapper for http.ResponseWriter that allows to capture response stats
type responseWriter struct {
	http.ResponseWriter
	status      int
	size        int
	wroteHeader bool
}

func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w}
}
func (rw *responseWriter) Status() int {
	return rw.status
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}

	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
	rw.wroteHeader = true

	return
}

func (rw *responseWriter) Write(body []byte) (int, error) {
	rw.size = len(body)
	return rw.ResponseWriter.Write(body)
}
