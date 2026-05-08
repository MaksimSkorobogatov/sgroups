package tlsutil

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"os"

	"github.com/PRO-Robotech/sgroups/internal/shared/misc"

	"github.com/H-BF/corlib/pkg/option"
	"github.com/pkg/errors"
)

// ClientTLS creates client TLS config builder
func ClientTLS() (r clientTLSbuilder) {
	return r
}

type (
	// ServerAuthority provides CAs and SAN for server verification on client side
	ServerAuthority interface {
		CertAuthorites(context.Context) ([]*x509.Certificate, error)
		SubjectAlternateName(context.Context) (option.ValueOf[string], error)
	}

	// ClientCerts provides client certificates
	ClientCerts interface {
		Certs(context.Context) ([]tls.Certificate, error)
	}

	clientTLSbuilder struct {
		server option.ValueOf[ServerAuthority]
		client option.ValueOf[ClientCerts]
	}
)

type (
	// ClientCertsFromFiles implements [ClientCerts] using provided keypair paths
	ClientCertsFromFiles struct {
		CertPath string
		KeyPath  string
	}

	// ServerAuthorityFromFiles implements [ServerAuthority] using CA file paths
	ServerAuthorityFromFiles struct {
		CAPaths []string
		SAN     option.ValueOf[string]
	}
)

var (
	_ ClientCerts     = (*ClientCertsFromFiles)(nil)
	_ ServerAuthority = (*ServerAuthorityFromFiles)(nil)
)

// VerifyServer enables server verification, using provided sa
func (c clientTLSbuilder) VerifyServer(sa ServerAuthority) clientTLSbuilder {
	c.server.Set(sa)
	return c
}

// VerifyClient sets provided client certificates
func (c clientTLSbuilder) VerifyClient(cc ClientCerts) clientTLSbuilder {
	c.client.Set(cc)
	return c
}

// BuildConfig builds TLS config
func (c clientTLSbuilder) BuildConfig(ctx context.Context) (ret *tls.Config, err error) {
	ret = &tls.Config{
		InsecureSkipVerify: true, //nolint:gosec
	}
	option.Match(c.server).Some(func(sa ServerAuthority) {
		ret.InsecureSkipVerify = false
		if san, e := sa.SubjectAlternateName(ctx); e != nil {
			err = e
		} else {
			option.Match(san).Some(func(s string) {
				ret.ServerName = s
			})
			if certs, e := sa.CertAuthorites(ctx); e != nil {
				err = e
			} else {
				ret.RootCAs = x509.NewCertPool()
				for _, cert := range certs {
					ret.RootCAs.AddCert(cert)
				}
			}
		}
	})
	if err == nil {
		option.Match(c.client).Some(func(cc ClientCerts) {
			ret.Certificates, err = cc.Certs(ctx)
			if err == nil && len(ret.Certificates) == 0 {
				err = errors.New("provided empty client cert list")
			}
		})
	}
	return misc.Tern(err == nil, ret, nil),
		errors.WithMessage(err, "build client tls config")
}

//-----------------------------------------------------------------------------------------------------------

// CertAuthorites -
func (s ServerAuthorityFromFiles) CertAuthorites(ctx context.Context) (_ []*x509.Certificate, err error) {
	return loadX509Certs(s.CAPaths)
}

// SubjectAlternateName -
func (s ServerAuthorityFromFiles) SubjectAlternateName(_ context.Context) (option.ValueOf[string], error) {
	return s.SAN, nil
}

// Certs -
func (c ClientCertsFromFiles) Certs(ctx context.Context) ([]tls.Certificate, error) {
	return loadKeyPair(c.CertPath, c.KeyPath)
}

func loadKeyPair(cert, key string) ([]tls.Certificate, error) {
	clientCert, err := tls.LoadX509KeyPair(cert, key)
	if err != nil {
		return nil, err
	}
	return misc.Sli(clientCert), nil
}

func loadX509Certs(caPaths []string) (certs []*x509.Certificate, err error) {
	if len(caPaths) == 0 {
		return nil, errors.New("no any server CA file is provided")
	}

	for _, caPath := range caPaths {
		var caBytes []byte
		if caBytes, err = os.ReadFile(caPath); err != nil {
			return nil, err
		}

		for {
			var block *pem.Block
			if block, caBytes = pem.Decode(caBytes); block == nil {
				break
			}

			var cert *x509.Certificate
			if cert, err = x509.ParseCertificate(block.Bytes); err != nil {
				return nil, err
			}
			certs = append(certs, cert)
		}
	}

	return misc.Tern(err == nil, certs, nil),
		errors.WithMessage(err, "build client tls config")
}
