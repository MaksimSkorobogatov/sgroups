package ss

import (
	"context"
	"time"

	ss "github.com/PRO-Robotech/sgroups/internal/sg-agent/socketscan"
	"github.com/PRO-Robotech/sgroups/internal/shared/usecases"

	"github.com/H-BF/corlib/pkg/filter"
	agentv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/agent/v1"
	"github.com/pkg/errors"
)

type (
	// ListUseCase use case for listing socket statistics
	ListUseCase interface {
		List(ctx context.Context, req *agentv1.SocketStatReq_List) (ret *agentv1.SocketStatResp_List, err error)
	}
	listUseCase struct {
		ssCacheTTL time.Duration
	}
)

// NewListUseCase creates a new instance of ListUseCase
func NewListUseCase(ssCacheTTL time.Duration) ListUseCase {
	return &listUseCase{ssCacheTTL: ssCacheTTL}
}

// List - returns list of socket statistics according to selectors
func (uc *listUseCase) List(ctx context.Context, req *agentv1.SocketStatReq_List) (ret *agentv1.SocketStatResp_List, err error) {
	var (
		si     []ss.SocketInfo
		sc     filter.Scope
		statPb []*agentv1.SockStat
	)
	if sc, err = scopeFromeReq(req); err != nil {
		return nil, usecases.InvalidArgument{Err: err}
	}
	opts := []ss.SSopt{ss.WithScope(sc)}
	if uc.ssCacheTTL > 0 {
		opts = append(opts, ss.WithCached(uc.ssCacheTTL))
	}
	if si, err = ss.ScanSockets(opts...); err != nil {
		return nil, usecases.InternalError{Err: errors.WithMessage(err, "scan socket failed")}
	}
	statPb, err = socketInfoToPb(si)
	if err != nil {
		err = usecases.InternalError{Err: errors.WithMessage(err, "convert socket info to protobuf failed")}
	}
	return &agentv1.SocketStatResp_List{Stats: statPb}, err
}
