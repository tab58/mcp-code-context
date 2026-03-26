package route

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
)

type Handler[I, O any] func(ctx context.Context, req *I) (*O, error)

type RegisterArgs[I, O any] struct {
	API       huma.API
	Operation huma.Operation
	Handler   Handler[I, O]
}

// Register registers a route with the given API and operation.
// It is a wrapper around huma.Register that allows for other things to be placed in handlers,
// such as authentication information, logging, metrics, etc.
func Register[I, O any](args RegisterArgs[I, O]) {
	api := args.API
	op := args.Operation
	handler := args.Handler

	huma.Register(api, op, handler)
}
