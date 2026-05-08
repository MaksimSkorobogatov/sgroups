package service_api

import (
	"context"

	"github.com/PRO-Robotech/sgroups/internal/sg-server/repository"
	"github.com/PRO-Robotech/sgroups/internal/sg-server/transport/grpc/service"
	"github.com/PRO-Robotech/sgroups/internal/sg-server/transport/grpc/service/service/dto"
	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"

	sgv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
)

// Upsert - create or update services
func (srv *svcService) Upsert(ctx context.Context, req *sgv1.ServiceReq_Upsert) (resp *sgv1.ServiceResp_Upsert, err error) {
	defer func() {
		err = service.CorrectError(err, service.WithAllErr[:]...)
	}()

	var services domain.Services
	if err = dto.Proto2Domain(dto.DTO(req, &services)); err != nil {
		return resp, err
	}
	if err = service.Validate(services...); err != nil {
		return resp, err
	}

	var svcResp []domain.Service
	err = srv.rep.Do(ctx, func(wr repository.WriterFace) (e error) {
		svcResp, e = wr.SyncService(ctx, services, repository.UpsertOp)
		return e
	})
	if err != nil {
		return resp, err
	}

	resp = new(sgv1.ServiceResp_Upsert)
	err = dto.Domain2Proto(dto.DTO(domain.Services(svcResp), &resp))

	return resp, err
}
