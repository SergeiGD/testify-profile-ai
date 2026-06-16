package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/SergeiGD/testify-profile/config"
	http2 "github.com/SergeiGD/testify-profile/internal/api/http"
	pgadapter "github.com/SergeiGD/testify-profile/internal/adapter/postgres"
	smtpadapter "github.com/SergeiGD/testify-profile/internal/adapter/smtp"
	"github.com/SergeiGD/testify-profile/internal/app"
	"github.com/SergeiGD/testify-profile/internal/server/httpserv"
	"github.com/SergeiGD/testify-profile/internal/services/clock"
	"github.com/SergeiGD/testify-profile/internal/services/linkbuilder"
	"github.com/SergeiGD/testify-profile/internal/services/password"
	"github.com/SergeiGD/testify-profile/internal/services/token"
	"github.com/SergeiGD/testify-profile/internal/usecases/register"
	"github.com/SergeiGD/testify-profile/pkg/logger"
	"github.com/SergeiGD/testify-profile/pkg/postgres"
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

	log := logger.NewLogger(&cfg)

	dbPool, err := postgres.NewClient(ctx, cfg)
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	defer dbPool.Close()

	userRepo := pgadapter.NewUserRepository(dbPool)
	confirmationRepo := pgadapter.NewConfirmationRepository(dbPool)

	var emailSender smtpadapter.EmailSender
	if cfg.SMTP.UseMockSender {
		emailSender = smtpadapter.NewMockEmailSender(log.Entry)
	} else {
		emailSender = smtpadapter.NewEmailSender(
			cfg.SMTP.Host,
			cfg.SMTP.Port,
			cfg.SMTP.Username,
			cfg.SMTP.Password,
			cfg.SMTP.From,
		)
	}

	passwordHasher := password.NewPasswordHasher()
	tokenSvc := token.NewTokenService()
	lb := linkbuilder.NewLinkBuilder(cfg.App.BaseURL)
	clk := clock.NewRealClock()

	registerUseCase := register.NewRegisterUseCase(
		userRepo,
		confirmationRepo,
		emailSender,
		passwordHasher,
		tokenSvc,
		lb,
		clk,
		cfg.Registration.TokenTTL,
		cfg.Registration.ResendCooldown,
	)

	handler := http2.NewServerHandler(registerUseCase)

	application := app.NewApp(
		&cfg,
		httpserv.NewHttpServer(&cfg, log, handler),
	)

	eg := errgroup.Group{}
	eg.Go(func() error { return application.HttpServer.Run(ctx) })

	return eg.Wait()
}
