//nolint:revive
package config

import (
	"context"
	"strconv"
	"time"

	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"

	"github.com/H-BF/corlib/logger"
	pkgNet "github.com/H-BF/corlib/pkg/net"
	conf "github.com/H-BF/corlib/pkg/plain-config"
	"github.com/pkg/errors"
)

/*// config-sample.yaml
hostinfo:
  name: host-name-string #mandatory; no-default
  namespace: host-namespace-string #mandatory; no-default
exit-on-success: true|false - do exit when we succeeded to apply netfilter config; def-val=false
continue-on-failure: true|false - when 'true' it means if something fails it internally restarts all workloads after some tomeout; when 'false' if something fails the app exits with code 1; default=true
continue-after-timeout: 10s - if 'continue-on-failure'=true then we use this value to do timeout befor restart
netns: NetworkNS #is optional; def-val = ""
graceful-shutdown: 10s
def-policy-accept: true|false #if 'true' it sets 'accept' verdict at ingress & egress chains in its default rules; default=false
dry-run: true|false #if 'true' it doesn't apply rules to the system but just logs them; default=false
base-rules: | #optional; json-string ;no-default
   [ <-- example
     {
       "nets": ["10.10.1.0/24", "10.10.2.0/24"], <-mandatory
       "ing": { <-- optional - ingress
          "tcp": { <-- optional
            "ports": "1-100, 80" <-- optional
          },
          "udp": { <-- optional
            "ports": "1-100, 80" <-- optional - egress
          },
          "icmp": { <-- optional
            "types": [1,2,3,....] <-- optional
          },
          "icmp6": { <-- optional
            "types": [1,2,3,....] <-- optional
          }
       },
       "egr": { <-- optional - egress
          #similar to "ing"
       }
     },
     {
     ....
     }
   ]

fqdn-rules:
  strategy: dns #default = dns
logger:
  level: INFO
netlink:
  watcher: #netlink watcher
    linger: 10s
dns:
  nameservers: ["8.8.8.8", "1.1.1.1", "...", ] #default ["8.8.8.8"]
  proto: tcp|udp #default udp
  port: 53 #default 53
  dial-duration: 3s #default 3s
  read-duration: 5s #default 5s
  write-duration: 5s #default 5s
  retries: 5 #default 1
  retry-timeout: 3s #default 1s

extapi:
  svc:
    def-dial-duration: 10s
    sgroups:
      dial-duration: 3s #override def-dial-duration
      address: tcp://127.0.0.1:9006
      sync-status:
        interval: 20s #mandatory
        push: true
      use-json-codec: <true|false> # = false by default
      api-path-prefix: "a/b/c" # = is not set by default
      authn:
        type: <none|tls> # 'none' is by default
        tls:
          key-file: "key-file.pem"
          cert-file: "cert-file.pem"
          server:
            verify: <true|false> # false is by default
            name: "server-name" # is not present by default
            ca-files: ["file1.pem", "file2.pem", ...] # is not present by default

telemetry:
  address: 127.0.0.1:5000 #mandatory; default= tcp://127.0.0.1:5000
  useragent: "string"
  nft-collector:
    min-frequency: 1s
  metrics:
    enable: true
  healthcheck:
    enable: true
  profile:
    enable: true

ss: #socket scan section
  strategy: # oneof<cached|non-cached> #default=cached
  cache: # optional; used when strategy=cached
    ttl: 2s #default=2s

nft: #nftables section
  scan: #nftables scan section
    sync-interval: 1s #default=1s; used in 'watch' usecase to do periodic resync with nftables state
    strategy: # oneof<cached|non-cached> #default=cached
    cache:
      ttl: 2s #default=2s

api:
  address: tcp://127.0.0.1:5000 #mandatory; default= tcp://127.0.0.1:5000
  authn:
    type: oneof<none|tls> # 'none' is by default
    tls:
      key-file: "key-file.pem"
      cert-file: "cert-file.pem"
      client:
        verify: oneof<skip|certs-required|verify> #optional; default='skip'
        ca-files: ["file1.pem", "file2.pem", "file3.pem", ...] #CA files when client/verify points to 'verify'
*/

// hostinfo section
const (
	// HostName host name
	HostName conf.ValueT[string] = "hostinfo/name"
	// HostNamespace host namespace
	HostNamespace conf.ValueT[string] = "hostinfo/namespace"
)

const (
	// ExitOnSuccess do exit when we succeeded to apply netfilter config; def-val=false
	ExitOnSuccess conf.ValueT[bool] = "exit-on-success"
	// ContinueOnFailure (default = true)
	// when 'true' it means if something fails it internally restarts all workloads after some tomeout
	// when 'false' if something fails the app exits with code 1
	ContinueOnFailure conf.ValueT[bool] = "continue-on-failure"
	// ContinueAfterTimeout (default = '10s' )
	// if 'continue-on-failure'=true then we use this value to do timeout befor restart
	ContinueAfterTimeout conf.ValueT[time.Duration] = "continue-after-timeout"
	// AppLoggerLevel log level [optional]
	AppLoggerLevel conf.ValueT[logger.LoggerLevelConf] = "logger/level"
	// AppGracefulShutdown [optional]
	AppGracefulShutdown conf.ValueT[time.Duration] = "graceful-shutdown"
	// NetNS network namespace
	NetNS conf.ValueT[string] = "netns"
	// DefPolicyAccept sets accept verdict as default rule at ingress&egress; default=false
	DefPolicyAccept conf.ValueT[bool] = "def-policy-accept"
	// DryRun if 'true' it doesn't apply rules to the system but just logs them; default=false
	DryRun conf.ValueT[bool] = "dry-run"
	// NetlinkWatcherLinger netlingk watched linger duration, min(1s)
	NetlinkWatcherLinger conf.ValueT[time.Duration] = "netlink/watcher/linger"
	// BaseRulesConfig -
	BaseRulesConfig conf.ValueT[string] = "base-rules"
	// FqdnStrategy use strategy to build SG-FQDN rules (DNS|NDPI|Combine); DNS is default
	FqdnStrategy FqdnRulesStrategySelector = "fqdn-rules/strategy"
)

// dns section
const (
	// DnsNameservers IP list of trusted nameservers; default = ["8.8.8.8"]
	DnsNameservers conf.ValueT[[]conf.IP] = "dns/nameservers"
	// DnsProto tcp or udp protp we shoud use; default = udp
	DnsProto conf.ValueT[string] = "dns/proto"
	// DnsPort use port to ask nameserver(s); default = 53
	DnsPort conf.ValueT[uint16] = "dns/port"
	// DnsRetries on failure retries count; default=3
	DnsRetries conf.ValueT[uint8] = "dns/retries"
	// DnsRetriesTmo timeout before retry; default=1s
	DnsRetriesTmo conf.ValueT[time.Duration] = "dns/retry-timeout"
	// DnsDialDuration dial max duration; default = 3s
	DnsDialDuration conf.ValueT[time.Duration] = "dns/dial-duration"
	// DnsWriteDuration packet write max duration; default = 5s
	DnsWriteDuration conf.ValueT[time.Duration] = "dns/write-duration"
	// DnsReadDuration response wait+read max duration; default = 5s
	DnsReadDuration conf.ValueT[time.Duration] = "dns/read-duration"
)

// extapi/svc section
const (
	// ServicesDefDialDuration default dial duraton to conect a service [optional]
	ServicesDefDialDuration conf.ValueT[time.Duration] = "extapi/svc/def-dial-duration"
	//SGroupsAddress service address [mandatory]
	SGroupsAddress conf.ValueT[string] = "extapi/svc/sgroups/address"
	//SGroupsDialDuration sgroups service dial duration [optional]
	SGroupsDialDuration conf.ValueT[time.Duration] = "extapi/svc/sgroups/dial-duration"
	//SGroupsSyncStatusInterval interval(duration) backend 'sync-status' check [mandatory]
	SGroupsSyncStatusInterval conf.ValueT[time.Duration] = "extapi/svc/sgroups/sync-status/interval"
	//SGroupsSyncStatusPush use push model of 'sync-status'
	SGroupsSyncStatusPush conf.ValueT[bool] = "extapi/svc/sgroups/sync-status/push"
	// SGroupsUseJsonCodec use GRPC+JSON codec instead of GRPC+PROTO
	SGroupsUseJsonCodec conf.ValueT[bool] = "extapi/svc/sgroups/use-json-codec"
	// SGroupsAPIpathPrefix add path prefix when call SGROUPS API - is not set by default
	SGroupsAPIpathPrefix conf.ValueT[string] = "extapi/svc/sgroups/api-path-prefix"
)

// Authn section
const (
	// SGroupsAuthnType -
	SGroupsAuthnType conf.AuthnTypeSelector = "extapi/svc/sgroups/authn/type"
	// SGroupsTLScertFile client cert file
	SGroupsTLScertFile conf.TLScertFile = "extapi/svc/sgroups/authn/tls/cert-file"
	// SGroupsTLSprivKeyFile client private key
	SGroupsTLSprivKeyFile conf.TLScertFile = "extapi/svc/sgroups/authn/tls/key-file"
	// SGroupsTLSserverVerify if true we need verify server host or IPs
	SGroupsTLSserverVerify conf.ValueT[bool] = "extapi/svc/sgroups/authn/tls/server/verify"
	// SGroupsTLSserverName server hostname we need to verify - not set by default
	SGroupsTLSserverName conf.TLSverifysServerName = "extapi/svc/sgroups/authn/tls/server/name"
	// SGroupsTLSserverCAs server CA files
	SGroupsTLSserverCAs conf.TLScaFiles = "extapi/svc/sgroups/authn/tls/server/ca-files"
)

// telemetry section
const (
	// TelemetryAddr server addr
	TelemetryAddr conf.ValueT[string] = "telemetry/address"
	// MetricsEnable enable api metrics
	MetricsEnable conf.ValueT[bool] = "telemetry/metrics/enable"
	// HealthcheckEnable enables|disables health check handler
	HealthcheckEnable conf.ValueT[bool] = "telemetry/healthcheck/enable"
	// UserAgent -
	UserAgent conf.ValueT[string] = "telemetry/useragent"
	// ProfileEnable available at /debug/pprof/index
	ProfileEnable conf.ValueT[bool] = "telemetry/profile/enable"
	// NftablesCollectorMinFrequency states how often to update cache with nft metrics
	NftablesCollectorMinFrequency conf.ValueT[time.Duration] = "telemetry/nft-collector/min-frequency"
)

// ss section
const (
	// SocketScanStrategy oneof<cached|non-cached> #default=cached
	SocketScanStrategy ScanStrategySelector = "ss/strategy"
	// SocketScanCacheTTL cache TTL for socket scan results; default=2s
	SocketScanCacheTTL conf.ValueT[time.Duration] = "ss/cache/ttl"
)

// nft section
const (
	// NftScanSyncInterval interval(duration) to sync nftables state
	NftScanSyncInterval conf.ValueT[time.Duration] = "nft/scan/sync-interval"
	// NftScanStrategy oneof<cached|non-cached> #default=cached
	NftScanStrategy ScanStrategySelector = "nft/scan/strategy"
	// NftScanCacheTTL cache TTL for nft scan results; default=2s
	NftScanCacheTTL conf.ValueT[time.Duration] = "nft/scan/cache/ttl"
)

// agent api section
const (
	// ApiAddress -
	ApiAddress conf.ValueT[string] = "api/address"
	// ApiAuthnType selects authn type <none|tls>
	ApiAuthnType conf.AuthnTypeSelector = "api/authn/type"
	// ApiTLSKeyFile server private key PEM encoded file
	ApiTLSKeyFile conf.TLSprivKeyFile = "api/authn/tls/key-file"
	// ApiTLSCertFile server cert PEM encoded file
	ApiTLSCertFile conf.TLScertFile = "api/authn/tls/cert-file"
	// ApiTLSClientVerifyStrategy verify client and certs
	ApiTLSClientVerifyStrategy conf.TLSclientVerifyStrategy = "api/authn/tls/client/verify"
	// ApiTLSClientCAfiles client cert authority PEM files
	ApiTLSClientCAfiles conf.TLScaFiles = "api/authn/tls/client/ca-files"
)

// ErrNotFound - alias for config.ErrNotFound
var ErrNotFound = conf.ErrNotFound

type (
	// IP - alias for conf.IP
	IP = conf.IP
	// AuthnType - alias for conf.AuthnType
	AuthnType = conf.AuthnType
)

const (
	// AuthnTypeNONE - alias for conf.AuthnTypeNONE
	AuthnTypeNONE = conf.AuthnTypeNONE
	// AuthnTypeTLS - alias for conf.AuthnTypeTLS
	AuthnTypeTLS = conf.AuthnTypeTLS
)

// GetHostID - get host ID (name+namespace) from config
func GetHostID(ctx context.Context) (ret domain.ResourceIdentifier, err error) {
	name, err := HostName.Value(ctx)
	if err != nil {
		return ret, err
	}
	namespace, err := HostNamespace.Value(ctx)
	if err != nil {
		return ret, err
	}
	return domain.ResourceIdentifier{
		Name:      domain.ResourceName(name),
		Namespace: domain.ResourceNamespace(namespace),
	}, nil
}

// GetEndpoints get host endpoints from config
func GetEndpoints(ctx context.Context) (ret domain.HostEndpoints, err error) {
	var (
		metricAddr    string
		apiAddr       string
		metricPortNum domain.PortNumber
		apiPortNum    domain.PortNumber
	)
	if metricAddr, err = TelemetryAddr.Value(ctx); err != nil {
		return ret, err
	}
	if metricPortNum, err = portFromAddr(metricAddr); err != nil {
		return ret, err
	}
	if apiAddr, err = ApiAddress.Value(ctx); err != nil {
		return ret, err
	}
	if apiPortNum, err = portFromAddr(apiAddr); err != nil {
		return ret, err
	}
	ret.Ports = []domain.NamedPort{
		{
			Name: domain.AgentTelemetryEndpointName,
			Port: metricPortNum,
		},
		{
			Name: domain.AgentApiEndpointName,
			Port: apiPortNum,
		},
	}

	return ret, nil
}

// AddrsEqual compares two endpoints addresses by their ports and network types
func AddrsEqual(a1, a2 string) (bool, error) {
	port1, err := portFromAddr(a1)
	if err != nil {
		return false, err
	}
	port2, err := portFromAddr(a2)
	if err != nil {
		return false, err
	}
	if port1 != port2 {
		return false, nil
	}
	ep1, _ := pkgNet.ParseEndpoint(a1)
	ep2, _ := pkgNet.ParseEndpoint(a2)
	return ep1.Network() == ep2.Network(), nil
}

func portFromAddr(addr string) (ret domain.PortNumber, err error) {
	var (
		ep      *pkgNet.Endpoint
		port    string
		portNum int
	)
	ep, err = pkgNet.ParseEndpoint(addr)
	if err != nil {
		return ret, errors.WithMessagef(err, "parse telemetry endpoint (%s): %v", addr, err)
	}
	_, port, err = ep.HostPort()
	if err != nil {
		return ret, errors.WithMessagef(err, "get telemetry endpoint port (%s): %v", addr, err)
	}
	portNum, err = strconv.Atoi(port)
	if err != nil {
		return ret, errors.WithMessagef(err, "convert telemetry endpoint port (%s): %v", port, err)
	}

	return domain.PortNumber(portNum), nil //nolint:gosec
}
