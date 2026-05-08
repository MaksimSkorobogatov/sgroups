package dto

import (
	"net"
	"testing"

	pb "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"
	"github.com/stretchr/testify/require"
)

func Test_Domain2Proto_NetworkSpec_NoMasking(t *testing.T) {
	src := domain.NetworkSpec{
		CIDR: domain.IPNet{
			IPNet: net.IPNet{
				IP:   net.ParseIP("10.0.0.1").To4(),
				Mask: net.CIDRMask(24, 32),
			},
		},
	}

	var got *pb.Network_Spec
	err := Domain2Proto(DTO(src, &got))
	require.NoError(t, err)
	require.Equal(t, "10.0.0.1/24", got.GetCidr())
}
