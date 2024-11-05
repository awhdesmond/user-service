package api

import (
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

type MuxLoggingMiddleware struct {
	logger *zap.Logger
}

func NewMuxLoggingMiddleware(logger *zap.Logger) *MuxLoggingMiddleware {
	return &MuxLoggingMiddleware{logger}
}

func (m *MuxLoggingMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
		statusCode := "0"

		wrw, ok := w.(*wrappedResponseWriter)
		if ok {
			statusCode = fmt.Sprintf("%d", wrw.statusCode)
		}

		m.logger.Debug(
			"request",
			zap.String("proto", r.Proto),
			zap.String("uri", r.RequestURI),
			zap.String("method", r.Method),
			zap.String("remote", r.RemoteAddr),
			zap.String("user-agent", r.UserAgent()),
			zap.String("statusCode", statusCode),
		)
	})
}
