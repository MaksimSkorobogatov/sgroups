package host

import (
	"context"

	"github.com/PRO-Robotech/sgroups/internal/sg-server/repository"
	"github.com/PRO-Robotech/sgroups/internal/sg-server/transport/sgroups/grpc/service"
	"github.com/PRO-Robotech/sgroups/internal/sg-server/transport/sgroups/grpc/service/host/dto"
	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"
	"github.com/PRO-Robotech/sgroups/internal/shared/transport"

	sgv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
)

// List - returns list of hosts
func (srv *hostService) List(ctx context.Context, req *sgv1.HostReq_List) (resp *sgv1.HostResp_List, err error) {
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
	var nsResp domain.HostList
	if nsResp.Items, err = rd.ListHosts(ctx, resSelector); err != nil {
		return resp, err
	}

	if nsResp.ResourceVersion, err = rd.GetResourceVersion(ctx); err != nil {
		return resp, err
	}

	resp = new(sgv1.HostResp_List)
	err = dto.Domain2Proto(dto.DTO(nsResp, &resp))

	return resp, err
}
