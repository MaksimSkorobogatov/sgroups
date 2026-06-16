package network

import (
	"context"

	"github.com/PRO-Robotech/sgroups/internal/sg-server/repository"
	"github.com/PRO-Robotech/sgroups/internal/sg-server/transport/sgroups/grpc/service"

	"github.com/H-BF/corlib/server"
	sgPkg "github.com/PRO-Robotech/sgroups-proto/pkg"
	sgv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
	grpcRt "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

// NewSgNetworkService creates service
func NewSgNetworkService(ctx context.Context, r repository.Repository, opts ...service.Option) server.APIService {
	ret := &networkService{
		appCtx: ctx,
		rep:    r,
	}
	ret.ApplyOptions(opts...)
	return ret
}

type networkService struct {
	sgv1.UnimplementedSGroupsNetworksAPIServer
	service.CommonOptions
	appCtx context.Context
	rep    repository.Repository
}

var (
	_ sgv1.SGroupsNetworksAPIServer = (*networkService)(nil)
	_ service.Service               = (*networkService)(nil)

	// SgNetworkSwaggerUtil ...
	SgNetworkSwaggerUtil sgPkg.SwaggerUtil[sgv1.SGroupsNetworksAPIServer]
)

// Description impl server.APIService
func (srv *networkService) Description() grpc.ServiceDesc {
	return sgv1.SGroupsNetworksAPI_ServiceDesc
}

// RegisterGRPC impl server/APIService
func (srv *networkService) RegisterGRPC(_ context.Context, s *grpc.Server) error {
	return service.RegisterGRPC(srv, s,
		service.WithAPIpathPrefixes(srv.PathPrefixes...),
		service.WithAdditionalServiceNames(srv.AdditionalServiceNames...),
	)
}

// RegisterProxyGW impl server.APIGatewayProxy
func (srv *networkService) RegisterProxyGW(ctx context.Context, mux *grpcRt.ServeMux, c *grpc.ClientConn) error {
	return sgv1.RegisterSGroupsNetworksAPIHandler(ctx, mux, c)
}
