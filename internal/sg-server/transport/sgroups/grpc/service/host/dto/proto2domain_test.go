package dto

import (
	"testing"

	"github.com/PRO-Robotech/sgroups-proto/pkg/api/common"
	pb "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func Test_UpdHealthStatus_Proto2Domain_HealthyTrue(t *testing.T) {
	uid := uuid.MustParse("aaaa1111-1111-1111-1111-111111111111")

	src := &pb.HostReq_UpdHealthStatus{
		Hosts: []*pb.HostReq_UpdHealthStatus_Host{
			{
				Metadata: &common.MetadataScope{
					Uid:       uid.String(),
					Name:      "h-health",
					Namespace: "ns-1",
				},
				Spec: &pb.HostReq_UpdHealthStatus_Host_Spec{
					Healthy: true,
				},
			},
		},
	}

	var got domain.Hosts
	err := Proto2Domain(DTO(src, &got))
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Equal(t, uid, got[0].Metadata.ID.UID)
	require.Equal(t, domain.ResourceName("h-health"), got[0].Metadata.ID.Name)
	require.Equal(t, domain.ResourceNamespace("ns-1"), got[0].Metadata.ID.Namespace)
	require.True(t, got[0].Spec.Healthy)
}

func Test_UpdHealthStatus_Proto2Domain_HealthyFalse(t *testing.T) {
	uid := uuid.MustParse("bbbb2222-2222-2222-2222-222222222222")

	src := &pb.HostReq_UpdHealthStatus{
		Hosts: []*pb.HostReq_UpdHealthStatus_Host{
			{
				Metadata: &common.MetadataScope{
					Uid:       uid.String(),
					Name:      "h-unhealthy",
					Namespace: "ns-2",
				},
				Spec: &pb.HostReq_UpdHealthStatus_Host_Spec{
					Healthy: false,
				},
			},
		},
	}

	var got domain.Hosts
	err := Proto2Domain(DTO(src, &got))
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.False(t, got[0].Spec.Healthy)
}

func Test_UpdHealthStatus_Proto2Domain_Multiple(t *testing.T) {
	uid1 := uuid.MustParse("cccc3333-3333-3333-3333-333333333333")
	uid2 := uuid.MustParse("dddd4444-4444-4444-4444-444444444444")

	src := &pb.HostReq_UpdHealthStatus{
		Hosts: []*pb.HostReq_UpdHealthStatus_Host{
			{
				Metadata: &common.MetadataScope{Uid: uid1.String(), Name: "h-1", Namespace: "ns"},
				Spec:     &pb.HostReq_UpdHealthStatus_Host_Spec{Healthy: true},
			},
			{
				Metadata: &common.MetadataScope{Uid: uid2.String(), Name: "h-2", Namespace: "ns"},
				Spec:     &pb.HostReq_UpdHealthStatus_Host_Spec{Healthy: false},
			},
		},
	}

	var got domain.Hosts
	err := Proto2Domain(DTO(src, &got))
	require.NoError(t, err)
	require.Len(t, got, 2)
	require.True(t, got[0].Spec.Healthy)
	require.False(t, got[1].Spec.Healthy)
}

func Test_UpdHealthStatus_Proto2Domain_Empty(t *testing.T) {
	src := &pb.HostReq_UpdHealthStatus{}

	var got domain.Hosts
	err := Proto2Domain(DTO(src, &got))
	require.NoError(t, err)
	require.Nil(t, got)
}

func Test_UpdHealthStatus_Proto2Domain_SetsOnlyHealthy(t *testing.T) {
	uid := uuid.MustParse("eeee5555-5555-5555-5555-555555555555")

	src := &pb.HostReq_UpdHealthStatus{
		Hosts: []*pb.HostReq_UpdHealthStatus_Host{
			{
				Metadata: &common.MetadataScope{Uid: uid.String(), Name: "h-1", Namespace: "ns"},
				Spec:     &pb.HostReq_UpdHealthStatus_Host_Spec{Healthy: true},
			},
		},
	}

	var got domain.Hosts
	err := Proto2Domain(DTO(src, &got))
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.True(t, got[0].Spec.Healthy)
	require.Empty(t, got[0].Spec.IPs.IPv4.Values())
	require.Empty(t, got[0].Spec.IPs.IPv6.Values())
	require.Empty(t, string(got[0].Spec.DisplayName))
}

func Test_SpecToDomain_DoesNotReadHealthy(t *testing.T) {
	src := &pb.Host_Spec{
		DisplayName: "host-1",
		Healthy:     true,
	}

	var got domain.HostSpec
	err := Proto2Domain(DTO(src, &got))
	require.NoError(t, err)
	require.False(t, got.Healthy, "specToDomain must not read Healthy from proto (output-only field)")
}
