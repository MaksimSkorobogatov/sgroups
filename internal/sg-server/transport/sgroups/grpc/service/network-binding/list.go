package network_binding

import (
	"context"

	"github.com/PRO-Robotech/sgroups/internal/sg-server/repository"
	"github.com/PRO-Robotech/sgroups/internal/sg-server/transport/sgroups/grpc/service"
	"github.com/PRO-Robotech/sgroups/internal/sg-server/transport/sgroups/grpc/service/network-binding/dto"
	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"
	"github.com/PRO-Robotech/sgroups/internal/shared/transport"

	sgv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
)

// List - returns list of network bindings
func (srv *nbService) List(ctx context.Context, req *sgv1.NetworkBindingReq_List) (resp *sgv1.NetworkBindingResp_List, err error) {
	defer func() {
		err = transport.CorrectError(err, service.WithAllErr[:]...)
	}()
	var rd repository.Reader
	if rd, err = srv.rep.Reader(ctx); err != nil {
		return resp, err
	}
	defer func() { _ = rd.Close() }()

	var sel domain.NetworkBindingSelectorList
	if err = dto.Proto2Domain(dto.DTO(req, &sel)); err != nil {
		return resp, err
	}
	var nbResp domain.NetworkBindingList
	if nbResp.Items, err = rd.ListNetworkBindings(ctx, sel); err != nil {
		return resp, err
	}

	if nbResp.ResourceVersion, err = rd.GetResourceVersion(ctx); err != nil {
		return resp, err
	}

	resp = new(sgv1.NetworkBindingResp_List)
	err = dto.Domain2Proto(dto.DTO(nbResp, &resp))

	return resp, err
}
