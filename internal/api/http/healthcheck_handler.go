package http

import "context"

type healthcheckHandler struct {
}

func NewHealthcheckHandler() *healthcheckHandler {
	return &healthcheckHandler{}
}

func (h *healthcheckHandler) Healthcheck(ctx context.Context, request HealthcheckRequestObject) (HealthcheckResponseObject, error) {
	return Healthcheck200JSONResponse{
		IsOk: true,
	}, nil
}
