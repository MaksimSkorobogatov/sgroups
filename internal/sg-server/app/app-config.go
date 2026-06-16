//nolint:revive
package app

import (
	"time"

	"github.com/H-BF/corlib/logger"
	config "github.com/H-BF/corlib/pkg/plain-config"
)

/*//Sample of config
logger:
  level: INFO

metrics:
  enable: true

healthcheck:
  enable: true

server:
  endpoint: tcp://127.0.0.1:9006
  graceful-shutdown: 30s
  api-path-prefix: "static/path/prefix" #is empty by default
  additional-api-paths: ["static/path"] #is empty by default
  use-proto-validator: true #is false by default; use protovalidate interceptors for grpc requests validation

storage:
  type: postgres ; only the postgres is supported for now
  postgres:
    url: postgres://un:psw@host/db

authn:
  type: <none|tls> #authentication type; `none` is by default
  tls:
    key-file: "filename1.pem"
    cert-file: "filename2.pem"
    client:
      verify: <skip|certs-required|verify> # 'skip' is by default
      ca-files: ["file1.pem", "file2.pem", "file3.pem", ...]

extapi:
  def-dial-duration: 10s #optional; default=10s
  agents:
    dial-duration: 3s #override def-dial-duration
    authn:
      type: "none|tls" #optional; default="none"
      tls:
        key-file: "key-file.pem"
        cert-file: "cert-file.pem"
        server:
          verify: <true|false> # false is by default
          name: "server-name" # is not present by default
          ca-files: ["file1.pem", "file2.pem", ...] # is not present by default

*/

// logger section
const (
	// LoggerLevel log level
	LoggerLevel config.ValueT[logger.LoggerLevelConf] = "logger/level"
)

// server section
const (
	// ServerEndpoint server endpoint
	ServerEndpoint config.ValueT[string] = "server/endpoint"

	// ServerGracefulShutdown graceful shutdown period
	ServerGracefulShutdown config.ValueT[time.Duration] = "server/graceful-shutdown"

	// ServerAPIpathPrefix is a path api prefix
	ServerAPIpathPrefix config.ValueT[string] = "server/api-path-prefix"

	// ServerAdditionalAPIpaths is additional api paths
	ServerAdditionalAPIpaths config.ValueT[[]string] = "server/additional-api-paths"

	// ServerUseBufProtoValidator enables protovalidate interceptors for grpc requests validation
	ServerUseBufProtoValidator config.ValueT[bool] = "server/use-proto-validator"
)

// metrics section
const (
	// MetricsEnable enable api metrics
	MetricsEnable config.ValueT[bool] = "metrics/enable"
	// HealthcheckEnable enables|disables health check handler
	HealthcheckEnable config.ValueT[bool] = "healthcheck/enable"
)

// storage section
const (
	// StorageType selects storage DB backend
	StorageType config.ValueT[string] = "storage/type"

	// PostgresURL URL to connect Postgres DB
	PostgresURL config.ValueT[string] = "storage/postgres/url"
)

// auth section
const (
	// AuthnType selects authn type <none|tls> where `none` is by default
	AuthnType config.AuthnTypeSelector = "authn/type"

	// TLSprivKeyFile server private key PEM encoded file
	TLSprivKeyFile config.TLSprivKeyFile = "authn/tls/key-file"

	// TLScertFile server cert PEM encoded file
	TLScertFile config.TLScertFile = "authn/tls/cert-file"

	// TLSclientCAfiles client cert authority PEM files
	TLSclientCAfiles config.TLScaFiles = "authn/tls/client/ca-files"

	// TLSclientVerifyStrategy verify client and certs a.k.a MTLS; 'skip' is by default
	TLSclientVerifyStrategy config.TLSclientVerifyStrategy = "authn/tls/client/verify"
)

// extapi section
const (
	// ExtapiDefDialDuration -
	ExtapiDefDialDuration config.ValueT[time.Duration] = "extapi/def-dial-duration"
	// ExtapiAgentDialDuration -
	ExtapiAgentDialDuration config.ValueT[time.Duration] = "extapi/agents/dial-duration"
	// ExtapiAgentAuthnType -
	ExtapiAgentAuthnType config.AuthnTypeSelector = "extapi/agents/authn/type"
	// ExtapiAgentTLSCertFile -
	ExtapiAgentTLSCertFile config.TLScertFile = "extapi/agents/authn/tls/cert-file"
	// ExtapiAgentTLSKeyFile -
	ExtapiAgentTLSKeyFile config.TLSprivKeyFile = "extapi/agents/authn/tls/key-file"
	// ExtapiAgentTLSServerVerify -
	ExtapiAgentTLSServerVerify config.ValueT[bool] = "extapi/agents/authn/tls/server/verify"
	// ExtapiAgentTLSServerName -
	ExtapiAgentTLSServerName config.TLSverifysServerName = "extapi/agents/authn/tls/server/name"
	// ExtapiAgentTLSServerCAfiles -
	ExtapiAgentTLSServerCAfiles config.TLScaFiles = "extapi/agents/authn/tls/server/ca-files"
)
