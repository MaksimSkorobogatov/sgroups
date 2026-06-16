package ag

import (
	"context"

	"github.com/PRO-Robotech/sgroups/internal/sg-server/repository"
	"github.com/PRO-Robotech/sgroups/internal/sg-server/transport/sgroups/grpc/service"
	"github.com/PRO-Robotech/sgroups/internal/sg-server/transport/sgroups/grpc/service/ag/dto"
	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"
	"github.com/PRO-Robotech/sgroups/internal/shared/transport"

	sgv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
)

// Upsert - create or update address groups
func (srv *addressGroupService) Upsert(ctx context.Context, req *sgv1.AddressGroupReq_Upsert) (resp *sgv1.AddressGroupResp_Upsert, err error) {
	defer func() {
		err = transport.CorrectError(err, service.WithAllErr[:]...)
	}()

	var addressGroups domain.AddressGroups
	if err = dto.Proto2Domain(dto.DTO(req, &addressGroups)); err != nil {
		return resp, err
	}
	if err = transport.Validate(addressGroups...); err != nil {
		return resp, err
	}

	var agResp []domain.AddressGroup
	err = srv.rep.Do(ctx, func(wr repository.WriterFace) (e error) {
		agResp, e = wr.SyncAddressGroup(ctx, addressGroups, repository.UpsertOp)
		return e
	})
	if err != nil {
		return resp, err
	}

	resp = new(sgv1.AddressGroupResp_Upsert)
	err = dto.Domain2Proto(dto.DTO(domain.AddressGroups(agResp), &resp))

	return resp, err
}
