package service

import (
	"context"

	uc "github.com/PRO-Robotech/sgroups/internal/sg-agent/usecases/ss"
	"github.com/PRO-Robotech/sgroups/internal/shared/transport"

	agentv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/agent/v1"
)

// ListSocketStatistics -
func (srv *agentAPI) ListSocketStatistics(ctx context.Context, req *agentv1.SocketStatReq_List) (resp *agentv1.SocketStatResp_List, err error) {
	defer func() {
		err = transport.CorrectError(err, transport.WithValidationErr, transport.WithUsecaseErr)
	}()
	if err = ValidateSelectors(req.GetSelectors()); err != nil {
		return nil, err
	}
	u := uc.NewListUseCase(srv.ssCacheTTL)

	return u.List(ctx, req)
}
