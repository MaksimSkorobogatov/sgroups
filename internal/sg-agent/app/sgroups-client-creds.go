package app

import (
	"context"
	"crypto/tls"

	conf "github.com/PRO-Robotech/sgroups/internal/sg-agent/config"
	"github.com/PRO-Robotech/sgroups/internal/shared/misc"
	"github.com/PRO-Robotech/sgroups/internal/shared/tlsutil"

	"github.com/H-BF/corlib/pkg/option"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

func makeSgroupsClientCreds(ctx context.Context) (creds credentials.TransportCredentials, err error) { //nolint:gocyclo
	defer func() {
		err = errors.WithMessage(err, "make SGroups client creds")
	}()
	var authnType conf.AuthnType
	if authnType, err = conf.SGroupsAuthnType.Value(ctx); err != nil {
		return nil, err
	}
	switch authnType {
	case conf.AuthnTypeNONE:
		creds = insecure.NewCredentials()
	case conf.AuthnTypeTLS:
		var (
			verifyServer bool
			caPaths      []string
			serverName   option.ValueOf[string]
			tls          *tls.Config
		)
		builder := tlsutil.ClientTLS()

		if verifyServer, err = conf.SGroupsTLSserverVerify.Value(ctx); err != nil {
			return nil, err
		}
		if verifyServer {
			if caPaths, err = conf.SGroupsTLSserverCAs.Value(ctx); err != nil {
				return nil, err
			}
			if serverName, err = misc.GetOptionalConfig(ctx, conf.SGroupsTLSserverName); err != nil {
				return nil, err
			}
			builder = builder.VerifyServer(tlsutil.ServerAuthorityFromFiles{
				CAPaths: caPaths,
				SAN:     serverName,
			})
		}

		cert, e1 := misc.GetOptionalConfig(ctx, conf.SGroupsTLScertFile)
		pkey, e2 := misc.GetOptionalConfig(ctx, conf.SGroupsTLSprivKeyFile)
		if err = multierr.Combine(e1, e2); err == nil {
			option.Match(cert).Some(func(certFile string) {
				option.Match(pkey).Some(func(pkeyFile string) {
					builder = builder.VerifyClient(tlsutil.ClientCertsFromFiles{
						CertPath: certFile,
						KeyPath:  pkeyFile,
					})
				})
			})
			tls, err = builder.BuildConfig(ctx)
		}
		creds = misc.Tern(err == nil, credentials.NewTLS(tls), nil)

	default:
		err = errors.Errorf("unsupported authn type '%s'", authnType)
	}
	return creds, nil
}
