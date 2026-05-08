package validators

import (
	"testing"

	"github.com/PRO-Robotech/sgroups-proto/pkg/api/common"
	"github.com/stretchr/testify/require"
)

func Test_ValidateTransport_NilTransport_OK(t *testing.T) {
	require.NoError(t, validateTransport(nil))
}

func Test_ValidateTransport_ICMP_WithPorts_Rejects(t *testing.T) {
	tr := &common.Transport{
		Protocol: common.Transport_ICMP,
		Entries: []*common.Transport_Entry{
			{Description: "e0", Ports: "8080", Types: []uint32{0, 8}},
		},
	}
	err := validateTransport(tr)
	require.Error(t, err)
	require.Contains(t, err.Error(), "ports field is not allowed for ICMP")
	require.Contains(t, err.Error(), "entries[0]")
}

func Test_ValidateTransport_TCPUDP_WithTypes_Rejects(t *testing.T) {
	tests := []struct {
		name  string
		proto common.Transport_Protocol
	}{
		{"TCP", common.Transport_TCP},
		{"UDP", common.Transport_UDP},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tr := &common.Transport{
				Protocol: tc.proto,
				Entries: []*common.Transport_Entry{
					{Ports: "8080-9090", Types: []uint32{1, 2}},
				},
			}
			err := validateTransport(tr)
			require.Error(t, err)
			require.Contains(t, err.Error(), "types field is not allowed for TCP/UDP")
		})
	}
}

func Test_ValidateTransport_ICMP_TypeOverflow_Rejects(t *testing.T) {
	tests := []struct {
		name  string
		types []uint32
	}{
		{"256", []uint32{256}},
		{"257", []uint32{257}},
		{"65535", []uint32{65535}},
		{"mixed valid and overflow", []uint32{1, 256, 8}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tr := &common.Transport{
				Protocol: common.Transport_ICMP,
				Entries:  []*common.Transport_Entry{{Types: tc.types}},
			}
			err := validateTransport(tr)
			require.Error(t, err)
			require.Contains(t, err.Error(), "exceeds max 255")
		})
	}
}

func Test_ValidateTransport_ErrorPathIncludesIndex(t *testing.T) {
	tr := &common.Transport{
		Protocol: common.Transport_TCP,
		Entries: []*common.Transport_Entry{
			{Ports: "80"},                         // valid
			{Ports: "443", Types: []uint32{1, 2}}, // invalid (types with TCP)
		},
	}
	err := validateTransport(tr)
	require.Error(t, err)
	require.Contains(t, err.Error(), "entries[1]")
}

func Test_ValidateTransport_HappyPaths(t *testing.T) {
	tests := []struct {
		name string
		tr   *common.Transport
	}{
		{"TCP with ports only", &common.Transport{
			Protocol: common.Transport_TCP,
			Entries:  []*common.Transport_Entry{{Ports: "80-443"}},
		}},
		{"UDP with ports only", &common.Transport{
			Protocol: common.Transport_UDP,
			Entries:  []*common.Transport_Entry{{Ports: "53"}},
		}},
		{"ICMP with types only", &common.Transport{
			Protocol: common.Transport_ICMP,
			Entries:  []*common.Transport_Entry{{Types: []uint32{0, 8, 255}}},
		}},
		{"ICMP with empty entries", &common.Transport{
			Protocol: common.Transport_ICMP,
			Entries:  nil,
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.NoError(t, validateTransport(tc.tr))
		})
	}
}
