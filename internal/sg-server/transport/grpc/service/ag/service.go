package ag

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

// NewSgAddressGroupService creates service
func NewSgAddressGroupService(ctx context.Context, r repository.Repository, opts ...service.Option) server.APIService {
	ret := &addressGroupService{
		appCtx: ctx,
		rep:    r,
	}
	ret.ApplyOptions(opts...)
	return ret
}

type addressGroupService struct {
	sgv1.UnimplementedSGroupsAddressGroupsAPIServer
	service.CommonOptions
	appCtx context.Context
	rep    repository.Repository
}

var (
	_ sgv1.SGroupsAddressGroupsAPIServer = (*addressGroupService)(nil)
	_ service.Service                    = (*addressGroupService)(nil)

	//SgAddressGroupSwaggerUtil ...
	SgAddressGroupSwaggerUtil sgPkg.SwaggerUtil[sgv1.SGroupsAddressGroupsAPIServer]
)

// Description impl server.APIService
func (srv *addressGroupService) Description() grpc.ServiceDesc {
	return sgv1.SGroupsAddressGroupsAPI_ServiceDesc
}

// RegisterGRPC impl server/APIService
func (srv *addressGroupService) RegisterGRPC(_ context.Context, s *grpc.Server) error {
	return service.RegisterGRPC(srv, s,
		service.WithAPIpathPrefixes(srv.PathPrefixes...),
		service.WithAdditionalServiceNames(srv.AdditionalServiceNames...),
	)
}

// RegisterProxyGW impl server.APIGatewayProxy
func (srv *addressGroupService) RegisterProxyGW(ctx context.Context, mux *grpcRt.ServeMux, c *grpc.ClientConn) error {
	return sgv1.RegisterSGroupsAddressGroupsAPIHandler(ctx, mux, c)
}
