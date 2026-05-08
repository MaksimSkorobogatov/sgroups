package network_binding

import (
	"context"

	"github.com/PRO-Robotech/sgroups/internal/sg-server/repository"
	"github.com/PRO-Robotech/sgroups/internal/sg-server/transport/grpc/service"
	"github.com/PRO-Robotech/sgroups/internal/sg-server/transport/grpc/service/network-binding/dto"
	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"

	sgv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
)

// Upsert - create or update network bindings
func (srv *nbService) Upsert(ctx context.Context, req *sgv1.NetworkBindingReq_Upsert) (resp *sgv1.NetworkBindingResp_Upsert, err error) {
	defer func() {
		err = service.CorrectError(err, service.WithAllErr[:]...)
	}()

	var networkBindings domain.NetworkBindings
	if err = dto.Proto2Domain(dto.DTO(req, &networkBindings)); err != nil {
		return resp, err
	}
	if err = service.Validate(networkBindings...); err != nil {
		return resp, err
	}
	var nbResp []domain.NetworkBinding
	err = srv.rep.Do(ctx, func(wr repository.WriterFace) (e error) {
		nbResp, e = wr.SyncNetworkBinding(ctx, networkBindings, repository.UpsertOp)
		return e
	})
	if err != nil {
		return resp, err
	}

	resp = new(sgv1.NetworkBindingResp_Upsert)
	err = dto.Domain2Proto(dto.DTO(domain.NetworkBindings(nbResp), &resp))

	return resp, err
}
