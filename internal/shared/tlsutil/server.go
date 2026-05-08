package tlsutil

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"reflect"

	"github.com/PRO-Robotech/sgroups/internal/shared/misc"

	"github.com/H-BF/corlib/pkg/option"
	"github.com/pkg/errors"
)

// ServerTLS creates server TLS config builder
func ServerTLS[CertSrcT certSource](certs CertSrcT) serverTLSbuilder[CertSrcT] {
	return serverTLSbuilder[CertSrcT]{srvCerts: certs}
}

// RequireAnyCert sets [tls.RequireAnyClientCert] in the config
func RequireAnyCert() clientAuthority {
	return requireCerts{}
}

// RequireAndVerifyCerts sets [tls.RequireAndVerifyClientCert] and provided
// client CAs in the config
func RequireAndVerifyCerts(f clientCAs) clientAuthority {
	return requireCerts{
		clientCAs: option.MustNewOption(f),
	}
}

// VerifyClient enables client verification (off by default)
func (s serverTLSbuilder[T]) VerifyClient(a clientAuthority) serverTLSbuilder[T] {
	s.verifyClient.Set(a)
	return s
}

// BuildConfig builds TLS config
func (s serverTLSbuilder[T]) BuildConfig(ctx context.Context) (_ *tls.Config, err error) {
	var ret tls.Config
	switch src := any(s.srvCerts).(type) {
	case []tls.Certificate:
		ret.Certificates = src
	case func(context.Context) ([]tls.Certificate, error):
		ret.Certificates, err = src(ctx)
	default:
		return nil, errors.Errorf("unsupported server certificate source: '%s'", reflect.TypeOf(src))
	}
	if err != nil {
		return nil, errors.WithMessage(err, "get server certificates")
	}
	option.Match(s.verifyClient).Some(func(ca clientAuthority) {
		switch v := ca.(type) {
		case requireCerts:
			option.Match(v.clientCAs).Some(func(cc clientCAs) {
				ret.ClientAuth = tls.RequireAndVerifyClientCert
				if cas, e := cc(ctx); err != nil {
					err = e
				} else {
					ret.ClientCAs = x509.NewCertPool()
					for _, cert := range cas {
						ret.ClientCAs.AddCert(cert)
					}
				}
			}).None(func() {
				ret.ClientAuth = tls.RequireAnyClientCert
			})
		default:
			err = errors.Errorf("unsupported client authority: '%s'", reflect.TypeOf(v))
		}
	}).None(func() {
		ret.ClientAuth = tls.NoClientCert
	})
	return misc.Tern(err == nil, &ret, nil),
		errors.WithMessage(err, "build server tls config")
}

type (
	requireCerts struct {
		clientAuthority
		clientCAs option.ValueOf[clientCAs]
	}

	clientCAs = func(context.Context) ([]*x509.Certificate, error)

	clientAuthority interface {
		isClientAuthority()
	}

	serverTLSbuilder[T certSource] struct {
		srvCerts     T
		verifyClient option.ValueOf[clientAuthority]
	}

	certSource interface {
		[]tls.Certificate |
			func(context.Context) ([]tls.Certificate, error)
	}
)

// ServerCertsFromFiles creates server certs func that loads certs from provided file paths
func ServerCertsFromFiles(cert, key string) func(context.Context) ([]tls.Certificate, error) {
	return func(_ context.Context) ([]tls.Certificate, error) {
		return loadKeyPair(cert, key)
	}
}

// ClientCAsFromFiles creates clientCAs func that loads certs from provided file paths
func ClientCAsFromFiles(caPaths ...string) func(context.Context) ([]*x509.Certificate, error) {
	return func(_ context.Context) ([]*x509.Certificate, error) {
		return loadX509Certs(caPaths)
	}
}
