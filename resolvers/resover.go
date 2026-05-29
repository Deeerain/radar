package resolvers

import "context"

type Resolver interface {
	Resolve(ctx context.Context) (string, error)
}
