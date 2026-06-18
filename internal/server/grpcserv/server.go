package grpcserv

import (
	"context"
	"net"

	grpcapi "github.com/SergeiGD/testify-profile/internal/api/grpc"
	"github.com/SergeiGD/testify-profile/internal/api/grpc/pb"
	"github.com/SergeiGD/testify-profile/config"
	"github.com/SergeiGD/testify-profile/pkg/logger"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const apiKeyHeader = "x-api-key"

type GrpcServer struct {
	cfg     *config.Config
	logger  *logger.Logger
	handler *grpcapi.ProfileHandler
}

func NewGrpcServer(cfg *config.Config, logger *logger.Logger, handler *grpcapi.ProfileHandler) *GrpcServer {
	return &GrpcServer{
		cfg:     cfg,
		logger:  logger,
		handler: handler,
	}
}

func (s *GrpcServer) Run(ctx context.Context) error {
	lis, err := net.Listen("tcp", ":"+s.cfg.GRPC.Port)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("failed to listen for gRPC server")
		return err
	}

	srv := grpc.NewServer(grpc.UnaryInterceptor(s.apiKeyInterceptor))
	pb.RegisterProfileServiceServer(srv, s.handler)

	errCh := make(chan error, 1)
	go func() {
		s.logger.WithFields(logrus.Fields{
			"port": s.cfg.GRPC.Port,
		}).Info("gRPC server started")
		if err := srv.Serve(lis); err != nil {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		srv.GracefulStop()
		return nil
	case err := <-errCh:
		s.logger.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("gRPC server error")
		return err
	}
}

func (s *GrpcServer) apiKeyInterceptor(
	ctx context.Context,
	req any,
	_ *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing metadata")
	}

	values := md.Get(apiKeyHeader)
	if len(values) == 0 || values[0] != s.cfg.GRPC.APIKey {
		return nil, status.Error(codes.Unauthenticated, "invalid api key")
	}

	return handler(ctx, req)
}
