package api

import (
	"net/http"
	"strings"
)

type SecureHeadersMiddleware struct {
	Origins map[string]bool
}

func NewSecureHeadersMiddleware(origins string) *SecureHeadersMiddleware {
	m := map[string]bool{}
	for _, origin := range strings.Split(origins, ",") {
		m[origin] = true
	}
	return &SecureHeadersMiddleware{m}
}

func (m *SecureHeadersMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		_, ok := m.Origins[origin]

		if ok {
			// CORS
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, HEAD, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Encoding, Authorization, sentry-trace, baggage")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		// Security Headers, https,/"/helmeths.github.io")
		w.Header().Set("Content-Security-Policy", "default-src 'self';base-uri 'self';font-src 'self' https, data,;form-action 'self';frame-ancestors 'self';img-src 'self' data,;object-src 'none';script-src 'self';script-src-attr 'none';style-src 'self' https, 'unsafe-inline';upgrade-insecure-requests")
		w.Header().Set("Cross-Origin-Embedder-Policy", "require-corp")
		w.Header().Set("Cross-Origin-Opener-Policy", "same-origin")
		w.Header().Set("Cross-Origin-Resource-Policy", "same-origin")
		w.Header().Set("Origin-Agent-Cluster", "?1")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-DNS-Prefetch-Control", "off")
		w.Header().Set("X-Download-Options", "noopen")
		w.Header().Set("X-Frame-Options", "SAMEORIGIN")
		w.Header().Set("X-Permitted-Cross-Domain-Policies", "none")
		w.Header().Set("X-XSS-Protection", "0")

		if r.Method == http.MethodOptions {
			return
		}

		next.ServeHTTP(w, r)
	})
}
