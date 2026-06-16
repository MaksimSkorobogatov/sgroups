package ag

import (
	"context"

	"github.com/PRO-Robotech/sgroups/internal/sg-server/repository"
	"github.com/PRO-Robotech/sgroups/internal/sg-server/transport/sgroups/grpc/service"
	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"
	"github.com/PRO-Robotech/sgroups/internal/shared/transport"

	sgv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// List - returns sync status
func (srv *statusService) List(ctx context.Context, _ *emptypb.Empty) (resp *sgv1.SyncStatusResp, err error) {
	defer func() {
		err = transport.CorrectError(err, service.WithAllErr[:]...)
	}()
	var rd repository.Reader
	if rd, err = srv.rep.Reader(ctx); err != nil {
		return resp, err
	}
	defer func() { _ = rd.Close() }()

	var st domain.SyncStatus
	if st, err = srv.getSyncStatus(ctx); err != nil {
		return resp, err
	}

	return &sgv1.SyncStatusResp{UpdatedAt: timestamppb.New(st.UpdatedAt)}, nil
}
