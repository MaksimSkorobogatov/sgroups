package app

import (
	"context"

	"github.com/PRO-Robotech/sgroups/internal/shared/app"

	"github.com/H-BF/corlib/logger"
	"github.com/H-BF/corlib/pkg/signals"
	"go.uber.org/zap"
)

// SetupContext -
func SetupContext() {
	ctx, cancel := context.WithCancel(context.Background()) //nolint:gosec
	signals.WhenSignalExit(func() error {
		logger.SetLevel(zap.InfoLevel)
		logger.Info(ctx, "caught application stop signal")
		cancel()
		return nil
	})
	app.SetContext(ctx)
}
