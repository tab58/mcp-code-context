package resolver

import "context"

type Resolver interface {
	GetAppVersion() string
	Query(ctx context.Context, query string) (string, error)
}
