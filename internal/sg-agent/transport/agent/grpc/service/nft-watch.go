package service

import (
	uc "github.com/PRO-Robotech/sgroups/internal/sg-agent/usecases/nft"
	"github.com/PRO-Robotech/sgroups/internal/shared/transport"

	agentv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/agent/v1"
	"google.golang.org/grpc"
)

// WatchNftables -
func (srv *agentAPI) WatchNftables(req *agentv1.NftablesReq_Watch, stream grpc.ServerStreamingServer[agentv1.NftablesResp_Watch]) (err error) {
	defer func() {
		err = transport.CorrectError(err, transport.WithUsecaseErr)
	}()

	u := uc.NewWatchUseCase(stream, srv.nftCacheTTL, srv.nftSyncInterval)
	return u.Watch(srv.appCtx, req)
}
