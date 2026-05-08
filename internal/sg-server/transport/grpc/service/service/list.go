package service_api

import (
	"context"

	"github.com/PRO-Robotech/sgroups/internal/sg-server/repository"
	"github.com/PRO-Robotech/sgroups/internal/sg-server/transport/grpc/service"
	"github.com/PRO-Robotech/sgroups/internal/sg-server/transport/grpc/service/service/dto"
	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"

	sgv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
)

// List - returns list of services
func (srv *svcService) List(ctx context.Context, req *sgv1.ServiceReq_List) (resp *sgv1.ServiceResp_List, err error) {
	defer func() {
		err = service.CorrectError(err, service.WithAllErr[:]...)
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
	var svcResp domain.ServiceList
	if svcResp.Items, err = rd.ListServices(ctx, resSelector); err != nil {
		return resp, err
	}

	if svcResp.ResourceVersion, err = rd.GetResourceVersion(ctx); err != nil {
		return resp, err
	}

	resp = new(sgv1.ServiceResp_List)
	err = dto.Domain2Proto(dto.DTO(svcResp, &resp))

	return resp, err
}
