package app

import (
	"context"
	"time"

	conf "github.com/PRO-Robotech/sgroups/internal/sg-agent/config"
	client "github.com/PRO-Robotech/sgroups/internal/sg-agent/transport/sgroups/grpc"

	grpc_client "github.com/H-BF/corlib/client/grpc"
	config "github.com/H-BF/corlib/pkg/plain-config"
	"github.com/pkg/errors"
	"google.golang.org/grpc/credentials"
)

type (
	// SGClientProvider -
	SGClientProvider interface {
		NewSGClient(ctx context.Context) (*client.Clients, error)
	}
	sgClientProviderImpl struct{}
)

// SGClientProviderInstance is the instance of SGClientProvider
var SGClientProviderInstance sgClientProviderImpl

var _ SGClientProvider = (*sgClientProviderImpl)(nil)

// NewSGClient makes 'sgroups' API clients
func (sgClientProviderImpl) NewSGClient(ctx context.Context) (ret *client.Clients, err error) {
	const api = "NewSGClient"

	defer func() {
		err = errors.WithMessage(err, api)
	}()

	addr, err := conf.SGroupsAddress.Value(ctx)
	if err != nil {
		return ret, err
	}
	var dialDuration time.Duration
	dialDuration, err = conf.SGroupsDialDuration.Value(ctx)
	if errors.Is(err, config.ErrNotFound) {
		dialDuration, err = conf.ServicesDefDialDuration.Value(ctx)
		if errors.Is(err, config.ErrNotFound) {
			err = nil
		}
	}
	if err != nil {
		return ret, err
	}
	var creds credentials.TransportCredentials
	if creds, err = makeSgroupsClientCreds(ctx); err != nil {
		return ret, err
	}
	bld := grpc_client.ClientFromAddress(addr).
		WithDialDuration(dialDuration).
		WithCreds(creds).
		WithUserAgent(conf.UserAgent.MustValue(ctx))
	if v, e := conf.SGroupsUseJsonCodec.Value(ctx); e == nil && v {
		bld = bld.WithDefaultCodecByName(client.JsonCodecName)
	} else if e != nil && !errors.Is(e, config.ErrNotFound) {
		return ret, e
	}
	if o, e := conf.SGroupsAPIpathPrefix.Value(ctx); e == nil {
		bld = bld.WithPathPrefix(o)
	} else if !errors.Is(e, config.ErrNotFound) {
		return ret, e
	}
	var c grpc_client.ClientConn
	if c, err = bld.New(ctx); err != nil {
		return ret, err
	}
	return client.NewClients(c)
}
