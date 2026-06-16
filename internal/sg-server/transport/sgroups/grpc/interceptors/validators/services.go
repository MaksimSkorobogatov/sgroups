package validators

import (
	sgv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
	"github.com/pkg/errors"
)

// validateServiceReqUpsert enforces structural invariants on services submitted
// via the Upsert RPC that aren't representable as buf.validate constraints —
// notably per-protocol entry shape and ICMP type range, reusing the same
// validateTransport helper as the rule path.
func validateServiceReqUpsert(req *sgv1.ServiceReq_Upsert) error {
	for i, svc := range req.GetServices() {
		for j, t := range svc.GetSpec().GetTransports() {
			if err := validateTransport(t); err != nil {
				return errors.Wrapf(err, "services[%d].spec.transports[%d]", i, j)
			}
		}
	}
	return nil
}
