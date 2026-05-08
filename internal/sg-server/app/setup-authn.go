package sgserver

import (
	"context"
	"crypto/tls"

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
	authn, e := AuthnType.Value(ctx)
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
		keyFilename  string
		certFilename string
		caFiles      []string
	)
	if keyFilename, err = TLSprivKeyFile.Value(ctx); err != nil {
		return ret, err
	}
	if certFilename, err = TLScertFile.Value(ctx); err != nil {
		return ret, err
	}
	caFiles, err = TLSclientCAfiles.Value(ctx)
	if err != nil && !errors.Is(err, config.ErrNotFound) {
		return ret, err
	}

	builder := tlsutil.ServerTLS(tlsutil.ServerCertsFromFiles(certFilename, keyFilename))

	var verifyClient config.TLSclientVerification
	if verifyClient, err = TLSclientVerifyStrategy.Value(ctx); err != nil {
		return nil, err
	}
	switch verifyClient {
	case config.TLSclientSkipVerify:

	case config.TLSclentCertsRequied:
		builder = builder.VerifyClient(tlsutil.RequireAnyCert())
	case config.TLSclientMustVerify:
		if len(caFiles) == 0 {
			return nil, errors.Errorf("should provide CA client cert(s) in config '%s' ",
				TLSclientCAfiles)
		}
		builder = builder.VerifyClient(tlsutil.RequireAndVerifyCerts(
			tlsutil.ClientCAsFromFiles(caFiles...),
		))
	default:
		return nil, errors.Errorf("unsupported '%s' client-verification", verifyClient)
	}

	return builder.BuildConfig(ctx)
}
