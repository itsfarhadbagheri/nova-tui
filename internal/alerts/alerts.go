package alerts

import "context"

type Analyzer interface {
	Start(ctx context.Context) (any, error)
}
