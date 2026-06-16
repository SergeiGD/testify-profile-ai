package httpserv

import (
	"context"

	"github.com/SergeiGD/testify-profile/config"
	http2 "github.com/SergeiGD/testify-profile/internal/api/http"
	"github.com/SergeiGD/testify-profile/pkg/logger"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	middleware "github.com/oapi-codegen/gin-middleware"
)

type HttpServer struct {
	cfg     *config.Config
	logger  *logger.Logger
	handler http2.StrictServerInterface
}

func NewHttpServer(cfg *config.Config, logger *logger.Logger, handler http2.StrictServerInterface) *HttpServer {
	return &HttpServer{
		cfg:     cfg,
		logger:  logger,
		handler: handler,
	}
}

func (s *HttpServer) Run(ctx context.Context) error {
	swagger, err := http2.GetSwagger()
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("error on init swagger schema")
		return err
	}

	r := gin.Default()

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"http://localhost:8081"}
	r.Use(cors.New(corsConfig))
	r.Use(middleware.OapiRequestValidator(swagger))

	strictHandler := http2.NewStrictHandler(s.handler, make([]http2.StrictMiddlewareFunc, 0))
	http2.RegisterHandlers(r, strictHandler)

	err = r.Run(":8080")
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("error on starting gin server")
		return err
	}

	return nil
}
