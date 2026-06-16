package namespace

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

// NewSgNamespaceService creates service
func NewSgNamespaceService(ctx context.Context, r repository.Repository, opts ...service.Option) server.APIService {
	ret := &namespaceService{
		appCtx: ctx,
		rep:    r,
	}
	ret.ApplyOptions(opts...)
	return ret
}

type namespaceService struct {
	sgv1.UnimplementedSGroupsNamespaceAPIServer
	service.CommonOptions
	appCtx context.Context
	rep    repository.Repository
}

var (
	_ sgv1.SGroupsNamespaceAPIServer = (*namespaceService)(nil)
	_ service.Service                = (*namespaceService)(nil)

	//SgNamespaceSwaggerUtil ...
	SgNamespaceSwaggerUtil sgPkg.SwaggerUtil[sgv1.SGroupsNamespaceAPIServer]
)

// Description impl server.APIService
func (srv *namespaceService) Description() grpc.ServiceDesc {
	return sgv1.SGroupsNamespaceAPI_ServiceDesc
}

// RegisterGRPC impl server.APIService
func (srv *namespaceService) RegisterGRPC(_ context.Context, s *grpc.Server) error {
	return service.RegisterGRPC(srv, s,
		service.WithAPIpathPrefixes(srv.PathPrefixes...),
		service.WithAdditionalServiceNames(srv.AdditionalServiceNames...),
	)
}

// RegisterProxyGW impl server.APIGatewayProxy
func (srv *namespaceService) RegisterProxyGW(ctx context.Context, mux *grpcRt.ServeMux, c *grpc.ClientConn) error {
	return sgv1.RegisterSGroupsNamespaceAPIHandler(ctx, mux, c)
}
