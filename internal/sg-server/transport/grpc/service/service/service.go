package service_api

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

// NewSgServiceService creates service
func NewSgServiceService(ctx context.Context, r repository.Repository, opts ...service.Option) server.APIService {
	ret := &svcService{
		appCtx: ctx,
		rep:    r,
	}
	ret.ApplyOptions(opts...)
	return ret
}

type svcService struct {
	sgv1.UnimplementedSGroupsServicesAPIServer
	service.CommonOptions
	appCtx context.Context
	rep    repository.Repository
}

var (
	_ sgv1.SGroupsServicesAPIServer = (*svcService)(nil)
	_ service.Service               = (*svcService)(nil)

	//SgServicesSwaggerUtil ...
	SgServicesSwaggerUtil sgPkg.SwaggerUtil[sgv1.SGroupsServicesAPIServer]
)

// Description impl server.APIService
func (srv *svcService) Description() grpc.ServiceDesc {
	return sgv1.SGroupsServicesAPI_ServiceDesc
}

// RegisterGRPC impl server/APIService
func (srv *svcService) RegisterGRPC(_ context.Context, s *grpc.Server) error {
	return service.RegisterGRPC(srv, s,
		service.WithAPIpathPrefixes(srv.PathPrefixes...),
		service.WithAdditionalServiceNames(srv.AdditionalServiceNames...),
	)
}

// RegisterProxyGW impl server.APIGatewayProxy
func (srv *svcService) RegisterProxyGW(ctx context.Context, mux *grpcRt.ServeMux, c *grpc.ClientConn) error {
	return sgv1.RegisterSGroupsServicesAPIHandler(ctx, mux, c)
}
