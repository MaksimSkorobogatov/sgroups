package host_binding

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

// NewSgHostBindingService creates service
func NewSgHostBindingService(ctx context.Context, r repository.Repository, opts ...service.Option) server.APIService {
	ret := &hbService{
		appCtx: ctx,
		rep:    r,
	}
	ret.ApplyOptions(opts...)
	return ret
}

type hbService struct {
	sgv1.UnimplementedSGroupsHostBindingAPIServer
	service.CommonOptions
	appCtx context.Context
	rep    repository.Repository
}

var (
	_ sgv1.SGroupsHostBindingAPIServer = (*hbService)(nil)
	_ service.Service                  = (*hbService)(nil)

	//SgHostBindingSwaggerUtil ...
	SgHostBindingSwaggerUtil sgPkg.SwaggerUtil[sgv1.SGroupsHostBindingAPIServer]
)

// Description impl server.APIService
func (srv *hbService) Description() grpc.ServiceDesc {
	return sgv1.SGroupsHostBindingAPI_ServiceDesc
}

// RegisterGRPC impl server/APIService
func (srv *hbService) RegisterGRPC(_ context.Context, s *grpc.Server) error {
	return service.RegisterGRPC(srv, s,
		service.WithAPIpathPrefixes(srv.PathPrefixes...),
		service.WithAdditionalServiceNames(srv.AdditionalServiceNames...),
	)
}

// RegisterProxyGW impl server.APIGatewayProxy
func (srv *hbService) RegisterProxyGW(ctx context.Context, mux *grpcRt.ServeMux, c *grpc.ClientConn) error {
	return sgv1.RegisterSGroupsHostBindingAPIHandler(ctx, mux, c)
}
