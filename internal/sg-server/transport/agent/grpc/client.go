package grpc

import (
	"context"
	"io"

	"github.com/H-BF/corlib/client/grpc"
	agentv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/agent/v1"
)

type (
	// Client -
	Client struct {
		agentv1.AgentAPIClient
		io.Closer
	}
	// ClientProvider -
	ClientProvider interface {
		New(ctx context.Context, addr string) (Client, error)
	}
	clientConnProvider interface {
		New(ctx context.Context, addr string) (grpc.ClientConn, error)
	}
	clientProviderImpl struct {
		clientConnProvider
	}
)

// NewClientProvider -
func NewClientProvider(f clientConnProvider) ClientProvider {
	return clientProviderImpl{clientConnProvider: f}
}

// New -
func (cp clientProviderImpl) New(ctx context.Context, addr string) (ret Client, err error) {
	c, e := cp.clientConnProvider.New(ctx, addr)
	if e != nil {
		return ret, e
	}
	ret.AgentAPIClient = agentv1.NewAgentAPIClient(c)
	ret.Closer = c
	return ret, nil
}
