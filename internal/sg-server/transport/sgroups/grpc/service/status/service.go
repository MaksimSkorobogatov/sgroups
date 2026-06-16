package ag

import (
	"context"

	"github.com/PRO-Robotech/sgroups/internal/sg-server/repository"
	"github.com/PRO-Robotech/sgroups/internal/sg-server/transport/sgroups/grpc/service"

	"github.com/H-BF/corlib/server"
	sgPkg "github.com/PRO-Robotech/sgroups-proto/pkg"
	sgv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
	grpcRt "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// NewSgStatusService creates service
func NewSgStatusService(ctx context.Context, r repository.Repository, opts ...service.Option) server.APIService {
	ret := &statusService{
		appCtx: ctx,
		rep:    r,
	}
	ret.ApplyOptions(opts...)
	return ret
}

type statusService struct {
	sgv1.UnimplementedSGroupsStatusAPIServer
	service.CommonOptions
	appCtx context.Context
	rep    repository.Repository
}

var (
	_ sgv1.SGroupsStatusAPIServer = (*statusService)(nil)
	_ service.Service             = (*statusService)(nil)

	//SgStatusSwaggerUtil ...
	SgStatusSwaggerUtil sgPkg.SwaggerUtil[sgv1.SGroupsStatusAPIServer]

	errServiceIsClosing = status.Error(codes.Unavailable,
		"'sgroups' service is about to be closed")
)

// Description impl server.APIService
func (srv *statusService) Description() grpc.ServiceDesc {
	return sgv1.SGroupsStatusAPI_ServiceDesc
}

// RegisterGRPC impl server/APIService
func (srv *statusService) RegisterGRPC(_ context.Context, s *grpc.Server) error {
	return service.RegisterGRPC(srv, s,
		service.WithAPIpathPrefixes(srv.PathPrefixes...),
		service.WithAdditionalServiceNames(srv.AdditionalServiceNames...),
	)
}

// RegisterProxyGW impl server.APIGatewayProxy
func (srv *statusService) RegisterProxyGW(ctx context.Context, mux *grpcRt.ServeMux, c *grpc.ClientConn) error {
	return sgv1.RegisterSGroupsStatusAPIHandler(ctx, mux, c)
}
