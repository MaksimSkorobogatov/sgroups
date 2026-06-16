package validators

import (
	"math"

	"github.com/PRO-Robotech/sgroups-proto/pkg/api/common"
	"github.com/pkg/errors"
)

func validateTransport(t *common.Transport) error {
	if t == nil {
		return nil
	}
	protocol := t.GetProtocol()
	for i, et := range t.GetEntries() {
		if err := validateTransportEntry(protocol, et); err != nil {
			return errors.WithMessagef(err, "entries[%d]", i)
		}
	}

	return nil
}

func validateTransportEntry(proto common.Transport_Protocol, et *common.Transport_Entry) error {
	switch proto {
	case common.Transport_ICMP:
		if len(et.GetPorts()) != 0 {
			return errors.New("ports field is not allowed for ICMP transport entry")
		}
		for _, icmpType := range et.GetTypes() {
			if icmpType > math.MaxUint8 {
				return errors.Errorf("ICMP type value %d exceeds max %d", icmpType, math.MaxUint8)
			}
		}
	case common.Transport_TCP, common.Transport_UDP:
		if icmpTypes := et.GetTypes(); len(icmpTypes) > 0 {
			return errors.New("icmp types field is not allowed for TCP/UDP transport entry")
		}
	default:
		return errors.Errorf("unsupported protocol %v", proto)
	}

	return nil
}
