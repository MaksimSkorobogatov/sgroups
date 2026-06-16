package host

import (
	"context"

	"github.com/PRO-Robotech/sgroups/internal/sg-server/transport/sgroups/grpc/service"
	uc "github.com/PRO-Robotech/sgroups/internal/sg-server/transport/sgroups/grpc/service/host/usecases/ss"
	"github.com/PRO-Robotech/sgroups/internal/shared/transport"

	sgv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
)

// ListSocketStatistics - returns list of socket statistics for host
func (srv *hostService) ListSocketStatistics(ctx context.Context, req *sgv1.HostReq_SocketStatistics_List) (resp *sgv1.HostResp_SocketStatistics_List, err error) {
	defer func() {
		err = transport.CorrectError(err, service.WithAllErr[:]...)
	}()
	u := uc.ListSS{
		Rep:                 srv.rep,
		AgentClientProvider: srv.clientProvider,
	}
	return u.Perform(ctx, req)
}
