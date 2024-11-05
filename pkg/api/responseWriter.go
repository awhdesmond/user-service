package api

import (
	"bufio"
	"errors"
	"net"
	"net/http"
)

type wrappedResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func NewWrappedResponseWriter(w http.ResponseWriter) *wrappedResponseWriter {
	return &wrappedResponseWriter{w, http.StatusOK}
}

func (wrw *wrappedResponseWriter) WriteHeader(code int) {
	wrw.statusCode = code
	wrw.ResponseWriter.WriteHeader(code)
}

func (wrw *wrappedResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := wrw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("hijack not supported")
	}
	return h.Hijack()
}

type WrappedReponseWriterMiddleware struct{}

func NewWrappedReponseWriterMiddleware() *WrappedReponseWriterMiddleware {
	return &WrappedReponseWriterMiddleware{}
}

func (mw *WrappedReponseWriterMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wrw := NewWrappedResponseWriter(w)
		next.ServeHTTP(wrw, r)
	})
}
