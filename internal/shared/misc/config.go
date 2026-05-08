package misc

import (
	"context"

	"github.com/H-BF/corlib/pkg/option"
	plain_config "github.com/H-BF/corlib/pkg/plain-config"
	"github.com/pkg/errors"
)

// GetOptionalConfig ignores ErrNotFound
func GetOptionalConfig[T any](ctx context.Context, cfg plain_config.ValueT[T]) (opt option.ValueOf[T], err error) {
	var val T
	if val, err = cfg.Value(ctx); err == nil {
		opt.Set(val)
	} else if errors.Is(err, plain_config.ErrNotFound) {
		err = nil
	}
	return opt, err
}
