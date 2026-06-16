package validators

import (
	"testing"

	"github.com/PRO-Robotech/sgroups-proto/pkg/api/common"
	sgv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
	"github.com/stretchr/testify/require"
)

func mkServiceWithIcmpTypes(types ...uint32) *sgv1.Service {
	return &sgv1.Service{
		Spec: &sgv1.Service_Spec{
			Transports: []*common.Transport{
				{
					Protocol: common.Transport_ICMP,
					Ipv:      common.IpAddrFamily_IPV6,
					Entries:  []*common.Transport_Entry{{Types: types}},
				},
			},
		},
	}
}

func Test_validateServiceReqUpsert_IcmpTypeRange(t *testing.T) {
	tests := []struct {
		name    string
		types   []uint32
		wantErr bool
	}{
		{"in-range single", []uint32{8}, false},
		{"in-range multiple", []uint32{0, 128, 255}, false},
		{"boundary 256 (uint8 wrap)", []uint32{256}, true},
		{"large value", []uint32{4242}, true},
		{"mixed valid + invalid", []uint32{8, 256}, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := &sgv1.ServiceReq_Upsert{Services: []*sgv1.Service{mkServiceWithIcmpTypes(tc.types...)}}
			err := validateServiceReqUpsert(req)
			if tc.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), "ICMP type")
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func Test_validateServiceReqUpsert_PathInError(t *testing.T) {
	// Second service, first transport, ICMP types overflow.
	req := &sgv1.ServiceReq_Upsert{Services: []*sgv1.Service{
		mkServiceWithIcmpTypes(8),
		mkServiceWithIcmpTypes(300),
	}}
	err := validateServiceReqUpsert(req)
	require.Error(t, err)
	require.Contains(t, err.Error(), "services[1].spec.transports[0]")
}
