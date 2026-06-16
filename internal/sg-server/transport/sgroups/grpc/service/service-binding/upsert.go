package service_binding

import (
	"context"

	"github.com/PRO-Robotech/sgroups/internal/sg-server/repository"
	"github.com/PRO-Robotech/sgroups/internal/sg-server/transport/sgroups/grpc/service"
	"github.com/PRO-Robotech/sgroups/internal/sg-server/transport/sgroups/grpc/service/service-binding/dto"
	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"
	"github.com/PRO-Robotech/sgroups/internal/shared/transport"

	sgv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
)

// Upsert - create or update service bindings
func (srv *sbService) Upsert(ctx context.Context, req *sgv1.ServiceBindingReq_Upsert) (resp *sgv1.ServiceBindingResp_Upsert, err error) { //nolint:dupl
	defer func() {
		err = transport.CorrectError(err, service.WithAllErr[:]...)
	}()

	var serviceBindings domain.ServiceBindings
	if err = dto.Proto2Domain(dto.DTO(req, &serviceBindings)); err != nil {
		return resp, err
	}
	if err = transport.Validate(serviceBindings...); err != nil {
		return resp, err
	}

	var sbResp []domain.ServiceBinding
	err = srv.rep.Do(ctx, func(wr repository.WriterFace) (e error) {
		sbResp, e = wr.SyncServiceBinding(ctx, serviceBindings, repository.UpsertOp)
		return e
	})
	if err != nil {
		return resp, err
	}

	resp = new(sgv1.ServiceBindingResp_Upsert)
	err = dto.Domain2Proto(dto.DTO(domain.ServiceBindings(sbResp), &resp))

	return resp, err
}
