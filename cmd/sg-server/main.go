package main

import (
	. "github.com/PRO-Robotech/sgroups/internal/sg-server/app" //nolint:revive
	"github.com/PRO-Robotech/sgroups/internal/shared/app"

	_ "github.com/H-BF/corlib/app/identity"
	"github.com/H-BF/corlib/logger"
	pkgNet "github.com/H-BF/corlib/pkg/net"
	config "github.com/H-BF/corlib/pkg/plain-config"
	"github.com/H-BF/corlib/server"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

func main() {
	SetupContext()
	ctx := app.Context()
	logger.SetLevel(zap.InfoLevel)
	logger.InfoKV(ctx, "-= HELLO =-", "version", app.GetVersion())
	err := config.InitGlobalConfig(
		config.WithAcceptEnvironment{EnvPrefix: "SG"},
		config.WithSourceFile{FileName: ConfigFile},
		config.WithEnvKeyReplacer{Old: []string{"-", "/"}, New: []string{"_", "_"}},

		config.WithDefValue(LoggerLevel, logger.LoggerLevelDEBUG),
		config.WithDefValue(MetricsEnable, true),
		config.WithDefValue(HealthcheckEnable, true),
		config.WithDefValue(ServerGracefulShutdown, "10s"),
		config.WithDefValue(ServerEndpoint, "tcp://127.0.0.1:9000"),
		config.WithDefValue(ServerUseBufProtoValidator, false),
		config.WithDefValue(StorageType, "POSTGRES"),
		config.WithDefValue(AuthnType, config.AuthnTypeNONE),
	)
	if err != nil {
		logger.Fatal(ctx, err)
	}
	if err = SetupLogger(); err != nil {
		logger.Fatal(ctx, errors.WithMessage(err, "when setup logger"))
	}

	if MetricsEnable.MustValue(ctx) {
		opt := app.AddMetrics{
			Metrics: []prometheus.Collector{app.NewHealthcheckMetric(nil)},
		}
		err = app.SetupMetrics(opt)
		if err != nil {
			logger.Fatal(ctx, errors.WithMessage(err, "when setup metrics"))
		}
	}

	var ep *pkgNet.Endpoint
	_, err = ServerEndpoint.Value(ctx, ServerEndpoint.OptSink(func(v string) error {
		var e error
		if ep, e = pkgNet.ParseEndpoint(v); e != nil {
			logger.Fatalf(ctx, "parse server endpoint (%s): %v", v, err)
		}
		return nil
	}))
	if err != nil && errors.Is(err, config.ErrNotFound) {
		logger.Fatal(ctx, errors.WithMessage(err, "server endpoint is absent"))
	}
	if err = SetupRepository(); err != nil {
		logger.Fatal(ctx, errors.WithMessage(err, "on opening db storage"))
	}
	var srv *server.APIServer
	if srv, err = SetupSgServer(ctx); err != nil {
		logger.Fatalf(ctx, "setup server: %v", err)
	}
	gracefulDuration, _ := ServerGracefulShutdown.Value(ctx)
	app.SetHealthState(true)
	if err = srv.Run(ctx, ep, server.RunWithGracefulStop(gracefulDuration)); err != nil {
		logger.Fatalf(ctx, "run server: %v", err)
	}
	logger.SetLevel(zap.InfoLevel)
	logger.Info(ctx, "-= BYE =-")
}
