package dto

import (
	"net/netip"
	"testing"

	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"
	"github.com/H-BF/corlib/pkg/dict"
	"github.com/google/uuid"
	pb "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
	"github.com/stretchr/testify/require"
)

func Test_UpdHealthStatus_Domain2Proto_HealthyTrue(t *testing.T) {
	uid := uuid.MustParse("aaaa1111-1111-1111-1111-111111111111")

	src := domain.Hosts{
		{
			Metadata: domain.ResMetadata{
				ID: domain.NamespacedMetadataIdentity{
					ClusterScopeMetadataIdentity: domain.ClusterScopeMetadataIdentity{
						UID:  uid,
						Name: domain.ResourceName("h-health"),
					},
					Namespace: domain.ResourceNamespace("ns-1"),
				},
			},
			Spec: domain.HostSpec{
				Healthy: true,
			},
		},
	}

	dest, err := updHealthToProto(src)
	require.NoError(t, err)
	require.Len(t, dest.Hosts, 1)
	require.Equal(t, pb.Healthy_HEALTHY_TRUE, dest.Hosts[0].GetSpec().GetHealthy())
	require.Equal(t, uid.String(), dest.Hosts[0].GetMetadata().GetUid())
	require.Equal(t, "h-health", dest.Hosts[0].GetMetadata().GetName())
}

func Test_UpdHealthStatus_Domain2Proto_HealthyFalse(t *testing.T) {
	src := domain.Hosts{
		{
			Metadata: domain.ResMetadata{
				ID: domain.NamespacedMetadataIdentity{
					ClusterScopeMetadataIdentity: domain.ClusterScopeMetadataIdentity{
						UID:  uuid.MustParse("bbbb2222-2222-2222-2222-222222222222"),
						Name: domain.ResourceName("h-unhealthy"),
					},
					Namespace: domain.ResourceNamespace("ns-2"),
				},
			},
			Spec: domain.HostSpec{
				Healthy: false,
			},
		},
	}

	dest, err := updHealthToProto(src)
	require.NoError(t, err)
	require.Len(t, dest.Hosts, 1)
	require.Equal(t, pb.Healthy_HEALTHY_FALSE, dest.Hosts[0].GetSpec().GetHealthy())
}

func Test_UpdHealthStatus_Domain2Proto_Empty(t *testing.T) {
	dest, err := updHealthToProto(nil)
	require.NoError(t, err)
	require.Empty(t, dest.Hosts)
}

func Test_SpecToProto_IncludesHealthy(t *testing.T) {
	ip1 := netip.MustParseAddr("10.0.0.1")
	src := domain.HostSpec{
		Healthy: true,
		IPs: domain.DualStackIPs{
			IPv4: dict.MakeHSet(ip1),
		},
	}

	dest, err := specToProto(src)
	require.NoError(t, err)
	require.Equal(t, pb.Healthy_HEALTHY_TRUE, dest.GetHealthy())
}
