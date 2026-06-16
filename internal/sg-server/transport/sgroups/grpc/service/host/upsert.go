package host

import (
	"context"

	"github.com/PRO-Robotech/sgroups/internal/sg-server/repository"
	"github.com/PRO-Robotech/sgroups/internal/sg-server/repository/scopes"
	"github.com/PRO-Robotech/sgroups/internal/sg-server/transport/sgroups/grpc/service"
	"github.com/PRO-Robotech/sgroups/internal/sg-server/transport/sgroups/grpc/service/host/dto"
	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"
	"github.com/PRO-Robotech/sgroups/internal/shared/transport"

	"github.com/H-BF/corlib/pkg/filter"
	sgv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
)

// Upsert - create or update hosts
func (srv *hostService) Upsert(ctx context.Context, req *sgv1.HostReq_Upsert) (*sgv1.HostResp_Upsert, error) { //nolint:dupl
	return syncHosts(ctx, srv.rep, req,
		func(req *sgv1.HostReq_Upsert) (domain.Hosts, error) {
			var hosts domain.Hosts
			err := dto.Proto2Domain(dto.DTO(req, &hosts))
			return hosts, err
		},
		func(hosts domain.Hosts) (*sgv1.HostResp_Upsert, error) {
			resp := new(sgv1.HostResp_Upsert)
			err := dto.Domain2Proto(dto.DTO(hosts, &resp))
			return resp, err
		},
		func(hosts domain.Hosts) filter.Scope {
			return scopes.ScopeByHosts{Hosts: hosts}
		},
	)
}

func syncHosts[reqT, respT any](
	ctx context.Context,
	rep repository.Repository,
	req reqT,
	toDomain func(req reqT) (domain.Hosts, error),
	toProto func(domain.Hosts) (respT, error),
	toScope func(domain.Hosts) filter.Scope,
) (resp respT, err error) {
	defer func() {
		err = transport.CorrectError(err, service.WithAllErr[:]...)
	}()

	var hosts domain.Hosts
	if hosts, err = toDomain(req); err != nil {
		return resp, err
	}
	if err = transport.Validate(hosts...); err != nil {
		return resp, err
	}
	var hostResp []domain.Host

	err = rep.Do(ctx, func(wr repository.WriterFace) (e error) {
		hostResp, e = wr.SyncHost(ctx, toScope(hosts), repository.UpsertOp)
		return e
	})
	if err != nil {
		return resp, err
	}

	return toProto(domain.Hosts(hostResp))
}
