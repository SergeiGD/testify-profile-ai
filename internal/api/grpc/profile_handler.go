package grpc

import (
	"context"
	"errors"

	"github.com/SergeiGD/testify-profile/internal/api/grpc/pb"
	"github.com/SergeiGD/testify-profile/internal/domain"
	"github.com/SergeiGD/testify-profile/internal/usecases/profile"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ProfileHandler struct {
	pb.UnimplementedProfileServiceServer
	profileUseCase profile.ProfileUseCase
}

func NewProfileHandler(profileUseCase profile.ProfileUseCase) *ProfileHandler {
	return &ProfileHandler{profileUseCase: profileUseCase}
}

func (h *ProfileHandler) GetProfile(ctx context.Context, req *pb.GetProfileRequest) (*pb.GetProfileResponse, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid profile id")
	}

	user, err := h.profileUseCase.GetProfile(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "profile not found")
		}
		if errors.Is(err, domain.ErrAccountNotConfirmed) {
			return nil, status.Error(codes.NotFound, "profile not found")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &pb.GetProfileResponse{
		Id:        user.ID.String(),
		Email:     user.Email,
		Username:  user.Username,
		BirthDate: user.BirthDate.Format("2006-01-02"),
	}, nil
}
