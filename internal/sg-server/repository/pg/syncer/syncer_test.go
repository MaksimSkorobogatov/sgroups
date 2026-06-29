package syncer

import (
	"context"
	"net"
	"testing"
	"time"

	pg "github.com/PRO-Robotech/sgroups/internal/sg-server/repository/pg/domain"
	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"

	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type namespaceSyncerTestSuite struct {
	suite.Suite
}

func Test_NamespaceSyncer(t *testing.T) {
	suite.Run(t, new(namespaceSyncerTestSuite))
}

func (sui *namespaceSyncerTestSuite) Test_Upsert() {
	ctx := context.Background()

	mock, err := pgxmock.NewConn()
	sui.Require().NoError(err)
	defer func() {
		_ = mock.Close(ctx)
	}()

	uid := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	ts := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)

	input := domain.Namespace{
		Metadata: domain.NsMetadata{
			ID: domain.ClusterScopeMetadataIdentity{
				UID:  uid,
				Name: domain.ResourceName("ns-1"),
			},
			Labels:            map[string]string{"env": "dev"},
			Annotations:       map[string]string{"a": "b"},
			CreationTimestamp: ts,
			ResourceVersion:   "10",
		},
		Spec: domain.CommonSpec{
			DisplayName: domain.DisplayName("Namespace 1"),
			Comment:     "c1",
			Description: "d1",
		},
	}

	rows := mock.NewRows([]string{
		"uid",
		"name",
		"display_name",
		"comment",
		"description",
		"labels",
		"annotations",
		"creation_timestamp",
		"resource_version",
	}).AddRow(
		uid,
		"ns-1",
		"Namespace 1",
		"c1",
		"d1",
		map[string]string{"env": "dev"},
		map[string]string{"a": "b"},
		ts,
		"10",
	)

	pgObj, err := NamespaceSyncer.toPgConv(input)
	sui.Require().NoError(err)
	args := NamespaceSyncer.syncArgs(pgObj)

	expected, err := NamespaceSyncer.fromPgConv(pg.Namespace{
		NsMetadata: pg.NsMetadata{
			NsPK: pg.NsPK{
				UID:  uid,
				Name: "ns-1",
			},
			CommonMetadata: pg.CommonMetadata{
				DisplayName: "Namespace 1",
				Comment:     "c1",
				Description: "d1",
				Labels:      map[string]string{"env": "dev"},
				Annotations: map[string]string{"a": "b"},
			},
		},
		CreationTimestamp: ts,
		ResourceVersion:   "10",
	})
	sui.Require().NoError(err)

	mock.ExpectQuery(`^select \* from sgroups\.sync_namespaces\('ups',\s*row\(\$1,\$2,\$3,\$4,\$5,\$6,\$7\)\)$`).
		WithArgs(args...).
		WillReturnRows(rows)

	got, err := NamespaceSyncer.Sync(ctx, mock, Upsert, []domain.Namespace{input})
	sui.Require().NoError(err)
	sui.Require().Len(got, 1)
	sui.Require().Equal(expected, got[0])
	sui.Require().NoError(mock.ExpectationsWereMet())
}

func (sui *namespaceSyncerTestSuite) Test_Delete() {
	ctx := context.Background()

	mock, err := pgxmock.NewConn()
	sui.Require().NoError(err)
	defer func() {
		_ = mock.Close(ctx)
	}()

	uid := uuid.MustParse("22222222-2222-2222-2222-222222222222")

	input := domain.Namespace{
		Metadata: domain.NsMetadata{
			ID: domain.ClusterScopeMetadataIdentity{
				UID:  uid,
				Name: domain.ResourceName("ns-del"),
			},
			Labels:      map[string]string{"env": "prod"},
			Annotations: map[string]string{"x": "y"},
		},
		Spec: domain.CommonSpec{
			DisplayName: domain.DisplayName("Namespace Del"),
			Comment:     "c-del",
			Description: "d-del",
		},
	}

	pgObj, err := NamespaceSyncer.toPgConv(input)
	sui.Require().NoError(err)
	args := NamespaceSyncer.syncArgs(pgObj)

	mock.ExpectExec(`^select \* from sgroups\.sync_namespaces\('del',\s*row\(\$1,\$2,\$3,\$4,\$5,\$6,\$7\)\)$`).
		WithArgs(args...).
		WillReturnResult(pgxmock.NewResult("SELECT", 0))

	got, err := NamespaceSyncer.Sync(ctx, mock, Delete, []domain.Namespace{input})
	sui.Require().NoError(err)
	sui.Require().Len(got, 0)
	sui.Require().NoError(mock.ExpectationsWereMet())
}

func (sui *namespaceSyncerTestSuite) Test_DisabledOp() {
	ctx := context.Background()

	mock, err := pgxmock.NewConn()
	sui.Require().NoError(err)
	defer func() {
		_ = mock.Close(ctx)
	}()

	s := NamespaceSyncer
	s.disabledOps = []SyncOp{Upsert}

	input := domain.Namespace{
		Metadata: domain.NsMetadata{
			ID: domain.ClusterScopeMetadataIdentity{
				UID:  uuid.MustParse("33333333-3333-3333-3333-333333333333"),
				Name: domain.ResourceName("ns-1"),
			},
		},
	}

	_, err = s.Sync(ctx, mock, Upsert, []domain.Namespace{input})
	sui.Require().Error(err)

	// No expectations should be pending.
	sui.Require().NoError(mock.ExpectationsWereMet())
}

func (sui *namespaceSyncerTestSuite) Test_ConvertsPgDomain() {
	// Minimal smoke-test that pg->domain conversion is wired (uses dto.Pg2Domain under the hood).
	ctx := context.Background()

	mock, err := pgxmock.NewConn()
	sui.Require().NoError(err)
	defer func() {
		_ = mock.Close(ctx)
	}()

	uid := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	ts := time.Date(2025, 6, 7, 8, 9, 10, 0, time.UTC)

	input := domain.Namespace{
		Metadata: domain.NsMetadata{
			ID:          domain.ClusterScopeMetadataIdentity{UID: uid, Name: domain.ResourceName("ns-pg")},
			Labels:      map[string]string{"k": "v"},
			Annotations: map[string]string{},
		},
		Spec: domain.CommonSpec{DisplayName: domain.DisplayName("ns-pg"), Comment: "", Description: ""},
	}

	pgRow := pg.Namespace{
		NsMetadata: pg.NsMetadata{
			NsPK: pg.NsPK{
				UID:  uid,
				Name: "ns-pg",
			},
			CommonMetadata: pg.CommonMetadata{
				DisplayName: "ns-pg",
				Comment:     "",
				Description: "",
				Labels:      map[string]string{"k": "v"},
				Annotations: map[string]string{},
			},
		},
		CreationTimestamp: ts,
		ResourceVersion:   "99",
	}

	rows := mock.NewRows([]string{
		"uid",
		"name",
		"display_name",
		"comment",
		"description",
		"labels",
		"annotations",
		"creation_timestamp",
		"resource_version",
	}).AddRow(
		pgRow.UID,
		pgRow.Name,
		pgRow.DisplayName,
		pgRow.Comment,
		pgRow.Description,
		pgRow.Labels,
		pgRow.Annotations,
		pgRow.CreationTimestamp,
		pgRow.ResourceVersion,
	)

	pgObj, err := NamespaceSyncer.toPgConv(input)
	sui.Require().NoError(err)
	args := NamespaceSyncer.syncArgs(pgObj)

	mock.ExpectQuery(`^select \* from sgroups\.sync_namespaces\('ups',\s*row\(\$1,\$2,\$3,\$4,\$5,\$6,\$7\)\)$`).
		WithArgs(args...).
		WillReturnRows(rows)

	got, err := NamespaceSyncer.Sync(ctx, mock, Upsert, []domain.Namespace{input})
	sui.Require().NoError(err)
	sui.Require().Len(got, 1)
	sui.Require().Equal("99", got[0].Metadata.ResourceVersion)
	sui.Require().Equal(ts, got[0].Metadata.CreationTimestamp)
	sui.Require().NoError(mock.ExpectationsWereMet())
}

func Test_HostHealthSyncer_Upsert(t *testing.T) {
	ctx := context.Background()

	mock, err := pgxmock.NewConn()
	require.NoError(t, err)
	defer func() { _ = mock.Close(ctx) }()

	uid := uuid.MustParse("aaaa1111-1111-1111-1111-111111111111")
	ts := time.Date(2026, 6, 29, 12, 0, 0, 0, time.UTC)

	input := domain.Host{
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
	}

	pgObj, err := HostHealthSyncer.toPgConv(input)
	require.NoError(t, err)
	args := HostHealthSyncer.syncArgs(pgObj)

	require.Len(t, args, 4)
	require.Equal(t, uid, args[0])
	require.Equal(t, "h-health", args[1])
	require.Equal(t, "ns-1", args[2])
	require.Equal(t, true, args[3])

	rows := mock.NewRows([]string{
		"uid", "name", "namespace",
		"display_name", "comment", "description",
		"labels", "annotations",
		"ips", "meta_info", "refs",
		"creation_timestamp", "resource_version", "endpoints", "healthy",
	}).AddRow(
		uid, "h-health", "ns-1",
		"", "", "",
		map[string]string{}, map[string]string{},
		nil, nil, nil,
		ts, "11", nil, true,
	)

	mock.ExpectQuery(`^select \* from sgroups\.sync_host_health_status\('ups',\s*row\(\$1,\$2,\$3,\$4\)\)$`).
		WithArgs(args...).
		WillReturnRows(rows)

	got, err := HostHealthSyncer.Sync(ctx, mock, Upsert, []domain.Host{input})
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Equal(t, uid, got[0].Metadata.ID.UID)
	require.Equal(t, domain.ResourceName("h-health"), got[0].Metadata.ID.Name)
	require.True(t, got[0].Spec.Healthy)
	require.NoError(t, mock.ExpectationsWereMet())
}

func Test_HostHealthSyncer_SyncArgs_HealthyFalse(t *testing.T) {
	input := domain.Host{
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
	}

	pgObj, err := HostHealthSyncer.toPgConv(input)
	require.NoError(t, err)
	args := HostHealthSyncer.syncArgs(pgObj)

	require.Len(t, args, 4)
	require.Equal(t, false, args[3])
}

func Test_NetworkSyncer_DelArgs_PassesOnlyIdentity(t *testing.T) {
	input := domain.Network{
		Metadata: domain.ResMetadata{
			ID: domain.NamespacedMetadataIdentity{
				ClusterScopeMetadataIdentity: domain.ClusterScopeMetadataIdentity{
					UID:  uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
					Name: domain.ResourceName("nw-del"),
				},
				Namespace: domain.ResourceNamespace("ns-1"),
			},
		},
		Spec: domain.NetworkSpec{},
	}

	pgObj, err := NetworkSyncer.toPgConv(input)
	require.NoError(t, err)
	args := NetworkSyncer.delArgs(pgObj)
	require.Len(t, args, 9)
	require.Equal(t, pgObj.UID, args[0])
	require.Equal(t, "nw-del", args[1])
	require.Equal(t, "ns-1", args[2])
	for i := 3; i < 9; i++ {
		require.Nil(t, args[i])
	}
}

func Test_NetworkSyncer_SyncArgs_UsesCIDRWhenPresent(t *testing.T) {
	input := domain.Network{
		Metadata: domain.ResMetadata{
			ID: domain.NamespacedMetadataIdentity{
				ClusterScopeMetadataIdentity: domain.ClusterScopeMetadataIdentity{
					UID:  uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"),
					Name: domain.ResourceName("nw-ups"),
				},
				Namespace: domain.ResourceNamespace("ns-1"),
			},
		},
		Spec: domain.NetworkSpec{
			CIDR: func() domain.IPNet {
				_, ipnet, _ := net.ParseCIDR("10.0.0.0/24")
				return domain.IPNet{IPNet: *ipnet}
			}(),
		},
	}

	pgObj, err := NetworkSyncer.toPgConv(input)
	require.NoError(t, err)
	args := NetworkSyncer.syncArgs(pgObj)
	require.Len(t, args, 9)
	require.NotEqual(t, net.IPNet{}, args[8])
}
