package app

import (
	"context"
	"crypto/tls"
	"time"

	"github.com/PRO-Robotech/sgroups/internal/shared/misc"
	"github.com/PRO-Robotech/sgroups/internal/shared/tlsutil"

	"github.com/H-BF/corlib/client/grpc"
	"github.com/H-BF/corlib/pkg/option"
	config "github.com/H-BF/corlib/pkg/plain-config"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

func newAgentClientConnProvider(ctx context.Context) (ret *agentClientConnProvider, err error) {
	var dialDurOpt option.ValueOf[time.Duration]

	if dialDurOpt, err = misc.GetOptionalConfig(ctx, ExtapiAgentDialDuration); err != nil {
		return nil, err
	}
	if dialDurOpt.IsNone() {
		dialDurOpt, err = misc.GetOptionalConfig(ctx, ExtapiDefDialDuration)
		if err != nil {
			return nil, err
		}
	}

	ret = new(agentClientConnProvider)
	ret.dialDur, _ = dialDurOpt.Maybe()

	var authnType config.AuthnType
	if authnType, err = ExtapiAgentAuthnType.Value(ctx); err != nil {
		return nil, err
	}

	switch authnType {
	case config.AuthnTypeNONE:
	case config.AuthnTypeTLS:
		var (
			serverVerify bool
			caFiles      []string
			serverName   option.ValueOf[string]
		)

		builder := tlsutil.ClientTLS()

		if serverVerify, err = ExtapiAgentTLSServerVerify.Value(ctx); err != nil {
			return nil, err
		}
		if serverVerify {
			if caFiles, err = ExtapiAgentTLSServerCAfiles.Value(ctx); err != nil {
				return nil, err
			}
			if serverName, err = misc.GetOptionalConfig(ctx, ExtapiAgentTLSServerName); err != nil {
				return nil, err
			}
			builder = builder.VerifyServer(tlsutil.ServerAuthorityFromFiles{
				CAPaths: caFiles,
				SAN:     serverName,
			})
		}

		cert, e1 := misc.GetOptionalConfig(ctx, ExtapiAgentTLSCertFile)
		pkey, e2 := misc.GetOptionalConfig(ctx, ExtapiAgentTLSKeyFile)
		if err = multierr.Combine(e1, e2); err == nil {
			option.Match(cert).Some(func(certFile string) {
				option.Match(pkey).Some(func(pkeyFile string) {
					builder = builder.VerifyClient(tlsutil.ClientCertsFromFiles{
						CertPath: certFile,
						KeyPath:  pkeyFile,
					})
				})
			})
			ret.tlsCfg, err = builder.BuildConfig(ctx)
		}
		return misc.Tern(err == nil, ret, nil), err
	default:
		return nil, errors.Errorf(
			"unexpected %s config value %v, must be in %v",
			ExtapiAgentAuthnType, authnType, authnType.Variants(),
		)
	}

	return ret, nil
}

type agentClientConnProvider struct {
	dialDur time.Duration
	tlsCfg  *tls.Config
}

// New -
func (a agentClientConnProvider) New(ctx context.Context, addr string) (grpc.ClientConn, error) {
	creds := insecure.NewCredentials()
	if a.tlsCfg != nil {
		creds = credentials.NewTLS(a.tlsCfg)
	}
	return grpc.ClientFromAddress(addr).
		WithDialDuration(a.dialDur).
		WithCreds(creds).New(ctx)
}
