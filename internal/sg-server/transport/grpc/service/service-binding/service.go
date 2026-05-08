package service_binding

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

// NewSgServiceBindingService creates service
func NewSgServiceBindingService(ctx context.Context, r repository.Repository, opts ...service.Option) server.APIService {
	ret := &sbService{
		appCtx: ctx,
		rep:    r,
	}
	ret.ApplyOptions(opts...)
	return ret
}

type sbService struct {
	sgv1.UnimplementedSGroupsServiceBindingAPIServer
	service.CommonOptions
	appCtx context.Context
	rep    repository.Repository
}

var (
	_ sgv1.SGroupsServiceBindingAPIServer = (*sbService)(nil)
	_ service.Service                     = (*sbService)(nil)

	//SgServiceBindingSwaggerUtil ...
	SgServiceBindingSwaggerUtil sgPkg.SwaggerUtil[sgv1.SGroupsServiceBindingAPIServer]
)

// Description impl server.APIService
func (srv *sbService) Description() grpc.ServiceDesc {
	return sgv1.SGroupsServiceBindingAPI_ServiceDesc
}

// RegisterGRPC impl server/APIService
func (srv *sbService) RegisterGRPC(_ context.Context, s *grpc.Server) error {
	return service.RegisterGRPC(srv, s,
		service.WithAPIpathPrefixes(srv.PathPrefixes...),
		service.WithAdditionalServiceNames(srv.AdditionalServiceNames...),
	)
}

// RegisterProxyGW impl server.APIGatewayProxy
func (srv *sbService) RegisterProxyGW(ctx context.Context, mux *grpcRt.ServeMux, c *grpc.ClientConn) error {
	return sgv1.RegisterSGroupsServiceBindingAPIHandler(ctx, mux, c)
}
