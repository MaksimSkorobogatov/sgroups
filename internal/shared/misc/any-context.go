package misc

import "context"

// AnyContext returns a context that is canceled when any of the given
// parent contexts is canceled. Cancellation is propagated with the cause
// from the originating parent; the returned CancelFunc detaches the
// wiring and cancels the returned context with context.Canceled.
func AnyContext(parents ...context.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancelCause(context.Background())

	stops := make([]func() bool, len(parents))
	for i, parent := range parents {
		stops[i] = context.AfterFunc(parent, func() {
			cancel(context.Cause(parent))
		})
	}

	cancelGroup := func() {
		for _, stop := range stops {
			stop()
		}
		cancel(context.Canceled)
	}

	return ctx, cancelGroup
}
