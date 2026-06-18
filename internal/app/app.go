package app

import (
	"github.com/SergeiGD/testify-profile/config"
	"github.com/SergeiGD/testify-profile/internal/server"
	"github.com/SergeiGD/testify-profile/internal/server/grpcserv"
	"github.com/SergeiGD/testify-profile/internal/server/httpserv"
)

type App struct {
	cfg        *config.Config
	HttpServer server.IServer
	GrpcServer server.IServer
}

func NewApp(cfg *config.Config, httpServer *httpserv.HttpServer, grpcServer *grpcserv.GrpcServer) *App {
	return &App{
		cfg:        cfg,
		HttpServer: httpServer,
		GrpcServer: grpcServer,
	}
}
