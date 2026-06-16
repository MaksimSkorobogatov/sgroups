package app

import (
	"context"
	"crypto/tls"

	conf "github.com/PRO-Robotech/sgroups/internal/sg-agent/config"
	"github.com/PRO-Robotech/sgroups/internal/shared/misc"
	"github.com/PRO-Robotech/sgroups/internal/shared/tlsutil"

	config "github.com/H-BF/corlib/pkg/plain-config"
	"github.com/pkg/errors"
)

type authnType interface {
	isAuthnType()
}

type authnTLS struct {
	authnType
	conf *tls.Config
}

func whenAuthn(ctx context.Context, consume func(authnType) error) error {
	authn, e := conf.ApiAuthnType.Value(ctx)
	if e != nil {
		return e
	}
	switch authn {
	case config.AuthnTypeTLS:
		cnf, err := setupTLS(ctx)
		if err != nil {
			return errors.WithMessage(err, "on setup TLS")
		}
		return consume(authnTLS{conf: cnf})
	case config.AuthnTypeNONE:
	default:
		return errors.Errorf("unsupported authn type '%s'", authn)
	}
	return nil
}

func setupTLS(ctx context.Context) (ret *tls.Config, err error) {
	var (
		pkeyFilename       string
		certFilename       string
		verifyClientMethod config.TLSclientVerification
	)
	if pkeyFilename, err = conf.ApiTLSKeyFile.Value(ctx); err != nil {
		return nil, err
	}
	if certFilename, err = conf.ApiTLSCertFile.Value(ctx); err != nil {
		return nil, err
	}
	if verifyClientMethod, err = conf.ApiTLSClientVerifyStrategy.Value(ctx); err != nil {
		return nil, err
	}
	b := tlsutil.ServerTLS(
		tlsutil.ServerCertsFromFiles(certFilename, pkeyFilename),
	)
	switch verifyClientMethod {
	case config.TLSclientSkipVerify:
	case config.TLSclentCertsRequied:
		b = b.VerifyClient(tlsutil.RequireAnyCert())
	case config.TLSclientMustVerify:
		var clientCAFiles []string
		if clientCAFiles, err = conf.ApiTLSClientCAfiles.Value(ctx); err != nil {
			return nil, err
		}
		b = b.VerifyClient(tlsutil.RequireAndVerifyCerts(
			tlsutil.ClientCAsFromFiles(clientCAFiles...)),
		)
	default:
		return nil, errors.Errorf(
			"unsupported client verify method '%s'", verifyClientMethod,
		)
	}
	ret, err = b.BuildConfig(ctx)
	return misc.Tern(err == nil, ret, nil), err
}
