package api

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
)

type MetricsMiddleware struct {
	Histogram *prometheus.HistogramVec
}

func NewMetricsMiddleware() *MetricsMiddleware {
	histogram := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Subsystem: "http",
		Name:      "request_duration_seconds",
		Help:      "seconds spent serving HTTP requests",
		Buckets:   prometheus.DefBuckets,
	}, []string{"method", "path", "status"})

	prometheus.MustRegister(histogram)
	return &MetricsMiddleware{Histogram: histogram}
}

func (mw *MetricsMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		begin := time.Now()

		path := mw.getRouteName(r)
		next.ServeHTTP(w, r)
		took := time.Since(begin)

		wrw, ok := w.(*wrappedResponseWriter)
		if ok {
			statusCode := fmt.Sprintf("%d", wrw.statusCode)
			mw.Histogram.WithLabelValues(r.Method, path, statusCode).Observe(took.Seconds())
		}
	})
}

func (mw *MetricsMiddleware) getRouteName(r *http.Request) string {
	currentRoute := mux.CurrentRoute(r)
	if currentRoute != nil {
		if path, err := currentRoute.GetPathTemplate(); err == nil {
			if len(path) > 0 {
				return mw.urlToLabel(path)
			}
		}
	}
	return mw.urlToLabel(r.RequestURI)
}

var invalidChars = regexp.MustCompile(`[^a-zA-Z0-9]+`)

func (mw *MetricsMiddleware) urlToLabel(path string) string {
	result := invalidChars.ReplaceAllString(path, "_")
	result = strings.ToLower(strings.Trim(result, "_"))
	if result == "" {
		return "root"
	}
	return result
}
