package host

import (
	"context"

	"github.com/PRO-Robotech/sgroups/internal/sg-server/transport/sgroups/grpc/service"
	uc "github.com/PRO-Robotech/sgroups/internal/sg-server/transport/sgroups/grpc/service/host/usecases/nft"
	"github.com/PRO-Robotech/sgroups/internal/shared/transport"

	sgv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
)

// ListNft - returns list of nftables rules for host
func (srv *hostService) ListNft(ctx context.Context, req *sgv1.HostReq_Nft_List) (resp *sgv1.HostResp_Nft_List, err error) {
	defer func() {
		err = transport.CorrectError(err, service.WithAllErr[:]...)
	}()
	u := uc.ListSS{
		Rep:                 srv.rep,
		AgentClientProvider: srv.clientProvider,
	}
	return u.Perform(ctx, req)
}
