package host

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

// NewSgHostService creates service
func NewSgHostService(ctx context.Context, r repository.Repository, opts ...service.Option) server.APIService {
	ret := &hostService{
		appCtx: ctx,
		rep:    r,
	}
	ret.ApplyOptions(opts...)
	return ret
}

type hostService struct {
	sgv1.UnimplementedSGroupsHostsAPIServer
	service.CommonOptions
	appCtx context.Context
	rep    repository.Repository
}

var (
	_ sgv1.SGroupsHostsAPIServer = (*hostService)(nil)
	_ service.Service            = (*hostService)(nil)

	//SgHostSwaggerUtil ...
	SgHostSwaggerUtil sgPkg.SwaggerUtil[sgv1.SGroupsHostsAPIServer]
)

// Description impl server.APIService
func (srv *hostService) Description() grpc.ServiceDesc {
	return sgv1.SGroupsHostsAPI_ServiceDesc
}

// RegisterGRPC impl server/APIService
func (srv *hostService) RegisterGRPC(_ context.Context, s *grpc.Server) error {
	return service.RegisterGRPC(srv, s,
		service.WithAPIpathPrefixes(srv.PathPrefixes...),
		service.WithAdditionalServiceNames(srv.AdditionalServiceNames...),
	)
}

// RegisterProxyGW impl server.APIGatewayProxy
func (srv *hostService) RegisterProxyGW(ctx context.Context, mux *grpcRt.ServeMux, c *grpc.ClientConn) error {
	return sgv1.RegisterSGroupsHostsAPIHandler(ctx, mux, c)
}
