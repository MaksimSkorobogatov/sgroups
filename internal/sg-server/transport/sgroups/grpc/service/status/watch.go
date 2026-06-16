package ag

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/PRO-Robotech/sgroups/internal/sg-server/repository"
	"github.com/PRO-Robotech/sgroups/internal/sg-server/transport/sgroups/grpc/service"
	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"
	"github.com/PRO-Robotech/sgroups/internal/shared/patterns"
	"github.com/PRO-Robotech/sgroups/internal/shared/transport"

	sgv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Watch - watch for changes in sync status
func (srv *statusService) Watch(_ *emptypb.Empty, stream grpc.ServerStreamingServer[sgv1.SyncStatusResp]) (err error) {
	defer func() {
		err = transport.CorrectError(err, service.WithAllErr[:]...)
	}()

	const updatePeriod = 3 * time.Second //TODO: In future move 'updatePeriod' onto config

	var commitCount int64
	var prevState *domain.SyncStatus
	commitCounter := func(_ patterns.EventType) {
		atomic.AddInt64(&commitCount, 1)
	}
	if subj := srv.rep.Subject(); subj != nil {
		obs := patterns.NewObserver(commitCounter, true, repository.DBUpdated{})
		subj.ObserversAttach(obs)
		defer subj.ObserversDetach(obs)
	}
	for ctx := stream.Context(); ; {
		if atomic.SwapInt64(&commitCount, 0) != 0 || prevState == nil {
			var newState domain.SyncStatus
			if newState, err = srv.getSyncStatus(ctx); err != nil {
				return err
			}
			doSend := (prevState == nil ||
				!newState.UpdatedAt.Equal(prevState.UpdatedAt))
			if doSend {
				resp := sgv1.SyncStatusResp{
					UpdatedAt: timestamppb.New(newState.UpdatedAt),
				}
				if err = stream.Send(&resp); err != nil {
					return nil
				}
				prevState = &newState
			}
		}
		select {
		case <-srv.appCtx.Done():
			return errServiceIsClosing
		case <-ctx.Done():
			return nil
		case <-time.After(updatePeriod):
		}
	}
}

func (srv *statusService) getSyncStatus(ctx context.Context) (ret domain.SyncStatus, err error) {
	var rd repository.Reader
	if rd, err = srv.rep.Reader(ctx); err != nil {
		return ret, err
	}
	defer func() { _ = rd.Close() }()

	return rd.GetSyncStatus(ctx)
}
