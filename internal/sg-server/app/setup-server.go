package sgserver

import (
	"context"

	"github.com/PRO-Robotech/sgroups/internal/sg-server/transport/grpc/interceptors/validators"
	"github.com/PRO-Robotech/sgroups/internal/sg-server/transport/grpc/service"
	agApi "github.com/PRO-Robotech/sgroups/internal/sg-server/transport/grpc/service/ag"
	hostApi "github.com/PRO-Robotech/sgroups/internal/sg-server/transport/grpc/service/host"
	hbApi "github.com/PRO-Robotech/sgroups/internal/sg-server/transport/grpc/service/host-binding"
	nsApi "github.com/PRO-Robotech/sgroups/internal/sg-server/transport/grpc/service/namespace"
	nwApi "github.com/PRO-Robotech/sgroups/internal/sg-server/transport/grpc/service/network"
	nbApi "github.com/PRO-Robotech/sgroups/internal/sg-server/transport/grpc/service/network-binding"
	rlApi "github.com/PRO-Robotech/sgroups/internal/sg-server/transport/grpc/service/rules"
	svcApi "github.com/PRO-Robotech/sgroups/internal/sg-server/transport/grpc/service/service"
	sbApi "github.com/PRO-Robotech/sgroups/internal/sg-server/transport/grpc/service/service-binding"
	statusApi "github.com/PRO-Robotech/sgroups/internal/sg-server/transport/grpc/service/status"
	"github.com/PRO-Robotech/sgroups/internal/shared/app"

	config "github.com/H-BF/corlib/pkg/plain-config"
	"github.com/H-BF/corlib/server"
	"github.com/H-BF/corlib/server/interceptors"
	serverPrometheusMetrics "github.com/H-BF/corlib/server/metrics/prometheus"
	"github.com/go-openapi/spec"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

// SetupSgServer -
func SetupSgServer(ctx context.Context) (*server.APIServer, error) {
	srv, err := setupSgServices(ctx)
	if err != nil {
		return nil, err
	}
	doc, err := getServicesDoc()
	if err != nil {
		return nil, err
	}
	xOpt := server.WithGatewayOptions(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				EmitUnpopulated: true,
			},
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: false, //we fail when find unknown field in request
			},
		}),
	)
	opts := []server.APIServerOption{
		server.WithServices(srv...),
		server.WithDocs(doc, ""),
		xOpt,
	}

	if useBufProtoValidators, _ := ServerUseBufProtoValidator.Value(ctx); useBufProtoValidators {
		opts = append(opts,
			server.WithUnaryInterceptors(validators.BufProtovalidateUnary()),
			server.WithStreamInterceptors(validators.BufProtovalidateStream()),
		)
	}

	opts = append(opts,
		server.WithUnaryInterceptors(validators.ProtoCrossValidateUnary()),
		server.WithStreamInterceptors(validators.ProtoCrossValidateStream()),
	)

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
		promHandler := promhttp.InstrumentMetricHandler(
			reg,
			promhttp.HandlerFor(reg, promhttp.HandlerOpts{}),
		)
		opts = append(opts, server.WithHttpHandler("/"+HandleMetrics, promHandler))
	})
	if err != nil {
		return nil, err
	}
	if hc, _ := HealthcheckEnable.Value(ctx); hc {
		opts = append(opts, server.WithHttpHandler("/"+HandleHealthcheck, app.HcHandler{}))
	}
	opts = append(opts, server.WithHttpHandler("/"+HandleDebug, app.PProfHandler()))

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
		return nil, err
	}
	return server.NewAPIServer(opts...)
}

func setupSgServices(ctx context.Context) ([]server.APIService, error) {
	var opts []service.Option
	if o, e := ServerAPIpathPrefix.Value(ctx); e == nil {
		opts = append(opts, service.WithAPIpathPrefixes(o))
	} else if !errors.Is(e, config.ErrNotFound) {
		return nil, e
	}
	if o, e := ServerAdditionalAPIpaths.Value(ctx); e == nil {
		opts = append(opts, service.WithAdditionalServiceNames(o...))
	} else if !errors.Is(e, config.ErrNotFound) {
		return nil, e
	}
	srv := []server.APIService{
		nsApi.NewSgNamespaceService(ctx, getAppRepository(), opts...),
		agApi.NewSgAddressGroupService(ctx, getAppRepository(), opts...),
		nwApi.NewSgNetworkService(ctx, getAppRepository(), opts...),
		hostApi.NewSgHostService(ctx, getAppRepository(), opts...),
		hbApi.NewSgHostBindingService(ctx, getAppRepository(), opts...),
		nbApi.NewSgNetworkBindingService(ctx, getAppRepository(), opts...),
		svcApi.NewSgServiceService(ctx, getAppRepository(), opts...),
		sbApi.NewSgServiceBindingService(ctx, getAppRepository(), opts...),
		rlApi.NewSgRulesService(ctx, getAppRepository(), opts...),
		statusApi.NewSgStatusService(ctx, getAppRepository(), opts...),
	}
	return srv, nil
}

func getServicesDoc() (*spec.Swagger, error) {
	var (
		doc *spec.Swagger
		err error
	)
	doc, err = agApi.SgAddressGroupSwaggerUtil.GetSpec()
	if err != nil {
		return nil, errors.Wrap(err, "get service doc")
	}

	err = server.ComposeSwaggers(doc)

	return doc, err
}
