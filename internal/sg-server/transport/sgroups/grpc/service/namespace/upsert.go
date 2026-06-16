package namespace

import (
	"context"

	"github.com/PRO-Robotech/sgroups/internal/sg-server/repository"
	"github.com/PRO-Robotech/sgroups/internal/sg-server/transport/sgroups/grpc/service"
	"github.com/PRO-Robotech/sgroups/internal/sg-server/transport/sgroups/grpc/service/namespace/dto"
	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"
	"github.com/PRO-Robotech/sgroups/internal/shared/transport"

	sgv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
)

// Upsert -
func (srv *namespaceService) Upsert(ctx context.Context, req *sgv1.NamespaceReq_Upsert) (resp *sgv1.NamespaceResp_Upsert, err error) {
	defer func() {
		err = transport.CorrectError(err, service.WithAllErr[:]...)
	}()

	var namespaces domain.Namespaces
	if err = dto.Proto2Domain(dto.DTO(req, &namespaces)); err != nil {
		return resp, err
	}
	if err = transport.Validate(namespaces...); err != nil {
		return resp, err
	}

	var nsResp []domain.Namespace
	err = srv.rep.Do(ctx, func(wr repository.WriterFace) (e error) {
		nsResp, e = wr.SyncNamespace(ctx, namespaces, repository.UpsertOp)
		return e
	})
	if err != nil {
		return resp, err
	}

	resp = new(sgv1.NamespaceResp_Upsert)
	err = dto.Domain2Proto(dto.DTO(domain.Namespaces(nsResp), &resp))

	return resp, err
}
