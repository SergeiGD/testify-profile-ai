package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/SergeiGD/testify-profile/config"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Client interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Begin(ctx context.Context) (pgx.Tx, error)
}

func NewClient(ctx context.Context, pgConf config.Config) (*pgxpool.Pool, error) {
	dns := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", pgConf.Database.User, pgConf.Database.Password, pgConf.Database.Host, pgConf.Database.Port, pgConf.Database.Db)

	var pool *pgxpool.Pool
	var err error

	maxAttemps := pgConf.Database.MaxAttemps

	for maxAttemps > 0 && pool == nil {
		ctx, cancel := context.WithTimeout(ctx, pgConf.Database.Timeout)
		defer cancel()

		pool, err = pgxpool.New(ctx, dns)
		if err != nil {
			time.Sleep(pgConf.Database.ConnDelay)
			maxAttemps--
			continue
		}

	}

	if maxAttemps == 0 && err != nil {
		return nil, err
	}

	return pool, nil

}
