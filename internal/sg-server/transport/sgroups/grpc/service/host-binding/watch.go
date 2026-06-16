package host_binding

import (
	"github.com/PRO-Robotech/sgroups/internal/sg-server/repository"
	"github.com/PRO-Robotech/sgroups/internal/sg-server/transport/sgroups/grpc/service"
	"github.com/PRO-Robotech/sgroups/internal/sg-server/transport/sgroups/grpc/service/host-binding/dto"
	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"
	"github.com/PRO-Robotech/sgroups/internal/shared/misc"
	"github.com/PRO-Robotech/sgroups/internal/shared/transport"

	sgv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
	"google.golang.org/grpc"
)

// Watch - watch for changes in host bindings
func (srv *hbService) Watch(req *sgv1.HostBindingReq_Watch, stream grpc.ServerStreamingServer[sgv1.HostBindingResp_Watch]) (err error) {
	defer func() {
		err = transport.CorrectError(err, service.WithAllErr[:]...)
	}()
	ctx, cancel := misc.AnyContext(srv.appCtx, stream.Context())
	defer cancel()

	var rd repository.Reader
	if rd, err = srv.rep.Reader(ctx); err != nil {
		return err
	}
	defer func() { _ = rd.Close() }()

	var scope repository.Scope
	if err = dto.Proto2Scope(dto.DTO(req, &scope)); err != nil {
		return err
	}

	if err = rd.WatchHostBindings(ctx, scope, func(hbe domain.HostBindingEvent) error {
		var msg *sgv1.HostBindingResp_Watch
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if err = dto.Domain2Proto(dto.DTO(hbe, &msg)); err != nil {
			return err
		}
		return stream.Send(msg)
	}); err != nil {
		return err
	}

	return err
}
