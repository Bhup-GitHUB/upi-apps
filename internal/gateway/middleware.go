package gateway

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type contextKey string

const traceIDKey contextKey = "trace_id"

func TraceMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := r.Header.Get("X-Trace-Id")
		if traceID == "" {
			traceID = fmt.Sprintf("trc_%d", time.Now().UnixNano())
		}

		fmt.Println("gateway:", r.Method, r.URL.Path, "trace", traceID)

		ctx := context.WithValue(r.Context(), traceIDKey, traceID)
		w.Header().Set("X-Trace-Id", traceID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)
		fmt.Println("gateway:", r.Method, r.URL.Path, rw.status, time.Since(start))
	})
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func TraceIDFromContext(ctx context.Context) string {
	val, _ := ctx.Value(traceIDKey).(string)
	return val
}
