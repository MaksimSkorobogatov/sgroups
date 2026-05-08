//nolint:revive
package config

import (
	"context"
	"time"

	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"

	"github.com/H-BF/corlib/logger"
	conf "github.com/H-BF/corlib/pkg/plain-config"
)

/*// config-sample.yaml
hostinfo:c "git
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
    def-daial-duration: 10s
    sgroups:
      dial-duration: 3s #override default-connect-tmo
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
  useragent: "string"
  nft-collector:
    min-frequency: 1s
  endpoint: 127.0.0.1:5000
  metrics:
    enable: true
  healthcheck:
    enable: true
  profile:
    enable: true
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
	ServicesDefDialDuration conf.ValueT[time.Duration] = "extapi/svc/def-daial-duration"
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
	// TelemetryEndpoint server endpoint
	TelemetryEndpoint conf.ValueT[string] = "telemetry/endpoint"
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
