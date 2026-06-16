package app

import (
	"context"

	conf "github.com/PRO-Robotech/sgroups/internal/sg-agent/config"
	"github.com/PRO-Robotech/sgroups/internal/shared/app"

	"github.com/H-BF/corlib/server"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// WhenSetupTelemtryServer -
func WhenSetupTelemtryServer(ctx context.Context, f func(*server.APIServer) error) error {
	var (
		opts []server.APIServerOption
		err  error
	)
	opts = append(opts, setupTelemetryEndpoints(ctx)...)
	if len(opts) == 0 {
		return nil
	}
	var srv *server.APIServer
	if srv, err = server.NewAPIServer(opts...); err == nil {
		err = f(srv)
	}
	return err
}

// IfUseTelemetryAddr -
func IfUseTelemetryAddr(ctx context.Context, consume func(string) error) error {
	var (
		tlAddr  string
		apiAddr string
		isEq    bool
		err     error
	)
	tlAddr, err = conf.TelemetryAddr.Value(ctx)
	if err != nil && errors.Is(err, conf.ErrNotFound) {
		return nil
	} else if err != nil {
		return err
	}
	apiAddr, err = conf.ApiAddress.Value(ctx)
	if err != nil {
		return err
	}
	isEq, err = conf.AddrsEqual(tlAddr, apiAddr)
	if err != nil {
		return err
	}
	if isEq {
		return nil
	}

	return consume(tlAddr)
}

func setupTelemetryEndpoints(ctx context.Context) (opts []server.APIServerOption) {
	app.WhenHaveMetricsRegistry(func(reg *prometheus.Registry) {
		opts = append(opts,
			server.WithHttpHandler("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg})),
		)
	})
	if hc, _ := conf.HealthcheckEnable.Value(ctx); hc {
		opts = append(opts, server.WithHttpHandler("/healthcheck", app.HcHandler{}))
	}
	if p, _ := conf.ProfileEnable.Value(ctx); p {
		opts = append(opts, server.WithHttpHandler("/debug", app.PProfHandler()))
	}
	return opts
}
