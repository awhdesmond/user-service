package users

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/awhdesmond/user-service/pkg/common"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
)

const (
	URLParamUsername = "username"
)

// errToHttpCode maps a specific error to a HTTP Status Code
func errToHttpCode(err error) int {
	if common.ErrorContains([]error{ErrUserNotFound}, err) {
		return http.StatusNotFound
	}
	if common.ErrorContains(
		[]error{
			ErrDoBFutureUsed,
			ErrDoBInvalid,
			ErrDoBTooOld,
			ErrUsernameContainsNonLetters,
			ErrUsernameIsEmpty,
		},
		err,
	) {
		return http.StatusBadRequest
	}
	return http.StatusInternalServerError
}

func MakeHandler(svc Service) http.Handler {
	r := mux.NewRouter()

	opts := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(common.EncodeErrorFactory(errToHttpCode)),
	}

	readHandler := kithttp.NewServer(
		NewReadEndpoint(svc),
		decodeReadRequest,
		encodeReadResponse,
		opts...,
	)
	upsertHandler := kithttp.NewServer(
		NewUpsertEndpoint(svc),
		decodeUpsertRequest,
		encodeUpsertResponse,
		opts...,
	)

	r.Handle("/hello/{username}", readHandler).Methods(http.MethodGet)
	r.Handle("/hello/{username}", upsertHandler).Methods(http.MethodPut)

	return r
}

func decodeReadRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	req := ReadRequest{vars[URLParamUsername]}
	return req, nil
}

func encodeReadResponse(ctx context.Context, w http.ResponseWriter, resp interface{}) error {
	if e, ok := resp.(common.Errorer); ok && e.Error() != nil {
		common.EncodeErrorFactory(errToHttpCode)(ctx, e.Error(), w)
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(resp)
}

func decodeUpsertRequest(_ context.Context, r *http.Request) (interface{}, error) {
	req := UpsertRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, common.ErrInvalidJSONBody
	}

	vars := mux.Vars(r)
	req.Username = vars[URLParamUsername]
	return req, nil
}

func encodeUpsertResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(common.Errorer); ok && e.Error() != nil {
		common.EncodeErrorFactory(errToHttpCode)(ctx, e.Error(), w)
		return nil
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}
