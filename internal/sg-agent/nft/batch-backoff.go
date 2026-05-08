package nft

import (
	"context"
	"time"

	"github.com/H-BF/corlib/pkg/backoff"
)

func makeBatchBackoff(ctx context.Context) backoff.Backoff {
	const (
		mul  = 1.3
		rand = 0
	)
	bk := backoff.ExponentialBackoffBuilder().
		WithMultiplier(mul).
		WithRandomizationFactor(rand)
	if dl, ok := ctx.Deadline(); ok {
		d := time.Until(dl)
		if d > 0 {
			bk = bk.WithMaxElapsedThreshold(d)
		}
	}
	return backoff.WithContext(
		bk.Build(), ctx,
	)
}
