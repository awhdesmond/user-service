package users

import (
	"context"

	"github.com/awhdesmond/user-service/pkg/common"
	"github.com/go-kit/kit/endpoint"
)

type BaseResponse struct {
	Err error `json:"error,omitempty"`
}

func (r BaseResponse) Error() error {
	return r.Err
}

type UpsertRequest struct {
	Username string `json:"username"`
	DoB      string `json:"dateOfBirth"`
}

type UpsertResponse struct {
	BaseResponse `json:",inline"`
}

func NewUpsertEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, epReq interface{}) (interface{}, error) {
		req, ok := epReq.(UpsertRequest)
		if !ok {
			return UpsertResponse{BaseResponse: BaseResponse{
				Err: common.ErrEndpointReqMismatch,
			}}, nil
		}
		err := svc.Upsert(ctx, req.Username, req.DoB)
		return UpsertResponse{BaseResponse: BaseResponse{Err: err}}, nil
	}
}

type ReadRequest struct {
	Username string `json:"username"`
}

type ReadResponse struct {
	BaseResponse `json:",inline"`
	Message      string `json:"message,omitempty"`
}

func NewReadEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, epReq interface{}) (interface{}, error) {
		req, ok := epReq.(ReadRequest)
		if !ok {
			return ReadResponse{BaseResponse: BaseResponse{
				Err: common.ErrEndpointReqMismatch,
			}}, nil
		}
		msg, err := svc.Read(ctx, req.Username)
		return ReadResponse{
			BaseResponse: BaseResponse{Err: err},
			Message:      msg,
		}, nil
	}
}
