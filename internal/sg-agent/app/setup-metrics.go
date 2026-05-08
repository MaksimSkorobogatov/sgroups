package app

import (
	"context"
	"os"
	"time"

	"github.com/PRO-Robotech/sgroups/internal/sg-agent/config"
	"github.com/PRO-Robotech/sgroups/internal/shared/app"

	"github.com/H-BF/corlib/logger"
	"github.com/H-BF/corlib/pkg/atomic"
	nftconf "github.com/H-BF/corlib/pkg/nftables"
	nfmetrics "github.com/H-BF/corlib/pkg/nftables/prometheus"
	"github.com/google/nftables"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/vishvananda/netns"
)

// AgentMetrics -
type AgentMetrics struct { //nolint
	appliedConfigCount prometheus.Counter
	errorCount         *prometheus.CounterVec
}

var agentMetricsHolder atomic.Value[*AgentMetrics]

const (
	labelUserAgent = "user_agent"
	labelHostName  = "host_name"
	labelSource    = "source"
	nsAgent        = "agent"
)

const ( // error sources
	// ESrcDNS -
	ESrcDNS = "dns"

	// ESrcNetWatcher -
	ESrcNetWatcher = "net-watcher"

	// ESrcSgBakend -
	ESrcSgBakend = "sgroups-svc"
)

// SetupMetrics -
func SetupMetrics(ctx context.Context) error {
	if !config.MetricsEnable.MustValue(ctx) {
		return nil
	}
	hostname, err := os.Hostname()
	if err != nil {
		return err
	}
	labels := prometheus.Labels{
		labelUserAgent: config.UserAgent.MustValue(ctx),
		labelHostName:  hostname,
	}
	am := new(AgentMetrics)
	am.init(labels)
	metricsOpt := app.AddMetrics{
		Metrics: []prometheus.Collector{
			app.NewHealthcheckMetric(labels),
			am.appliedConfigCount,
			am.errorCount,
		},
	}
	err = ifNetfilterMetricsCollector(ctx, "nftables-metrics", func(c prometheus.Collector) {
		metricsOpt.Metrics = append(metricsOpt.Metrics, c)
	})
	if err != nil {
		return errors.WithMessage(err, "on setup 'nftables' metrics collector")
	}
	if err = app.SetupMetrics(metricsOpt); err == nil {
		agentMetricsHolder.Store(am, nil)
	}
	return err
}

// GetAgentMetrics -
func GetAgentMetrics() *AgentMetrics {
	v, _ := agentMetricsHolder.Load()
	return v
}

func (am *AgentMetrics) init(labels prometheus.Labels) {
	am.appliedConfigCount = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace:   nsAgent,
		Name:        "applied_configs",
		Help:        "count of successfully applied configurations",
		ConstLabels: labels,
	})
	am.errorCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace:   nsAgent,
		Name:        "errors",
		Help:        "count of errors",
		ConstLabels: labels,
	}, []string{labelSource})
}

// ObserveError -
func (am *AgentMetrics) ObserveError(errSource string) {
	am.errorCount.WithLabelValues(errSource).Inc()
}

// ObserveApplyConfig -
func (am *AgentMetrics) ObserveApplyConfig() {
	am.appliedConfigCount.Inc()
}

func ifNetfilterMetricsCollector(ctx context.Context, logSource string, consumer func(prometheus.Collector)) error {
	var connOpts []nftables.ConnOption
	if netNS, _ := config.NetNS.Value(ctx); len(netNS) > 0 {
		n, e := netns.GetFromName(netNS)
		if e != nil {
			return errors.WithMessagef(e,
				"accessing netns '%s'", netNS)
		}
		connOpts = append(connOpts, nftables.WithNetNSFd(int(n)))
		defer n.Close() //nolint
	}
	nlConn, err := nftables.New(connOpts...)
	if err != nil {
		return err
	}
	log := logger.FromContext(ctx)
	if len(logSource) > 0 {
		log = log.Named(logSource)
	}
	var minRefreshInterval time.Duration
	if minRefreshInterval, err = config.NftablesCollectorMinFrequency.Value(ctx); err != nil {
		return err
	}
	ret := nfmetrics.NewCollector(ctx, nftconf.ListerFromConn(nlConn),
		nfmetrics.WithLogger(log),
		nfmetrics.WithMinFrequency(minRefreshInterval),
	)
	consumer(ret)
	return nil
}
