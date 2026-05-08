package main

import (
	"context"

	"github.com/PRO-Robotech/sgroups/internal/sg-agent/app"
	"github.com/PRO-Robotech/sgroups/internal/sg-agent/config"
	"github.com/PRO-Robotech/sgroups/internal/sg-agent/job"
	"github.com/PRO-Robotech/sgroups/internal/sg-agent/nft"
	"github.com/PRO-Robotech/sgroups/internal/sg-agent/resources"

	conf "github.com/H-BF/corlib/pkg/plain-config"
	"github.com/pkg/errors"
)

func makeRuleApplierProvider(ctx context.Context) (applier nft.RuleApplierProvider, err error) {
	var (
		opts         []nft.RuleApplierOpt
		fqdnStrategy config.FqdnRulesStrategy
	)
	if fqdnStrategy, err = config.FqdnStrategy.Value(ctx); err != nil {
		return nil, err
	}
	opts = append(opts, nft.WithFqdnStrategy(fqdnStrategy))

	err = app.IfBaseRulesFromConfig(ctx, func(br resources.BaseRuleList) error {
		if e := br.Validate(); e != nil {
			return e
		}
		opts = append(opts, nft.WithBaseRules(br))
		return nil
	})
	if err != nil {
		return nil, errors.WithMessage(err, "load base rules")
	}

	if v, e := config.NetNS.Value(ctx); e == nil {
		opts = append(opts, nft.WithNetNS(v))
	} else if !errors.Is(e, config.ErrNotFound) {
		return nil, e
	}

	if v, e := config.DefPolicyAccept.Value(ctx); e == nil && v {
		opts = append(opts, nft.UseAcceptDefPolicy())
	} else if e != nil && !errors.Is(e, conf.ErrNotFound) {
		return nil, e
	}

	if v, e := config.DryRun.Value(ctx); e == nil && v {
		opts = append(opts, nft.UseDryRun())
	} else if e != nil && !errors.Is(e, conf.ErrNotFound) {
		return nil, e
	}

	return nft.NewRuleApplierProvider(ctx, app.GetDnsResolver(), opts...)
}

func makeRuleApplier(ctx context.Context, rulesApplier nft.RuleApplier) (job.Task, error) {
	var (
		opts []job.Option
	)
	if v, e := config.NetNS.Value(ctx); e == nil {
		opts = append(opts, job.WithNetNS(v))
	} else if !errors.Is(e, conf.ErrNotFound) {
		return nil, e
	}

	if v, e := config.DefPolicyAccept.Value(ctx); e == nil && v {
		opts = append(opts, job.WithDefPolicyAccept(v))
	} else if e != nil && !errors.Is(e, conf.ErrNotFound) {
		return nil, e
	}
	return job.NewNftApplier(rulesApplier, app.SGClientProviderInstance, opts...), nil
}
