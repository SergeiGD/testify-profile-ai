package di

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strings"

	"github.com/SergeiGD/testify-profile/config"
	pgadapter "github.com/SergeiGD/testify-profile/internal/adapter/postgres"
	smtpadapter "github.com/SergeiGD/testify-profile/internal/adapter/smtp"
	"github.com/SergeiGD/testify-profile/internal/app"
	http2 "github.com/SergeiGD/testify-profile/internal/api/http"
	grpcapi "github.com/SergeiGD/testify-profile/internal/api/grpc"
	"github.com/SergeiGD/testify-profile/internal/server/grpcserv"
	"github.com/SergeiGD/testify-profile/internal/server/httpserv"
	"github.com/SergeiGD/testify-profile/internal/usecases/profile"
	"github.com/SergeiGD/testify-profile/internal/services/clock"
	jwtSvc "github.com/SergeiGD/testify-profile/internal/services/jwt"
	"github.com/SergeiGD/testify-profile/internal/services/linkbuilder"
	"github.com/SergeiGD/testify-profile/internal/services/password"
	"github.com/SergeiGD/testify-profile/internal/services/token"
	"github.com/SergeiGD/testify-profile/internal/usecases/auth"
	"github.com/SergeiGD/testify-profile/internal/usecases/register"
	"github.com/SergeiGD/testify-profile/pkg/logger"
	pkgpostgres "github.com/SergeiGD/testify-profile/pkg/postgres"
	"github.com/google/wire"
	pgxpool "github.com/jackc/pgx/v5/pgxpool"
)

var InfraSet = wire.NewSet(
	ProvidePostgresPool,
	wire.Bind(new(pkgpostgres.Client), new(*pgxpool.Pool)),
)

var RepositorySet = wire.NewSet(
	pgadapter.NewUserRepository,
	pgadapter.NewConfirmationRepository,
)

var ServiceSet = wire.NewSet(
	logger.NewLogger,
	password.NewPasswordHasher,
	token.NewTokenService,
	clock.NewRealClock,
	ProvideLinkBuilder,
	ProvideEmailSender,
	ProvideRSAPrivateKey,
	ProvideRSAPublicKey,
	ProvideJWTService,
)

var UseCaseSet = wire.NewSet(
	ProvideRegisterUseCase,
	auth.NewAuthUseCase,
	profile.NewProfileUseCase,
)

var ServerSet = wire.NewSet(
	http2.NewServerHandler,
	httpserv.NewHttpServer,
	grpcapi.NewProfileHandler,
	grpcserv.NewGrpcServer,
)

var AppSet = wire.NewSet(
	app.NewApp,
	InfraSet,
	RepositorySet,
	ServiceSet,
	UseCaseSet,
	ServerSet,
)

func ProvidePostgresPool(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, func(), error) {
	pool, err := pkgpostgres.NewClient(ctx, *cfg)
	if err != nil {
		return nil, nil, err
	}
	return pool, pool.Close, nil
}

func ProvideEmailSender(cfg *config.Config, log *logger.Logger) smtpadapter.EmailSender {
	if cfg.SMTP.UseMockSender {
		return smtpadapter.NewMockEmailSender(log.Entry)
	}
	return smtpadapter.NewEmailSender(
		cfg.SMTP.Host,
		cfg.SMTP.Port,
		cfg.SMTP.Username,
		cfg.SMTP.Password,
		cfg.SMTP.From,
	)
}

func ProvideRSAPrivateKey(cfg *config.Config) (*rsa.PrivateKey, error) {
	data := []byte(strings.ReplaceAll(cfg.JWT.PrivateKey, `\n`, "\n"))
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block for private key")
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}
	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA private key")
	}
	return rsaKey, nil
}

func ProvideRSAPublicKey(cfg *config.Config) (*rsa.PublicKey, error) {
	data := []byte(strings.ReplaceAll(cfg.JWT.PublicKey, `\n`, "\n"))
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block for public key")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse public key: %w", err)
	}
	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}
	return rsaPub, nil
}

func ProvideJWTService(privateKey *rsa.PrivateKey, publicKey *rsa.PublicKey, cfg *config.Config) jwtSvc.JWTService {
	return jwtSvc.NewJWTService(privateKey, publicKey, cfg.JWT.AccessTTL, cfg.JWT.RefreshTTL)
}

func ProvideLinkBuilder(cfg *config.Config) linkbuilder.LinkBuilder {
	return linkbuilder.NewLinkBuilder(cfg.App.BaseURL)
}

func ProvideRegisterUseCase(
	userRepo pgadapter.UserRepository,
	confirmationRepo pgadapter.ConfirmationRepository,
	emailSender smtpadapter.EmailSender,
	passwordHasher password.PasswordHasher,
	tokenService token.TokenService,
	lb linkbuilder.LinkBuilder,
	clk clock.Clock,
	cfg *config.Config,
) register.RegisterUseCase {
	return register.NewRegisterUseCase(
		userRepo,
		confirmationRepo,
		emailSender,
		passwordHasher,
		tokenService,
		lb,
		clk,
		cfg.Registration.TokenTTL,
		cfg.Registration.ResendCooldown,
	)
}
