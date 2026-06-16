package app

import (
	"github.com/SergeiGD/testify-profile/config"
	"github.com/SergeiGD/testify-profile/internal/server"
)

type App struct {
	cfg        *config.Config
	HttpServer server.IServer
}

func NewApp(cfg *config.Config, httpServer server.IServer) *App {
	return &App{
		cfg:        cfg,
		HttpServer: httpServer,
	}
}
