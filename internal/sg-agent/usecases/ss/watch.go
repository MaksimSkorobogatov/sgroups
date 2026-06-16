package ss

import (
	"context"
	"slices"
	"time"

	ss "github.com/PRO-Robotech/sgroups/internal/sg-agent/socketscan"
	"github.com/PRO-Robotech/sgroups/internal/shared/misc"
	"github.com/PRO-Robotech/sgroups/internal/shared/usecases"

	"github.com/H-BF/corlib/pkg/filter"
	agentv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/agent/v1"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

type (
	// WatchUseCase use case for watching socket statistics
	WatchUseCase interface {
		Watch(ctx context.Context, req *agentv1.SocketStatReq_Watch) (err error)
	}

	watchUseCase struct {
		stream     grpc.ServerStreamingServer[agentv1.SocketStatResp_Watch]
		ssCacheTTL time.Duration
	}
)

// NewWatchUseCase creates a new instance of WatchUseCase
func NewWatchUseCase(stream grpc.ServerStreamingServer[agentv1.SocketStatResp_Watch], ssCacheTTL time.Duration) WatchUseCase {
	return &watchUseCase{stream: stream, ssCacheTTL: ssCacheTTL}
}

// Watch - watches list of socket statistics according to selectors and updates it on change
func (uc *watchUseCase) Watch(ctx context.Context, req *agentv1.SocketStatReq_Watch) (err error) {
	ctx, cancel := misc.AnyContext(ctx, uc.stream.Context())
	defer cancel()

	var sc filter.Scope
	if sc, err = scopeFromeReq(req); err != nil {
		return usecases.InvalidArgument{Err: err}
	}

	opts := []ss.SSopt{ss.WithScope(sc)}
	if uc.ssCacheTTL > 0 {
		opts = append(opts, ss.WithCached(uc.ssCacheTTL))
	}

	t := time.NewTicker(time.Second)
	defer t.Stop()

	var lastSS []ss.SocketInfo

loop:
	for err == nil {
		select {
		case <-ctx.Done():
			break loop
		case <-t.C:
			var si []ss.SocketInfo
			if si, err = ss.ScanSockets(opts...); err != nil {
				err = errors.WithMessage(err, "scan socket failed")
				break loop
			}
			if len(si) == 0 {
				continue
			}
			slices.SortFunc(si, ss.SocketInfo.Cmp)
			if slices.EqualFunc(lastSS, si, ss.SocketInfo.IsEq) {
				continue
			}
			var siPb []*agentv1.SockStat
			if siPb, err = socketInfoToPb(si); err != nil {
				break loop
			}
			if err = uc.stream.Send(&agentv1.SocketStatResp_Watch{Stats: siPb}); err != nil {
				break loop
			}
			lastSS = si
		}
	}

	if err != nil {
		err = usecases.InternalError{Err: err}
	}

	return err
}
