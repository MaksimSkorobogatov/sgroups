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

// List - returns list of service bindings
func (srv *sbService) List(ctx context.Context, req *sgv1.ServiceBindingReq_List) (resp *sgv1.ServiceBindingResp_List, err error) { //nolint:dupl
	defer func() {
		err = transport.CorrectError(err, service.WithAllErr[:]...)
	}()
	var rd repository.Reader
	if rd, err = srv.rep.Reader(ctx); err != nil {
		return resp, err
	}
	defer func() { _ = rd.Close() }()

	var sel domain.ServiceBindingSelectorList
	if err = dto.Proto2Domain(dto.DTO(req, &sel)); err != nil {
		return resp, err
	}
	var sbResp domain.ServiceBindingList
	if sbResp.Items, err = rd.ListServiceBindings(ctx, sel); err != nil {
		return resp, err
	}

	if sbResp.ResourceVersion, err = rd.GetResourceVersion(ctx); err != nil {
		return resp, err
	}

	resp = new(sgv1.ServiceBindingResp_List)
	err = dto.Domain2Proto(dto.DTO(sbResp, &resp))

	return resp, err
}
