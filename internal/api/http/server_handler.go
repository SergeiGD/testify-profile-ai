package http

import (
	"context"

	"github.com/SergeiGD/testify-profile/internal/usecases/register"
)

// serverHandler combines all handler groups into a single StrictServerInterface.
type serverHandler struct {
	healthcheck *healthcheckHandler
	register    *registerHandler
}

func NewServerHandler(registerUseCase register.RegisterUseCase) StrictServerInterface {
	return &serverHandler{
		healthcheck: &healthcheckHandler{},
		register:    NewRegisterHandler(registerUseCase),
	}
}

func (h *serverHandler) Healthcheck(ctx context.Context, request HealthcheckRequestObject) (HealthcheckResponseObject, error) {
	return h.healthcheck.Healthcheck(ctx, request)
}

func (h *serverHandler) Register(ctx context.Context, request RegisterRequestObject) (RegisterResponseObject, error) {
	return h.register.Register(ctx, request)
}

func (h *serverHandler) RegisterConfirm(ctx context.Context, request RegisterConfirmRequestObject) (RegisterConfirmResponseObject, error) {
	return h.register.RegisterConfirm(ctx, request)
}
