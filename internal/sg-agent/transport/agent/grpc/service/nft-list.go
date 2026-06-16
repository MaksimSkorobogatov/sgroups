package service

import (
	"context"

	uc "github.com/PRO-Robotech/sgroups/internal/sg-agent/usecases/nft"
	"github.com/PRO-Robotech/sgroups/internal/shared/transport"

	agentv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/agent/v1"
)

// ListNftables -
func (srv *agentAPI) ListNftables(ctx context.Context, req *agentv1.NftablesReq_List) (resp *agentv1.NftablesResp_List, err error) {
	defer func() {
		err = transport.CorrectError(err, transport.WithUsecaseErr)
	}()

	u := uc.NewListUseCase(srv.nftCacheTTL)
	return u.List(ctx, req)
}
