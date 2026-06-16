package app

import (
	"context"
	"time"

	"github.com/PRO-Robotech/sgroups/internal/sg-agent/config"
	api "github.com/PRO-Robotech/sgroups/internal/sg-agent/transport/agent/grpc/service"
	"github.com/PRO-Robotech/sgroups/internal/shared/app"

	"github.com/H-BF/corlib/server"
	"github.com/H-BF/corlib/server/interceptors"
	serverPrometheusMetrics "github.com/H-BF/corlib/server/metrics/prometheus"
	"github.com/go-openapi/spec"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	// HandleMetrics -
	HandleMetrics = "metrics"

	// HandleHealthcheck -
	HandleHealthcheck = "healthcheck"

	// HandleDebug -
	HandleDebug = "debug"
)

// WhenSetupApiServer -
func WhenSetupApiServer(ctx context.Context, f func(*server.APIServer) error) error {
	var (
		ssCacheTTL      time.Duration
		nftCacheTTL     time.Duration
		nftSyncInterval time.Duration
		nftScanStrategy config.ScanStrategy
	)
	ssStrtegy, err := config.SocketScanStrategy.Value(ctx)
	if err != nil {
		return err
	}
	nftScanStrategy, err = config.NftScanStrategy.Value(ctx)
	if err != nil {
		return err
	}
	if ssStrtegy == config.ScanStrategyCached {
		if ssCacheTTL, err = config.SocketScanCacheTTL.Value(ctx); err != nil {
			return err
		}
	}
	if nftScanStrategy == config.ScanStrategyCached {
		if nftCacheTTL, err = config.NftScanCacheTTL.Value(ctx); err != nil {
			return err
		}
	}
	if nftSyncInterval, err = config.NftScanSyncInterval.Value(ctx); err != nil {
		return err
	}
	svc := api.NewAgentService(ctx, ssCacheTTL, nftCacheTTL, nftSyncInterval)
	xOpt := server.WithGatewayOptions(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				EmitUnpopulated: true,
			},
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: false,
			},
		}),
	)

	var swaggerSpec *spec.Swagger
	if swaggerSpec, err = api.AgentSwaggerUtil.GetSpec(); err != nil {
		return err
	}

	opts := []server.APIServerOption{
		server.WithServices(svc),
		server.WithDocs(swaggerSpec, ""),
		xOpt,
	}
	app.WhenHaveMetricsRegistry(func(reg *prometheus.Registry) {
		pm := serverPrometheusMetrics.NewMetrics(
			serverPrometheusMetrics.WithSubsystem("grpc"),
			serverPrometheusMetrics.WithNamespace("server"),
		)
		if err = reg.Register(pm); err != nil {
			return
		}
		recovery := interceptors.NewRecovery(
			interceptors.RecoveryWithObservers(pm.PanicsObserver()),
		)
		opts = append(opts, server.WithRecovery(recovery))
		opts = append(opts, server.WithStatsHandlers(pm.StatHandlers()...))
	})
	var useTelemetryEp bool
	err = IfUseTelemetryAddr(ctx, func(_ string) error {
		useTelemetryEp = true
		return nil
	})
	if err != nil {
		return err
	}

	if !useTelemetryEp {
		opts = append(opts, setupTelemetryEndpoints(ctx)...)
	}

	err = whenAuthn(ctx, func(at authnType) error {
		switch a := at.(type) {
		case authnTLS:
			opts = append(opts, server.WithTLS(a.conf))
		default:
			return errors.Errorf("unsupported auth '%T'", a)
		}
		return nil
	})
	if err != nil {
		return err
	}

	var srv *server.APIServer
	if srv, err = server.NewAPIServer(opts...); err == nil {
		err = f(srv)
	}
	return err
}
