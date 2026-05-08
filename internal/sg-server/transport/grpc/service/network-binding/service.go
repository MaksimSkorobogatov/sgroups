package network_binding

import (
	"context"

	"github.com/PRO-Robotech/sgroups/internal/sg-server/repository"
	"github.com/PRO-Robotech/sgroups/internal/sg-server/transport/grpc/service"

	"github.com/H-BF/corlib/server"
	sgPkg "github.com/PRO-Robotech/sgroups-proto/pkg"
	sgv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
	grpcRt "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

// NewSgNetworkBindingService creates service
func NewSgNetworkBindingService(ctx context.Context, r repository.Repository, opts ...service.Option) server.APIService {
	ret := &nbService{
		appCtx: ctx,
		rep:    r,
	}
	ret.ApplyOptions(opts...)
	return ret
}

type nbService struct {
	sgv1.UnimplementedSGroupsNetworkBindingAPIServer
	service.CommonOptions
	appCtx context.Context
	rep    repository.Repository
}

var (
	_ sgv1.SGroupsNetworkBindingAPIServer = (*nbService)(nil)
	_ service.Service                     = (*nbService)(nil)

	//SgNetworkBindingSwaggerUtil ...
	SgNetworkBindingSwaggerUtil sgPkg.SwaggerUtil[sgv1.SGroupsNetworkBindingAPIServer]
)

// Description impl server.APIService
func (srv *nbService) Description() grpc.ServiceDesc {
	return sgv1.SGroupsNetworkBindingAPI_ServiceDesc
}

// RegisterGRPC impl server/APIService
func (srv *nbService) RegisterGRPC(_ context.Context, s *grpc.Server) error {
	return service.RegisterGRPC(srv, s,
		service.WithAPIpathPrefixes(srv.PathPrefixes...),
		service.WithAdditionalServiceNames(srv.AdditionalServiceNames...),
	)
}

// RegisterProxyGW impl server.APIGatewayProxy
func (srv *nbService) RegisterProxyGW(ctx context.Context, mux *grpcRt.ServeMux, c *grpc.ClientConn) error {
	return sgv1.RegisterSGroupsNetworkBindingAPIHandler(ctx, mux, c)
}
