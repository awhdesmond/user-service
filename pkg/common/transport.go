package common

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
)

var (
	ErrInvalidJSONBody     = errors.New("common.ErrInvalidJSONBody")
	ErrEndpointReqMismatch = errors.New("common.ErrEndpointReqMismatch")
)

func EncodeErrorFactory(errToCode func(error) int) func(context.Context, error, http.ResponseWriter) {
	return func(ctx context.Context, err error, w http.ResponseWriter) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		code := errToCode(err)

		if code < 0 {
			switch err {
			case ErrInvalidJSONBody:
				w.WriteHeader(http.StatusBadRequest)
			default:
				w.WriteHeader(http.StatusInternalServerError)
			}
		} else {
			w.WriteHeader(code)
		}
		if err := json.NewEncoder(w).Encode(GenericJSON{"error": err.Error()}); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}
