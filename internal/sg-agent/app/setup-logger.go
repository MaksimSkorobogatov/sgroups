package app

import (
	conf "github.com/PRO-Robotech/sgroups/internal/sg-agent/config"
	"github.com/PRO-Robotech/sgroups/internal/shared/app"

	"github.com/H-BF/corlib/logger"
	"github.com/pkg/errors"
)

// SetupLogger setup app logger
func SetupLogger() error {
	ctx := app.Context()
	_, err := conf.AppLoggerLevel.Value(ctx, conf.AppLoggerLevel.OptSink(func(v logger.LoggerLevelConf) error {
		var l logger.LogLevel
		if e := l.UnmarshalText([]byte(v)); e != nil {
			return errors.Wrapf(e, "recognize '%s' logger level from config", v)
		}
		logger.SetLevel(l)
		return nil
	}))
	return err
}
