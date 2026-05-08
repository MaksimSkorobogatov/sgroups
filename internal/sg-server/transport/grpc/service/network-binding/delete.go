package network_binding

import (
	"context"

	"github.com/PRO-Robotech/sgroups/internal/sg-server/repository"
	"github.com/PRO-Robotech/sgroups/internal/sg-server/transport/grpc/service"
	"github.com/PRO-Robotech/sgroups/internal/sg-server/transport/grpc/service/network-binding/dto"
	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"

	sgv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Delete -
func (srv *nbService) Delete(ctx context.Context, req *sgv1.NetworkBindingReq_Delete) (resp *emptypb.Empty, err error) {
	defer func() {
		err = service.CorrectError(err, service.WithAllErr[:]...)
	}()
	resp = new(emptypb.Empty)

	var networkBindings domain.NetworkBindings
	if err = dto.Proto2Domain(dto.DTO(req, &networkBindings)); err != nil {
		return resp, err
	}

	if err = service.Validate(networkBindings.GetMetas()...); err != nil {
		return resp, err
	}

	err = srv.rep.Do(ctx, func(wr repository.WriterFace) (e error) {
		_, e = wr.SyncNetworkBinding(ctx, networkBindings, repository.DeleteOp)
		return e
	})

	return resp, err
}
