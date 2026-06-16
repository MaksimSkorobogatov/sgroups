package service

import (
	"math"

	"github.com/PRO-Robotech/sgroups/internal/shared/transport"

	agentv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/agent/v1"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

// ValidateSelectors validates the selectors in the request
func ValidateSelectors(req []*agentv1.SockStat_Selectors) (err error) {
	defer func() {
		if err != nil {
			err = multierr.Combine(err, transport.ErrValidate)
		}
	}()
	for _, s := range req {
		if p := s.GetLocalPort(); p < 0 || p > math.MaxUint16 {
			return errors.Errorf("invalid local port: %d", p)
		}
		if p := s.GetRemotePort(); p < 0 || p > math.MaxUint16 {
			return errors.Errorf("invalid remote port: %d", p)
		}
		if i := s.GetInode(); i < 0 {
			return errors.Errorf("invalid inode: %d", i)
		}
		if pid := s.GetPid(); pid < 0 {
			return errors.Errorf("invalid pid: %d", pid)
		}
	}
	return nil
}
