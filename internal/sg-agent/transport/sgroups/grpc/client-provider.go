package grpc

import (
	"context"

	inner "github.com/H-BF/corlib/client/grpc"
)

type (
	// ClientConn is a type alias
	ClientConn = inner.ClientConn

	// ConnProvider grpc client conn provider
	ConnProvider interface {
		New(ctx context.Context) (ClientConn, error)
	}
)
