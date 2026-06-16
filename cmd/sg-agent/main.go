package main

import (
	"context"
	"flag"
	"os"
	"time"

	. "github.com/PRO-Robotech/sgroups/internal/sg-agent/app"
	"github.com/PRO-Robotech/sgroups/internal/sg-agent/config"
	"github.com/PRO-Robotech/sgroups/internal/sg-agent/job"
	"github.com/PRO-Robotech/sgroups/internal/shared/app" //nolint:revive

	"github.com/H-BF/corlib/logger"
	pkgNet "github.com/H-BF/corlib/pkg/net"
	"github.com/H-BF/corlib/pkg/patterns/observer"
	conf "github.com/H-BF/corlib/pkg/plain-config"
	"github.com/H-BF/corlib/server"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func main() { //nolint:gocyclo
	flag.Parse()
	SetupContext()
	SetupAgentSubject()
	ctx := app.Context()
	logger.SetLevel(zap.InfoLevel)
	logger.InfoKV(ctx, "-= HELLO =-", "version", app.GetVersion())
	err := conf.InitGlobalConfig(
		conf.WithAcceptEnvironment{EnvPrefix: "NFT"},
		conf.WithSourceFile{FileName: ConfigFile},
		conf.WithEnvKeyReplacer{Old: []string{"-", "/"}, New: []string{"_", "_"}},

		conf.WithDefValue(config.ExitOnSuccess, false),
		conf.WithDefValue(config.ContinueOnFailure, true),
		conf.WithDefValue(config.ContinueAfterTimeout, "10s"),

		conf.WithDefValue(config.FqdnStrategy, config.FqdnRulesStartegyDNS),
		conf.WithDefValue(config.AppLoggerLevel, "DEBUG"),
		conf.WithDefValue(config.AppGracefulShutdown, 10*time.Second), //nolint:mnd
		conf.WithDefValue(config.NetNS, ""),
		conf.WithDefValue(config.DefPolicyAccept, false),
		conf.WithDefValue(config.DryRun, false),
		conf.WithDefValue(config.NetlinkWatcherLinger, "10s"),
		conf.WithDefValue(config.ServicesDefDialDuration, 10*time.Second), //nolint:mnd
		conf.WithDefValue(config.SGroupsAddress, "tcp://127.0.0.1:9000"),
		conf.WithDefValue(config.SGroupsSyncStatusInterval, "10s"),
		conf.WithDefValue(config.SGroupsSyncStatusPush, false),
		conf.WithDefValue(config.SGroupsUseJsonCodec, false),
		//DNS group
		conf.WithDefValue(config.DnsNameservers, `["8.8.8.8"]`),
		conf.WithDefValue(config.DnsProto, "udp"),
		conf.WithDefValue(config.DnsPort, 53),   //nolint:mnd
		conf.WithDefValue(config.DnsRetries, 3), //nolint:mnd
		conf.WithDefValue(config.DnsRetriesTmo, "1s"),
		conf.WithDefValue(config.DnsDialDuration, "3s"),
		conf.WithDefValue(config.DnsWriteDuration, "5s"),
		conf.WithDefValue(config.DnsReadDuration, "5s"),
		//telemetry group
		conf.WithDefValue(config.TelemetryAddr, "127.0.0.1:5000"),
		conf.WithDefValue(config.MetricsEnable, true),
		conf.WithDefValue(config.HealthcheckEnable, true),
		conf.WithDefValue(config.UserAgent, ""),
		conf.WithDefValue(config.ProfileEnable, true),
		conf.WithDefValue(config.NftablesCollectorMinFrequency, "10s"),
		//authn group
		conf.WithDefValue(config.SGroupsAuthnType, config.AuthnTypeNONE),
		//api
		conf.WithDefValue(config.ApiAddress, "tcp://127.0.0.1:5000"),
		conf.WithDefValue(config.ApiAuthnType, config.AuthnTypeNONE),
		// ss
		conf.WithDefValue(config.SocketScanStrategy, config.ScanStrategyCached),
		conf.WithDefValue(config.SocketScanCacheTTL, "2s"),
		// nft
		conf.WithDefValue(config.NftScanSyncInterval, "1s"),
		conf.WithDefValue(config.NftScanStrategy, config.ScanStrategyCached),
		conf.WithDefValue(config.NftScanCacheTTL, "2s"),
	)
	if err != nil {
		logger.Fatal(ctx, err)
	}

	if err = SetupLogger(); err != nil {
		logger.Fatal(ctx, errors.WithMessage(err, "setup logger"))
	}
	if err = SetupMetrics(ctx); err != nil {
		logger.Fatal(ctx, errors.WithMessage(err, "setup metrics"))
	}

	err = WhenSetupApiServer(ctx, func(srv *server.APIServer) error {
		ep, e := pkgNet.ParseEndpoint(config.ApiAddress.MustValue(ctx))
		if e != nil {
			return errors.WithMessagef(e, "parse API endpoint: %v", e)
		}
		go func() { //start API endpoint
			if e1 := srv.Run(ctx, ep); e1 != nil {
				logger.Fatalf(ctx, "API server is failed: %v", e1)
			}
		}()
		return nil
	})
	if err != nil {
		logger.Fatal(ctx, errors.WithMessage(err, "setup API server"))
	}

	err = IfUseTelemetryAddr(ctx, func(addr string) error {
		return WhenSetupTelemtryServer(ctx, func(srv *server.APIServer) error {
			ep, e := pkgNet.ParseEndpoint(addr)
			if e != nil {
				return errors.WithMessagef(e, "parse telemetry endpoint (%s): %v", addr, e)
			}
			go func() { //start telemetry endpoint
				if e1 := srv.Run(ctx, ep); e1 != nil {
					logger.Fatalf(ctx, "telemetry server is failed: %v", e1)
				}
			}()
			return nil
		})
	})
	if err != nil {
		logger.Fatal(ctx, errors.WithMessage(err, "setup telemetry server"))
	}

	if err = SetupDnsResolver(ctx); err != nil {
		logger.Fatal(ctx, errors.WithMessage(err, "setup DNS resolver"))
	}

	if exitOnSuccess := config.ExitOnSuccess.MustValue(ctx); exitOnSuccess {
		o := observer.NewObserver(exitOnSuccessHanler,
			true,
			job.AppliedConfEvent{},
		)
		AgentSubject().ObserversAttach(o)
	}

	AgentSubject().ObserversAttach(
		observer.NewObserver(agentMetricsObserver,
			false,
			job.AppliedConfEvent{},
			job.SyncStatusErrorEvent{},
			job.NetlinkErrorEvent{},
			job.DomainAddressesEvent{}),
	)

	if err = runTasks(ctx); err != nil {
		if !errors.Is(err, context.DeadlineExceeded) {
			logger.Fatal(ctx, err)
		}
	}

	logger.SetLevel(zap.InfoLevel)
	logger.Info(ctx, "-= BYE =-")
}

func exitOnSuccessHanler(ev observer.EventType) {
	switch ev.(type) {
	case job.AppliedConfEvent:
		os.Exit(0)
	}
	os.Exit(1)
}

func agentMetricsObserver(ev observer.EventType) {
	if metrics := GetAgentMetrics(); metrics != nil {
		switch o := ev.(type) {
		case job.AppliedConfEvent:
			metrics.ObserveApplyConfig()
		case job.SyncStatusErrorEvent:
			metrics.ObserveError(ESrcSgBakend)
		case job.NetlinkErrorEvent:
			metrics.ObserveError(ESrcNetWatcher)
		case job.DomainAddressesEvent:
			if o.DnsAnswer.Err != nil {
				metrics.ObserveError(ESrcDNS)
			}
		}
	}
}
