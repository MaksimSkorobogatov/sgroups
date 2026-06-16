package host

import (
	"github.com/PRO-Robotech/sgroups/internal/sg-server/transport/sgroups/grpc/service"
	uc "github.com/PRO-Robotech/sgroups/internal/sg-server/transport/sgroups/grpc/service/host/usecases/nft"
	"github.com/PRO-Robotech/sgroups/internal/shared/misc"
	"github.com/PRO-Robotech/sgroups/internal/shared/transport"

	sgv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
	"google.golang.org/grpc"
)

// WatchNft - watches nftables rules for host
func (srv *hostService) WatchNft(req *sgv1.HostReq_Nft_Watch, stream grpc.ServerStreamingServer[sgv1.HostResp_Nft_Watch]) (err error) {
	defer func() {
		err = transport.CorrectError(err, service.WithAllErr[:]...)
	}()
	ctx, cancel := misc.AnyContext(srv.appCtx, stream.Context())
	defer cancel()

	u := uc.WatchSS{
		Rep:                 srv.rep,
		AgentClientProvider: srv.clientProvider,
		OutStream:           stream,
	}
	return u.Perform(ctx, req)
}
