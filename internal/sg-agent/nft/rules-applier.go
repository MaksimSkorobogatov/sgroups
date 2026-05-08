package nft

import (
	"context"
	"time"

	"github.com/PRO-Robotech/sgroups/internal/sg-agent/dns"
	"github.com/PRO-Robotech/sgroups/internal/sg-agent/resources"
	"github.com/PRO-Robotech/sgroups/internal/shared/misc"

	"github.com/H-BF/corlib/logger"
	"github.com/H-BF/corlib/pkg/backoff"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"go.uber.org/multierr"
)

// NewRuleApplierProvider creates provider for resources.RuleApplier
func NewRuleApplierProvider(ctx context.Context, resolver dns.DomainAddressQuerier, opts ...RuleApplierOpt) (ret RuleApplierProvider, err error) {
	defer func() {
		err = errors.WithMessage(err, "nft/NewRuleApplierProvider")
	}()
	obj := new(rulesApplierProviderImpl)
	for _, op := range opts {
		if err = op.apply(&obj.rulesApplierOpts); err != nil {
			return nil, err
		}
	}
	obj.dnsResolver = resolver

	return obj, nil
}

var _ RuleApplier = (*rulesApplierImpl)(nil)

type (
	rulesApplierProviderImpl struct {
		rulesApplierOpts
	}

	rulesApplierImpl struct {
		rulesApplierOpts
	}
)

var _ RuleApplierProvider = (*rulesApplierProviderImpl)(nil)

// NewApplier creates resources.RuleApplier instance
func (prov *rulesApplierProviderImpl) NewApplier(ctx context.Context) (ret RuleApplier, err error) {
	return &rulesApplierImpl{
		rulesApplierOpts: prov.rulesApplierOpts,
	}, nil
}

// ApplyConfig -
func (impl *rulesApplierImpl) ApplyConfig(ctx context.Context, ld resources.LocalData) (applied AppliedRules, err error) {
	const api = "nft/ApplyConfig"

	log := logger.FromContext(ctx).Named("nft").Named("ApplyConfig")
	if len(impl.netNs) > 0 {
		log = log.WithField("net-ns", impl.netNs)
	}
	log.Info("begin")
	defer func() {
		if err == nil {
			log.Info("succeeded")
		} else {
			log.Error(err)
			err = errors.WithMessage(multierr.Combine(
				ErrNfTablesProcessor, err,
			), api)
		}
	}()

	b := newBatch(
		ctx,
		impl.rulesApplierOpts,
		func() (*Tx, error) {
			return NewTx(impl.netNs)
		},
		ld,
	)

	if err = b.execute(ctx); err != nil {
		_ = b.cleanOnFail(ctx)
	} else {
		applied.LocalData = ld
		applied.BaseRules = impl.baseRules
		applied.TargetTable = b.table.Name
		applied.NetNS = impl.netNs
		applied.ID = uuid.NewV4()
	}

	return applied, err
}

// ApplyPatch impl rc.RuleApplier
func (impl *rulesApplierImpl) ApplyPatch(ctx context.Context, rules AppliedRules, patch Patch) (err error) {
	const op = "nft/ApplyPatch"
	log := logger.FromContext(ctx).Named("nft.patch")
	if len(impl.netNs) > 0 {
		log = log.WithField("net-ns", impl.netNs)
	}
	defer func() {
		err = errors.WithMessage(err, op)
		if err == nil {
			log.Infof("%s is applied", patch)
		} else {
			misc.Tern(errors.Is(err, ErrPatchNotApplicable),
				log.Warnf, log.Errorf)("%v", err)
		}
	}()

	for bk := makeBatchBackoff(ctx); ; {
		err = patch.Apply(ctx, &rules)
		if err == nil || errors.Is(err, ErrPatchNotApplicable) {
			break
		}
		pauseDuration := bk.NextBackOff()
		if pauseDuration == backoff.Stop {
			break
		}
		log.Errorf("%s has failed: %v; will retry after %v",
			patch, err, pauseDuration)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(pauseDuration):
		}
	}
	return err
}

// Close -
func (impl *rulesApplierImpl) Close() error {
	return nil
}
