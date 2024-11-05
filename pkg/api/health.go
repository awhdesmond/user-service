package api

import (
	"net/http"
	"sync/atomic"
)

var (
	healthy int32
)

func GetHealthy() *int32 {
	return &healthy
}

func HealthzHandler(w http.ResponseWriter, r *http.Request) {
	atomic.StoreInt32(&healthy, 1)
	if atomic.LoadInt32(&healthy) == 1 {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("ok")); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusServiceUnavailable)
}
