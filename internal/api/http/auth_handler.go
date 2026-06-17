package http

import (
	"context"

	"github.com/SergeiGD/testify-profile/internal/dto"
	"github.com/SergeiGD/testify-profile/internal/usecases/auth"
)

type authHandler struct {
	useCase auth.AuthUseCase
}

func NewAuthHandler(useCase auth.AuthUseCase) *authHandler {
	return &authHandler{useCase: useCase}
}

func (h *authHandler) Login(ctx context.Context, request LoginRequestObject) (LoginResponseObject, error) {
	if err := validate.Struct(request.Body); err != nil {
		return Login400JSONResponse{Detail: mapValidationError(err)}, nil
	}

	input := dto.LoginInput{
		Email:    string(request.Body.Email),
		Password: request.Body.Password,
	}

	pair, err := h.useCase.Login(ctx, input)
	if err != nil {
		detail, status := mapAuthError(err)
		if status == 401 {
			return Login401JSONResponse{Detail: detail}, nil
		}
		return Login400JSONResponse{Detail: detail}, nil
	}

	return Login200JSONResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
	}, nil
}

func (h *authHandler) Refresh(ctx context.Context, request RefreshRequestObject) (RefreshResponseObject, error) {
	if err := validate.Struct(request.Body); err != nil {
		return Refresh400JSONResponse{Detail: mapValidationError(err)}, nil
	}

	pair, err := h.useCase.Refresh(ctx, request.Body.RefreshToken)
	if err != nil {
		detail, status := mapAuthError(err)
		if status == 401 {
			return Refresh401JSONResponse{Detail: detail}, nil
		}
		return Refresh400JSONResponse{Detail: detail}, nil
	}

	return Refresh200JSONResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
	}, nil
}
