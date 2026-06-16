package network

import (
	"context"

	"github.com/PRO-Robotech/sgroups/internal/sg-server/repository"
	"github.com/PRO-Robotech/sgroups/internal/sg-server/transport/sgroups/grpc/service"
	"github.com/PRO-Robotech/sgroups/internal/sg-server/transport/sgroups/grpc/service/network/dto"
	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"
	"github.com/PRO-Robotech/sgroups/internal/shared/transport"

	sgv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Delete -
func (srv *networkService) Delete(ctx context.Context, req *sgv1.NetworkReq_Delete) (resp *emptypb.Empty, err error) {
	defer func() {
		err = transport.CorrectError(err, service.WithAllErr[:]...)
	}()
	resp = new(emptypb.Empty)

	var networks domain.Networks
	if err = dto.Proto2Domain(dto.DTO(req, &networks)); err != nil {
		return resp, err
	}

	if err = transport.Validate(networks.GetMetas()...); err != nil {
		return resp, err
	}

	err = srv.rep.Do(ctx, func(wr repository.WriterFace) (e error) {
		_, e = wr.SyncNetwork(ctx, networks, repository.DeleteOp)
		return e
	})

	return resp, err
}
