package service

import (
	"context"
	"time"

	"github.com/H-BF/corlib/server"
	sgPkg "github.com/PRO-Robotech/sgroups-proto/pkg"
	agentv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/agent/v1"
	grpcRt "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

// NewAgentService -
func NewAgentService(ctx context.Context, ssCacheTTL, nftCacheTTL, nftSyncInterval time.Duration) *agentAPI {
	return &agentAPI{
		appCtx:          ctx,
		ssCacheTTL:      ssCacheTTL,
		nftCacheTTL:     nftCacheTTL,
		nftSyncInterval: nftSyncInterval,
	}
}

var (
	_ server.APIService      = (*agentAPI)(nil)
	_ server.APIGatewayProxy = (*agentAPI)(nil)

	// AgentSwaggerUtil -
	AgentSwaggerUtil sgPkg.SwaggerUtil[agentv1.AgentAPIServer]
)

type agentAPI struct {
	agentv1.UnimplementedAgentAPIServer
	appCtx          context.Context
	ssCacheTTL      time.Duration
	nftCacheTTL     time.Duration
	nftSyncInterval time.Duration
}

// Description impl server.APIService
func (srv *agentAPI) Description() grpc.ServiceDesc {
	return agentv1.AgentAPI_ServiceDesc
}

// RegisterGRPC impl server.APIService
func (srv *agentAPI) RegisterGRPC(_ context.Context, s *grpc.Server) error {
	desc := srv.Description()
	s.RegisterService(&desc, srv)
	return nil
}

// RegisterProxyGW impl server.APIGatewayProxy
func (srv *agentAPI) RegisterProxyGW(ctx context.Context, mux *grpcRt.ServeMux, c *grpc.ClientConn) error {
	return agentv1.RegisterAgentAPIHandler(ctx, mux, c)
}
