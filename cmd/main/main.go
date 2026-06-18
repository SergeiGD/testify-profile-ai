package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/SergeiGD/testify-profile/config"
	"github.com/SergeiGD/testify-profile/internal/di"
	"github.com/ilyakaznacheev/cleanenv"
	"golang.org/x/sync/errgroup"
)

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	var cfg config.Config
	if err := cleanenv.ReadConfig("config/config.yaml", &cfg); err != nil {
		return fmt.Errorf("parse config: %w", err)
	}

	application, cleanup, err := di.InitializeApp(ctx, &cfg)
	if err != nil {
		return fmt.Errorf("initialize app: %w", err)
	}
	defer cleanup()

	eg := errgroup.Group{}
	eg.Go(func() error { return application.HttpServer.Run(ctx) })
	eg.Go(func() error { return application.GrpcServer.Run(ctx) })

	return eg.Wait()
}
