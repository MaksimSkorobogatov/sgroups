package rules

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

// NewSgRulesService creates service
func NewSgRulesService(ctx context.Context, r repository.Repository, opts ...service.Option) server.APIService {
	ret := &rulesService{
		appCtx: ctx,
		rep:    r,
	}
	ret.ApplyOptions(opts...)
	return ret
}

type rulesService struct {
	sgv1.UnimplementedSGroupsRulesAPIServer
	service.CommonOptions
	appCtx context.Context
	rep    repository.Repository
}

var (
	_ sgv1.SGroupsRulesAPIServer = (*rulesService)(nil)
	_ service.Service            = (*rulesService)(nil)

	// SgRulesSwaggerUtil ...
	SgRulesSwaggerUtil sgPkg.SwaggerUtil[sgv1.SGroupsRulesAPIServer]
)

// Description impl server.APIService
func (srv *rulesService) Description() grpc.ServiceDesc {
	return sgv1.SGroupsRulesAPI_ServiceDesc
}

// RegisterGRPC impl server/APIService
func (srv *rulesService) RegisterGRPC(_ context.Context, s *grpc.Server) error {
	return service.RegisterGRPC(srv, s,
		service.WithAPIpathPrefixes(srv.PathPrefixes...),
		service.WithAdditionalServiceNames(srv.AdditionalServiceNames...),
	)
}

// RegisterProxyGW impl server.APIGatewayProxy
func (srv *rulesService) RegisterProxyGW(ctx context.Context, mux *grpcRt.ServeMux, c *grpc.ClientConn) error {
	return sgv1.RegisterSGroupsRulesAPIHandler(ctx, mux, c)
}
