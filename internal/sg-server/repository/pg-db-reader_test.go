package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	pg "github.com/PRO-Robotech/sgroups/internal/sg-server/repository/pg/domain"
	"github.com/PRO-Robotech/sgroups/internal/sg-server/repository/scopes"
	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
)

type mocConn struct {
	pgxmock.PgxConnIface
	channel string
	notifyQ []any
	cancel  context.CancelFunc
}

func (m *mocConn) WaitForNotification(_ context.Context) (*pgconn.Notification, error) {
	if len(m.notifyQ) == 0 {
		if m.cancel != nil {
			m.cancel()
		}
		return nil, nil
	}

	v := m.notifyQ[0]
	m.notifyQ = m.notifyQ[1:]
	if len(m.notifyQ) == 0 && m.cancel != nil {
		m.cancel()
	}

	j, err := json.Marshal(v)
	return &pgconn.Notification{
		PID:     uint32(os.Getpid()),
		Channel: m.channel,
		Payload: string(j),
	}, err
}

func Test_ListNamespaces(t *testing.T) {
	ctx := context.Background()

	mock, err := pgxmock.NewConn()
	require.NoError(t, err)
	defer func() {
		_ = mock.Close(ctx)
	}()

	uid1 := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	uid2 := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	ts1 := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
	ts2 := time.Date(2025, 2, 3, 4, 5, 6, 0, time.UTC)

	rows := mock.NewRows(pg.Namespace{}.Columns()).
		AddRow(uid1, "ns-1", "Namespace 1", "c1", "d1",
			map[string]string{"env": "dev"}, map[string]string{"a": "b"}, ts1, "10").
		AddRow(uid2, "ns-2", "Namespace 2", "c2", "d2",
			map[string]string{"env": "prod"}, map[string]string{"x": "y"}, ts2, "11")

	expQuery := fmt.Sprintf("SELECT %s FROM sgroups.list_namespaces($1)", strings.Join(pg.Namespace{}.Columns(), ", "))
	mock.ExpectQuery("^"+regexp.QuoteMeta(expQuery)+"$").
		WithArgs(pgx.QueryExecModeDescribeExec, pgxmock.AnyArg()).
		WillReturnRows(rows)

	rd := &pgDbReader{
		doIt: func(_ context.Context, f func(pgConn) error) error {
			return f(&mocConn{PgxConnIface: mock})
		},
	}

	got, err := rd.ListNamespaces(ctx, nil)
	require.NoError(t, err)

	expected := []domain.Namespace{
		{
			Metadata: domain.NsMetadata{
				ID: domain.ClusterScopeMetadataIdentity{
					UID:  uid1,
					Name: domain.ResourceName("ns-1"),
				},
				Labels:            map[string]string{"env": "dev"},
				Annotations:       map[string]string{"a": "b"},
				CreationTimestamp: ts1,
				ResourceVersion:   "10",
			},
			Spec: domain.CommonSpec{
				DisplayName: domain.DisplayName("Namespace 1"),
				Comment:     "c1",
				Description: "d1",
			},
		},
		{
			Metadata: domain.NsMetadata{
				ID: domain.ClusterScopeMetadataIdentity{
					UID:  uid2,
					Name: domain.ResourceName("ns-2"),
				},
				Labels:            map[string]string{"env": "prod"},
				Annotations:       map[string]string{"x": "y"},
				CreationTimestamp: ts2,
				ResourceVersion:   "11",
			},
			Spec: domain.CommonSpec{
				DisplayName: domain.DisplayName("Namespace 2"),
				Comment:     "c2",
				Description: "d2",
			},
		},
	}

	require.Equal(t, expected, got)
	require.NoError(t, mock.ExpectationsWereMet())
}

func Test_WatchNamespaces(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mock, err := pgxmock.NewConn()
	require.NoError(t, err)
	defer func() {
		_ = mock.Close(ctx)
	}()

	// scope = AND(resourceVersion=10, namespace selector {name=ns-1, labels={env:dev}})
	sel := domain.ResSelector{
		FieldSelector: domain.ResFieldSelector{
			ResourceIdentifier: domain.ResourceIdentifier{Name: domain.ResourceName("ns-1")},
		},
		LabelSelector: map[string]string{"env": "dev"},
	}
	scope := scopes.ScopedAnd{
		L: scopes.ScopeByResourceVersion{RV: "10"},
		R: scopes.ByResSelectors(sel),
	}

	makeEvt := func(rv string, name string, labels map[string]string) pg.ResourceEvent {
		ns := pg.Namespace{
			NsMetadata: pg.NsMetadata{
				NsPK: pg.NsPK{
					UID:  uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
					Name: name,
				},
				CommonMetadata: pg.CommonMetadata{
					DisplayName: name,

					Labels:      labels,
					Annotations: map[string]string{},
				},
			},
			CreationTimestamp: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			ResourceVersion:   rv,
		}
		obj, _ := json.Marshal(ns)
		return pg.ResourceEvent{
			TS:              time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			ResourceVersion: rv,
			ResourceType:    pg.ResourceType(domain.NamespaceResource.String()),
			EventType:       domain.ResourceModified.String(),
			Object:          obj,
		}
	}

	evList1 := makeEvt("11", "ns-1", map[string]string{"env": "dev"})
	evList2 := makeEvt("12", "ns-2", map[string]string{"env": "prod"})

	// initial outbox list: two events, but scope should pass только ns-1
	rows := mock.NewRows([]string{"ts", "resource_version", "resource_type", "event_type", "object"}).
		AddRow(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), "11", domain.NamespaceResource.String(), domain.ResourceModified.String(), evList1.Object).
		AddRow(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), "12", domain.NamespaceResource.String(), domain.ResourceAdded.String(), evList2.Object)

	mock.ExpectQuery(`^select \* from sgroups\.list_outbox_resource_events\(\$1, \$2, \$3\)$`).
		WithArgs(pgx.QueryExecModeDescribeExec, domain.NamespaceResource.String(), 10, 1000000).
		WillReturnRows(rows)

	mock.ExpectExec(`^listen ("resource_` + domain.NamespaceResource.String() + `"|resource_` + domain.NamespaceResource.String() + `)$`).WillReturnResult(pgxmock.NewResult("LISTEN", 1))
	mock.ExpectExec(`^unlisten ("resource_` + domain.NamespaceResource.String() + `"|resource_` + domain.NamespaceResource.String() + `)$`).WillReturnResult(pgxmock.NewResult("UNLISTEN", 1))

	conn := &mocConn{
		PgxConnIface: mock,
		channel:      "resource_" + domain.NamespaceResource.String(),
		cancel:       cancel,
		notifyQ: []any{
			makeEvt("13", "ns-2", map[string]string{"env": "prod"}), // should be filtered out
			makeEvt("14", "ns-1", map[string]string{"env": "dev"}),  // should pass
		},
	}

	rd := &pgDbReader{
		doIt: func(_ context.Context, f func(pgConn) error) error {
			return f(conn)
		},
	}

	var got []domain.NamespaceEvent
	err = rd.WatchNamespaces(ctx, scope, func(e domain.NamespaceEvent) error {
		got = append(got, e)
		return nil
	})
	require.NoError(t, err)

	require.Len(t, got, 2)
	for _, e := range got {
		require.Equal(t, domain.ResourceName("ns-1"), e.Object.Metadata.ID.Name)
		require.Equal(t, map[string]string{"env": "dev"}, e.Object.Metadata.Labels)
	}

	require.NoError(t, mock.ExpectationsWereMet())
}
