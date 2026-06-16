package host

import (
	"github.com/PRO-Robotech/sgroups/internal/sg-server/transport/sgroups/grpc/service"
	uc "github.com/PRO-Robotech/sgroups/internal/sg-server/transport/sgroups/grpc/service/host/usecases/ss"
	"github.com/PRO-Robotech/sgroups/internal/shared/misc"
	"github.com/PRO-Robotech/sgroups/internal/shared/transport"

	sgv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
	"google.golang.org/grpc"
)

// WatchSocketStatistics - watch for changes in socket statistics for host
func (srv *hostService) WatchSocketStatistics(req *sgv1.HostReq_SocketStatistics_Watch, stream grpc.ServerStreamingServer[sgv1.HostResp_SocketStatistics_Watch]) (err error) {
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
