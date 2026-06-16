package network

import (
	"context"

	"github.com/PRO-Robotech/sgroups/internal/sg-server/repository"
	"github.com/PRO-Robotech/sgroups/internal/sg-server/transport/sgroups/grpc/service"
	"github.com/PRO-Robotech/sgroups/internal/sg-server/transport/sgroups/grpc/service/network/dto"
	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"
	"github.com/PRO-Robotech/sgroups/internal/shared/transport"

	sgv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
)

// Upsert - create or update networks
func (srv *networkService) Upsert(ctx context.Context, req *sgv1.NetworkReq_Upsert) (resp *sgv1.NetworkResp_Upsert, err error) {
	defer func() {
		err = transport.CorrectError(err, service.WithAllErr[:]...)
	}()

	var networks domain.Networks
	if err = dto.Proto2Domain(dto.DTO(req, &networks)); err != nil {
		return resp, err
	}
	if err = transport.Validate(networks...); err != nil {
		return resp, err
	}
	var nwResp []domain.Network
	err = srv.rep.Do(ctx, func(wr repository.WriterFace) (e error) {
		nwResp, e = wr.SyncNetwork(ctx, networks, repository.UpsertOp)
		return e
	})
	if err != nil {
		return resp, err
	}

	resp = new(sgv1.NetworkResp_Upsert)
	err = dto.Domain2Proto(dto.DTO(domain.Networks(nwResp), &resp))

	return resp, err
}
