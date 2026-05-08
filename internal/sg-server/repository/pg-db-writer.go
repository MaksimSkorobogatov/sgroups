package repository

import (
	"context"
	"fmt"

	"github.com/PRO-Robotech/sgroups/internal/sg-server/repository/pg/syncer"
	"github.com/PRO-Robotech/sgroups/internal/sg-server/repository/scopes"
	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"
	"github.com/PRO-Robotech/sgroups/internal/shared/misc"

	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
)

type pgDbWriter struct {
	notImplWriterFace //nolint:unused

	tx     func() (pgx.Tx, error)
	commit func() error
	abort  func()
}

var _ Writer = (*pgDbWriter)(nil)

// SyncNamespace implements Writer
func (wr *pgDbWriter) SyncNamespace(ctx context.Context, ns []domain.Namespace, op SyncOp) (ret []domain.Namespace, err error) {
	err = wr.withTx("SyncNamespace", func(tx pgx.Tx) error {
		ret, err = syncer.NamespaceSyncer.Sync(ctx, tx, toSyncerOP(op), ns)
		return err
	})
	return ret, err
}

// SyncAddressGroup implements Writer
func (wr *pgDbWriter) SyncAddressGroup(ctx context.Context, ag []domain.AddressGroup, op SyncOp) (ret []domain.AddressGroup, err error) {
	err = wr.withTx("SyncAddressGroup", func(tx pgx.Tx) error {
		ret, err = syncer.AddressGroupSyncer.Sync(ctx, tx, toSyncerOP(op), ag)
		return err
	})
	return ret, err
}

// SyncNetwork implements Writer
func (wr *pgDbWriter) SyncNetwork(ctx context.Context, nw []domain.Network, op SyncOp) (ret []domain.Network, err error) {
	err = wr.withTx("SyncNetwork", func(tx pgx.Tx) error {
		ret, err = syncer.NetworkSyncer.Sync(ctx, tx, toSyncerOP(op), nw)
		return err
	})
	return ret, err
}

// SyncHost implements Writer
func (wr *pgDbWriter) SyncHost(ctx context.Context, scope Scope, op SyncOp) (ret []domain.Host, err error) {
	var (
		api        = "SyncHost"
		hosts      []domain.Host
		hostSyncer syncer.Syncer[domain.Host, domain.Host]
	)
	switch sc := scope.(type) {
	case scopes.ScopeByHostIPs:
		api = "SyncHostIPs"
		hosts = sc.Hosts
		hostSyncer = syncer.HostIPsSyncer
	case scopes.ScopeByHostInfo:
		api = "SyncHostInfo"
		hosts = sc.Hosts
		hostSyncer = syncer.HostInfoSyncer
	case scopes.ScopeByHosts:
		api = "SyncHost"
		hosts = sc.Hosts
		hostSyncer = syncer.HostSyncer
	default:
		return nil, errors.Errorf("%s: unsupported scope '%T'", api, sc)
	}
	err = wr.withTx(api, func(tx pgx.Tx) error {
		ret, err = hostSyncer.Sync(ctx, tx, toSyncerOP(op), hosts)
		return err
	})
	return ret, err
}

// SyncHostBinding implements Writer
func (wr *pgDbWriter) SyncHostBinding(ctx context.Context, hb []domain.HostBinding, op SyncOp) (ret []domain.HostBinding, err error) {
	err = wr.withTx("SyncHostBinding", func(tx pgx.Tx) error {
		ret, err = syncer.HostBindingSyncer.Sync(ctx, tx, toSyncerOP(op), hb)
		return err
	})
	return ret, err
}

// SyncNetworkBinding implements Writer
func (wr *pgDbWriter) SyncNetworkBinding(ctx context.Context, nb []domain.NetworkBinding, op SyncOp) (ret []domain.NetworkBinding, err error) {
	err = wr.withTx("SyncNetworkBinding", func(tx pgx.Tx) error {
		ret, err = syncer.NetworkBindingSyncer.Sync(ctx, tx, toSyncerOP(op), nb)
		return err
	})
	return ret, err
}

// SyncRules implements Writer
func (wr *pgDbWriter) SyncRules(ctx context.Context, rules []domain.Rule, op SyncOp) (ret []domain.Rule, err error) {
	err = wr.withTx("SyncRules", func(tx pgx.Tx) error {
		if toSyncerOP(op) == syncer.Delete {
			_, err = syncer.RuleSyncer.Sync(ctx, tx, toSyncerOP(op), rules)
			return err
		}
		for _, r := range rules {
			switch r.Type() {
			case domain.Ag2AgRule:
				ret, err = syncer.Ag2AgSyncer.Sync(ctx, tx, toSyncerOP(op), misc.Sli(r))
			case domain.Ag2AgIcmpRule:
				ret, err = syncer.Ag2AgIcmpSyncer.Sync(ctx, tx, toSyncerOP(op), misc.Sli(r))
			case domain.Ag2IcmpRule:
				ret, err = syncer.Ag2IcmpSyncer.Sync(ctx, tx, toSyncerOP(op), misc.Sli(r))
			case domain.Ag2CidrRule:
				ret, err = syncer.Ag2CidrSyncer.Sync(ctx, tx, toSyncerOP(op), misc.Sli(r))
			case domain.Ag2CidrIcmpRule:
				ret, err = syncer.Ag2CidrIcmpSyncer.Sync(ctx, tx, toSyncerOP(op), misc.Sli(r))
			case domain.Ag2FqdnRule:
				ret, err = syncer.Ag2FqdnSyncer.Sync(ctx, tx, toSyncerOP(op), misc.Sli(r))
			case domain.Svc2SvcRule:
				ret, err = syncer.Svc2SvcSyncer.Sync(ctx, tx, toSyncerOP(op), misc.Sli(r))
			case domain.Svc2FqdnRule:
				ret, err = syncer.Svc2FqdnSyncer.Sync(ctx, tx, toSyncerOP(op), misc.Sli(r))
			case domain.Svc2CidrRule:
				ret, err = syncer.Svc2CidrSyncer.Sync(ctx, tx, toSyncerOP(op), misc.Sli(r))
			case domain.Svc2CidrIcmpRule:
				ret, err = syncer.Svc2CidrIcmpSyncer.Sync(ctx, tx, toSyncerOP(op), misc.Sli(r))
			case domain.Ag2SvcRule:
				ret, err = syncer.Ag2SvcSyncer.Sync(ctx, tx, toSyncerOP(op), misc.Sli(r))
			case domain.Svc2AgRule:
				ret, err = syncer.Svc2AgSyncer.Sync(ctx, tx, toSyncerOP(op), misc.Sli(r))
			case domain.Svc2AgIcmpRule:
				ret, err = syncer.Svc2AgIcmpSyncer.Sync(ctx, tx, toSyncerOP(op), misc.Sli(r))
			case domain.Ag2SvcIcmpRule:
				ret, err = syncer.Ag2SvcIcmpSyncer.Sync(ctx, tx, toSyncerOP(op), misc.Sli(r))
			default:
				err = errors.Errorf("unsupported rule type '%s'", r.Type())
			}
			if err != nil {
				break
			}
		}
		return err
	})

	return ret, err
}

// SyncService implements Writer
func (wr *pgDbWriter) SyncService(ctx context.Context, svc []domain.Service, op SyncOp) (ret []domain.Service, err error) {
	err = wr.withTx("SyncService", func(tx pgx.Tx) error {
		ret, err = syncer.ServiceSyncer.Sync(ctx, tx, toSyncerOP(op), svc)
		return err
	})
	return ret, err
}

// SyncServiceBinding implements Writer
func (wr *pgDbWriter) SyncServiceBinding(ctx context.Context, sb []domain.ServiceBinding, op SyncOp) (ret []domain.ServiceBinding, err error) {
	err = wr.withTx("SyncServiceBinding", func(tx pgx.Tx) error {
		ret, err = syncer.ServiceBindingSyncer.Sync(ctx, tx, toSyncerOP(op), sb)
		return err
	})
	return ret, err
}

// Commit implements Writer
func (wr *pgDbWriter) Commit() error {
	return wr.commit()
}

// Abort implements Writer
func (wr *pgDbWriter) Abort() {
	wr.abort()
}

func (wr *pgDbWriter) withTx(apiName string, action func(pgx.Tx) error) (err error) {
	if apiName != "" {
		defer func() {
			err = correctPGError(errors.WithMessage(err, apiName))
		}()
	}
	var tx pgx.Tx
	if tx, err = wr.tx(); err != nil {
		return err
	}
	return action(tx)
}

func toSyncerOP(op SyncOp) syncer.SyncOp {
	switch v := op.(type) {
	case NoSyncOp:
		return syncer.NoSyncOp
	case Upsert:
		return syncer.Upsert
	case Delete:
		return syncer.Delete
	default:
		panic(fmt.Errorf("unsupported sync-op '%T'", v))
	}
}
