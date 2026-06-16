package service_binding

import (
	"context"

	"github.com/PRO-Robotech/sgroups/internal/sg-server/repository"
	"github.com/PRO-Robotech/sgroups/internal/sg-server/transport/sgroups/grpc/service"
	"github.com/PRO-Robotech/sgroups/internal/sg-server/transport/sgroups/grpc/service/service-binding/dto"
	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"
	"github.com/PRO-Robotech/sgroups/internal/shared/transport"

	sgv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Delete -
func (srv *sbService) Delete(ctx context.Context, req *sgv1.ServiceBindingReq_Delete) (resp *emptypb.Empty, err error) { //nolint:dupl
	defer func() {
		err = transport.CorrectError(err, service.WithAllErr[:]...)
	}()
	resp = new(emptypb.Empty)

	var serviceBindings domain.ServiceBindings
	if err = dto.Proto2Domain(dto.DTO(req, &serviceBindings)); err != nil {
		return resp, err
	}

	if err = transport.Validate(serviceBindings.GetMetas()...); err != nil {
		return resp, err
	}

	err = srv.rep.Do(ctx, func(wr repository.WriterFace) (e error) {
		_, e = wr.SyncServiceBinding(ctx, serviceBindings, repository.DeleteOp)
		return e
	})

	return resp, err
}
