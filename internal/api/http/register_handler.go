package http

import (
	"context"

	"github.com/SergeiGD/testify-profile/internal/dto"
	"github.com/SergeiGD/testify-profile/internal/usecases/register"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

type registerHandler struct {
	useCase register.RegisterUseCase
}

func NewRegisterHandler(useCase register.RegisterUseCase) *registerHandler {
	return &registerHandler{useCase: useCase}
}

func (h *registerHandler) Register(ctx context.Context, request RegisterRequestObject) (RegisterResponseObject, error) {
	if err := validate.Struct(request.Body); err != nil {
		return Register400JSONResponse{Detail: mapValidationError(err)}, nil
	}

	input := dto.RegisterInput{
		Email:     string(request.Body.Email),
		Password:  request.Body.Password,
		Username:  request.Body.Username,
		BirthDate: request.Body.BirthDate.Time,
	}

	if err := h.useCase.Register(ctx, input); err != nil {
		return Register400JSONResponse{Detail: mapRegisterError(err)}, nil
	}

	return Register201JSONResponse{IsOk: true}, nil
}

func (h *registerHandler) RegisterConfirm(ctx context.Context, request RegisterConfirmRequestObject) (RegisterConfirmResponseObject, error) {
	if err := h.useCase.Confirm(ctx, request.Token); err != nil {
		return RegisterConfirm400JSONResponse{Detail: mapConfirmError(err)}, nil
	}

	return RegisterConfirm200JSONResponse{IsOk: true}, nil
}

