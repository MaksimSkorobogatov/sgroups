package namespace

import (
	"context"

	"github.com/PRO-Robotech/sgroups/internal/sg-server/repository"
	"github.com/PRO-Robotech/sgroups/internal/sg-server/transport/grpc/service"
	"github.com/PRO-Robotech/sgroups/internal/sg-server/transport/grpc/service/namespace/dto"
	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"

	sgv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
)

// List -
func (srv *namespaceService) List(ctx context.Context, req *sgv1.NamespaceReq_List) (resp *sgv1.NamespaceResp_List, err error) {
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
	var nsResp domain.NamespaceList
	if nsResp.Items, err = rd.ListNamespaces(ctx, resSelector); err != nil {
		return resp, err
	}

	if nsResp.ResourceVersion, err = rd.GetResourceVersion(ctx); err != nil {
		return resp, err
	}

	resp = new(sgv1.NamespaceResp_List)
	err = dto.Domain2Proto(dto.DTO(nsResp, &resp))

	return resp, err
}
