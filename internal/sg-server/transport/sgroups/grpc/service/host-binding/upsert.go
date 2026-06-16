package host_binding

import (
	"context"

	"github.com/PRO-Robotech/sgroups/internal/sg-server/repository"
	"github.com/PRO-Robotech/sgroups/internal/sg-server/transport/sgroups/grpc/service"
	"github.com/PRO-Robotech/sgroups/internal/sg-server/transport/sgroups/grpc/service/host-binding/dto"
	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"
	"github.com/PRO-Robotech/sgroups/internal/shared/transport"

	sgv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
)

// Upsert - create or update host bindings
func (srv *hbService) Upsert(ctx context.Context, req *sgv1.HostBindingReq_Upsert) (resp *sgv1.HostBindingResp_Upsert, err error) {
	defer func() {
		err = transport.CorrectError(err, service.WithAllErr[:]...)
	}()

	var hostBindings domain.HostBindings
	if err = dto.Proto2Domain(dto.DTO(req, &hostBindings)); err != nil {
		return resp, err
	}
	if err = transport.Validate(hostBindings...); err != nil {
		return resp, err
	}
	var hbResp []domain.HostBinding
	err = srv.rep.Do(ctx, func(wr repository.WriterFace) (e error) {
		hbResp, e = wr.SyncHostBinding(ctx, hostBindings, repository.UpsertOp)
		return e
	})
	if err != nil {
		return resp, err
	}

	resp = new(sgv1.HostBindingResp_Upsert)
	err = dto.Domain2Proto(dto.DTO(domain.HostBindings(hbResp), &resp))

	return resp, err
}
