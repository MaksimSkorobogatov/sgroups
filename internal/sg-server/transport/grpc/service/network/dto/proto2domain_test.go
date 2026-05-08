package dto

import (
	"net"
	"testing"

	pb "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"
	"github.com/stretchr/testify/require"
)

func Test_Proto2Domain_NetworkSpec_ValidCIDR(t *testing.T) {
	src := &pb.Network_Spec{
		DisplayName: "network",
		Comment:     "comment",
		Description: "description",
		Cidr:        "10.0.0.0/24",
	}

	_, expected, err := net.ParseCIDR("10.0.0.0/24")
	require.NoError(t, err)

	var got domain.NetworkSpec
	err = Proto2Domain(DTO(src, &got))
	require.NoError(t, err)
	require.Equal(t, *expected, got.CIDR.IPNet)
}

func Test_Proto2Domain_NetworkSpec_BadCIDR(t *testing.T) {
	src := &pb.Network_Spec{Cidr: "bad"}
	var got domain.NetworkSpec

	err := Proto2Domain(DTO(src, &got))
	require.Error(t, err)
	require.Contains(t, err.Error(), "bad CIDR 'bad'")
}

func Test_Proto2Domain_NetworkSpec_HostBitsPreservesOriginalIP(t *testing.T) {
	src := &pb.Network_Spec{Cidr: "10.0.0.1/24"}
	var got domain.NetworkSpec

	err := Proto2Domain(DTO(src, &got))
	require.NoError(t, err)
	require.Equal(t, net.IP{10, 0, 0, 1}, got.CIDR.IP)
	require.Equal(t, net.CIDRMask(24, 32), net.IPMask(got.CIDR.Mask))
}

func Test_Proto2Domain_NetworkSpec_IPv6HostBitsPreservesOriginalIP(t *testing.T) {
	src := &pb.Network_Spec{Cidr: "2001:db8::1/32"}
	var got domain.NetworkSpec

	err := Proto2Domain(DTO(src, &got))
	require.NoError(t, err)
	require.Equal(t, net.ParseIP("2001:db8::1").To16(), got.CIDR.IP)
	require.Equal(t, net.CIDRMask(32, 128), net.IPMask(got.CIDR.Mask))
}

// Тесты полного flow: DTO → Validate.
// Если в DTO заменить origIP на *ipnet, валидация перестанет ловить неканоничные CIDR.
func Test_Proto2Domain_Then_Validate_NonCanonicalIPv4_Error(t *testing.T) {
	src := &pb.Network_Spec{Cidr: "10.0.0.1/24"}
	var got domain.NetworkSpec

	err := Proto2Domain(DTO(src, &got))
	require.NoError(t, err)

	err = got.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "not in canonical form")
	require.Contains(t, err.Error(), "use '10.0.0.0/24'")
}

func Test_Proto2Domain_Then_Validate_NonCanonicalIPv6_Error(t *testing.T) {
	src := &pb.Network_Spec{Cidr: "2001:db8::1/32"}
	var got domain.NetworkSpec

	err := Proto2Domain(DTO(src, &got))
	require.NoError(t, err)

	err = got.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "not in canonical form")
	require.Contains(t, err.Error(), "use '2001:db8::/32'")
}

func Test_Proto2Domain_Then_Validate_CanonicalIPv4_OK(t *testing.T) {
	src := &pb.Network_Spec{Cidr: "10.0.0.0/24"}
	var got domain.NetworkSpec

	err := Proto2Domain(DTO(src, &got))
	require.NoError(t, err)

	require.NoError(t, got.Validate())
}

func Test_Proto2Domain_Then_Validate_CanonicalIPv6_OK(t *testing.T) {
	src := &pb.Network_Spec{Cidr: "2001:db8::/32"}
	var got domain.NetworkSpec

	err := Proto2Domain(DTO(src, &got))
	require.NoError(t, err)

	require.NoError(t, got.Validate())
}
