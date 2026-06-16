package server

import "context"

type IServer interface {
	Run(ctx context.Context) error
}
