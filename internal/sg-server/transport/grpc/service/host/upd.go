package host

import (
	"context"

	"github.com/PRO-Robotech/sgroups/internal/sg-server/repository/scopes"
	"github.com/PRO-Robotech/sgroups/internal/sg-server/transport/grpc/service/host/dto"
	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"

	"github.com/H-BF/corlib/pkg/filter"
	sgv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
)

// UpdIPs - update IP addresses of hosts
func (srv *hostService) UpdIPs(ctx context.Context, req *sgv1.HostReq_UpdIPs) (*sgv1.HostResp_UpdIPs, error) { //nolint:dupl
	return syncHosts(ctx, srv.rep, req,
		func(req *sgv1.HostReq_UpdIPs) (domain.Hosts, error) {
			var hosts domain.Hosts
			err := dto.Proto2Domain(dto.DTO(req, &hosts))
			return hosts, err
		},
		func(hosts domain.Hosts) (*sgv1.HostResp_UpdIPs, error) {
			resp := new(sgv1.HostResp_UpdIPs)
			err := dto.Domain2Proto(dto.DTO(hosts, &resp))
			return resp, err
		},
		func(hosts domain.Hosts) filter.Scope {
			return scopes.ScopeByHostIPs{Hosts: hosts}
		},
	)
}

// UpdMetaInfo - update meta information of hosts
func (srv *hostService) UpdMetaInfo(ctx context.Context, req *sgv1.HostReq_UpdMetaInfo) (*sgv1.HostResp_UpdMetaInfo, error) { //nolint:dupl
	return syncHosts(ctx, srv.rep, req,
		func(req *sgv1.HostReq_UpdMetaInfo) (domain.Hosts, error) {
			var hosts domain.Hosts
			err := dto.Proto2Domain(dto.DTO(req, &hosts))
			return hosts, err
		},
		func(hosts domain.Hosts) (*sgv1.HostResp_UpdMetaInfo, error) {
			resp := new(sgv1.HostResp_UpdMetaInfo)
			err := dto.Domain2Proto(dto.DTO(hosts, &resp))
			return resp, err
		},
		func(hosts domain.Hosts) filter.Scope {
			return scopes.ScopeByHostInfo{Hosts: hosts}
		},
	)
}
