//go:build wireinject

package di

import (
	"context"

	"github.com/SergeiGD/testify-profile/config"
	"github.com/SergeiGD/testify-profile/internal/app"
	"github.com/google/wire"
)

func InitializeApp(ctx context.Context, cfg *config.Config) (*app.App, func(), error) {
	wire.Build(AppSet)
	return nil, nil, nil
}
