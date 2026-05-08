package job

import (
	"context"

	"github.com/PRO-Robotech/sgroups/internal/shared/patterns"

	"github.com/H-BF/corlib/pkg/patterns/observer"
)

// Task -
type Task interface {
	Run(ctx context.Context) error
	Close() error
	Subject() observer.Subject
	MakeObserver(ctx context.Context) patterns.Observer
}
