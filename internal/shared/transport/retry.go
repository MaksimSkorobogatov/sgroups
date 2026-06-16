package transport

import (
	"context"
	"time"

	"github.com/PRO-Robotech/sgroups/internal/shared/misc"

	"github.com/H-BF/corlib/logger"
	"github.com/H-BF/corlib/pkg/backoff"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Retry -
func Retry(ctx context.Context, api string, fn func() error) (err error) {
	defer func() {
		err = errors.WithMessage(err, api)
	}()
	const maxAttempts = 3
	log := logger.FromContext(ctx)
	var attempt int
	for bk := makeRetryBackoff(ctx); ; attempt++ {
		err = fn()
		if err == nil {
			return nil
		}
		if isPermanentError(err) {
			return errors.WithMessage(err, "not retryable error")
		}
		if attempt >= maxAttempts {
			return errors.WithMessagef(err, "too many retries: (%d)", attempt)
		}
		pause := bk.NextBackOff()
		if pause == backoff.Stop {
			return errors.WithMessagef(err, "too many retries: (%d)", attempt+1)
		}
		log.Warnf("retry for %s, due to err: %v; attempt %d; will retry after %v", api, err, attempt+1, pause)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(pause):
		}
	}
}

func isPermanentError(err error) bool {
	return misc.IsIn(status.Code(err),
		codes.InvalidArgument, codes.NotFound, codes.AlreadyExists,
		codes.PermissionDenied, codes.Unauthenticated, codes.Unimplemented,
		codes.FailedPrecondition, codes.OutOfRange,
	)
}

func makeRetryBackoff(ctx context.Context) backoff.Backoff {
	const (
		initInt    = 200 * time.Millisecond
		mul        = 1.5
		rand       = 0.5
		maxInt     = 2 * time.Second
		maxElapsed = 10 * time.Second
	)
	bk := backoff.ExponentialBackoffBuilder().
		WithInitialInterval(initInt).
		WithMultiplier(mul).
		WithRandomizationFactor(rand).
		WithMaxInterval(maxInt).
		WithMaxElapsedThreshold(maxElapsed)
	if dl, ok := ctx.Deadline(); ok {
		if d := time.Until(dl); d > 0 && d < maxElapsed {
			bk = bk.WithMaxElapsedThreshold(d)
		}
	}
	b := bk.Build()
	b.Reset()
	return backoff.WithContext(b, ctx)
}
