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

// List - returns list of address groups
func (srv *addressGroupService) List(ctx context.Context, req *sgv1.AddressGroupReq_List) (resp *sgv1.AddressGroupResp_List, err error) {
	defer func() {
		err = transport.CorrectError(err, service.WithAllErr[:]...)
	}()
	var rd repository.Reader
	if rd, err = srv.rep.Reader(ctx); err != nil {
		return resp, err
	}
	defer func() { _ = rd.Close() }()

	var resSelector domain.ResSelectorList
	if err = dto.Proto2Domain(dto.DTO(req, &resSelector)); err != nil {
		return resp, err
	}
	var nsResp domain.AddressGroupList
	if nsResp.Items, err = rd.ListAddressGroups(ctx, resSelector); err != nil {
		return resp, err
	}

	if nsResp.ResourceVersion, err = rd.GetResourceVersion(ctx); err != nil {
		return resp, err
	}

	resp = new(sgv1.AddressGroupResp_List)
	err = dto.Domain2Proto(dto.DTO(nsResp, &resp))

	return resp, err
}
