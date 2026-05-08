package validators

import (
	"net"

	"github.com/PRO-Robotech/sgroups-proto/pkg/api/common"
	sgv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
	"github.com/pkg/errors"
)

func validateRuleReqUpsert(req *sgv1.RuleReq_Upsert) error {
	for i, rule := range req.GetRules() {
		if err := validateRuleSpec(rule.GetSpec()); err != nil {
			return errors.Wrapf(err, "rules[%d]", i)
		}
	}

	return nil
}

func validateRuleSpec(spec *sgv1.Rule_Spec) error {
	traffic := spec.GetSession().GetTraffic()
	if spec.GetSession() == nil || traffic == common.Session_TRAFFIC_UNDEF {
		return errors.New("spec.session.traffic must be set explicitly (TRAFFIC_UNDEF not allowed)")
	}

	if err := validateEndpointsRemote(spec.GetEndpoints().GetRemote()); err != nil {
		return err
	}

	localType := spec.GetEndpoints().GetLocal().GetType()
	remoteType := spec.GetEndpoints().GetRemote().GetType()

	destIsService := false
	switch traffic {
	case common.Session_INGRESS:
		destIsService = localType == common.Endpoints_SERVICE
	case common.Session_EGRESS, common.Session_BOTH:
		destIsService = remoteType == common.Endpoints_SERVICE
	}

	if destIsService {
		if spec.GetTransport() != nil {
			return errors.Errorf(
				"spec.transport must not be set when traffic destination is a Service "+
					"(traffic=%s, local=%s, remote=%s)",
				traffic, localType, remoteType)
		}
	} else if remoteType != common.Endpoints_UNKNOWN {
		if spec.GetTransport() == nil {
			return errors.Errorf(
				"spec.transport is required when traffic destination is not a Service "+
					"(traffic=%s, local=%s, remote=%s)",
				traffic, localType, remoteType)
		}
	}

	if err := validateTransport(spec.GetTransport()); err != nil {
		return errors.Wrap(err, "spec.transport")
	}

	return nil
}

func validateEndpointsRemote(r *common.Endpoints_Remote) error {
	if r == nil {
		return nil
	}

	switch r.GetType() {
	case common.Endpoints_UNKNOWN:
		if r.GetName() != "" || r.GetNamespace() != "" || r.GetValue() != "" {
			return errors.New(
				"spec.endpoints.remote.type is required; specify one of AddressGroup, Service, CIDR, FQDN")
		}
	case common.Endpoints_ADDRESS_GROUP, common.Endpoints_SERVICE:
		if r.GetValue() != "" {
			return errors.Errorf(
				"spec.endpoints.remote.value is not allowed for type %s", r.GetType())
		}
	case common.Endpoints_CIDR:
		if r.GetName() != "" || r.GetNamespace() != "" {
			return errors.Errorf(
				"spec.endpoints.remote.{name,namespace} not allowed for type %s", r.GetType())
		}
		if r.GetValue() == "" {
			return errors.New("spec.endpoints.remote.value is required for type CIDR")
		}

		if _, _, err := net.ParseCIDR(r.GetValue()); err != nil {
			return errors.Errorf(
				"spec.endpoints.remote.value %q is not a valid CIDR", r.GetValue())
		}
	case common.Endpoints_FQDN:
		if r.GetName() != "" || r.GetNamespace() != "" {
			return errors.Errorf(
				"spec.endpoints.remote.{name,namespace} not allowed for type %s", r.GetType())
		}
		if r.GetValue() == "" {
			return errors.New("spec.endpoints.remote.value is required for type FQDN")
		}
	}

	return nil
}
