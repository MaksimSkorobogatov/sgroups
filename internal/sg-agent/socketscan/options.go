package socketscan

import (
	"time"

	"github.com/H-BF/corlib/pkg/filter"
	"github.com/pkg/errors"
)

type (
	// SSopt socket scan options interface
	SSopt interface {
		apply(*ssOpts) error
	}
	ssOpts struct {
		scope    filter.Scope
		cacheTTL time.Duration
	}
	ssOptsF func(*ssOpts) error
)

func (optf ssOptsF) apply(i *ssOpts) error {
	return optf(i)
}

// WithScope - set scope by socket selectors. If not set, all sockets are included.
func WithScope(scope filter.Scope) SSopt {
	return ssOptsF(func(o *ssOpts) error {
		o.scope = scope
		return nil
	})
}

// WithCached - set cache TTL for socket scan results.
func WithCached(ttl time.Duration) SSopt {
	return ssOptsF(func(o *ssOpts) error {
		if ttl < time.Second {
			return errors.Errorf("'ss ttl' is (%v) less than 1s", ttl)
		}
		o.cacheTTL = ttl
		return nil
	})
}
