package host

import (
	"context"

	"github.com/PRO-Robotech/sgroups/internal/sg-server/repository/scopes"
	"github.com/PRO-Robotech/sgroups/internal/sg-server/transport/sgroups/grpc/service/host/dto"
	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"

	"github.com/H-BF/corlib/pkg/filter"
	sgv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
)

// UpdHealthStatus - update health check status of hosts
func (srv *hostService) UpdHealthStatus(ctx context.Context, req *sgv1.HostReq_UpdHealthStatus) (*sgv1.HostResp_UpdHealthStatus, error) {
	return syncHosts(ctx, srv.rep, req,
		func(req *sgv1.HostReq_UpdHealthStatus) (domain.Hosts, error) {
			var hosts domain.Hosts
			err := dto.Proto2Domain(dto.DTO(req, &hosts))
			return hosts, err
		},
		func(hosts domain.Hosts) (*sgv1.HostResp_UpdHealthStatus, error) {
			resp := new(sgv1.HostResp_UpdHealthStatus)
			err := dto.Domain2Proto(dto.DTO(hosts, &resp))
			return resp, err
		},
		func(hosts domain.Hosts) filter.Scope {
			return scopes.ScopeByHostHealth{Hosts: hosts}
		},
	)
}
