package nft

import (
	"github.com/PRO-Robotech/sgroups/internal/sg-agent/config"
	"github.com/PRO-Robotech/sgroups/internal/sg-agent/dns"
	"github.com/PRO-Robotech/sgroups/internal/sg-agent/resources"

	"github.com/pkg/errors"
)

// RuleApplierOpt option uses by NewRuleApplier
type RuleApplierOpt interface {
	apply(*rulesApplierOpts) error
}

// WithNetNS is RuleApplierOpt
func WithNetNS[T ~string](ns T) RuleApplierOpt {
	return ruleApplierOptF(func(rai *rulesApplierOpts) error {
		rai.netNs = string(ns)
		return nil
	})
}

// UseAcceptDefPolicy -
func UseAcceptDefPolicy() RuleApplierOpt {
	return ruleApplierOptF(func(rai *rulesApplierOpts) error {
		rai.defaultPolicyAccept = true
		return nil
	})
}

// WithBaseRules is RuleApplierOpt
func WithBaseRules(r resources.BaseRuleList) RuleApplierOpt {
	return ruleApplierOptF(func(rai *rulesApplierOpts) error {
		err := errors.WithMessage(r.Validate(), "invalid base rules are passed")
		if err == nil {
			rai.baseRules = r
		}
		return err
	})
}

// WithFqdnStrategy is RuleApplierOpt
func WithFqdnStrategy(strategy config.FqdnRulesStrategy) RuleApplierOpt {
	return ruleApplierOptF(func(rai *rulesApplierOpts) error {
		rai.fqdnStrategy = strategy
		return nil
	})
}

// UseDryRun -
func UseDryRun() RuleApplierOpt {
	return ruleApplierOptF(func(rai *rulesApplierOpts) error {
		rai.dryRun = true
		return nil
	})
}

type (
	rulesApplierOpts struct {
		netNs               string
		baseRules           resources.BaseRuleList
		defaultPolicyAccept bool
		dnsResolver         dns.DomainAddressQuerier
		fqdnStrategy        config.FqdnRulesStrategy
		dryRun              bool
	}

	ruleApplierOptF func(*rulesApplierOpts) error
)

func (optf ruleApplierOptF) apply(i *rulesApplierOpts) error {
	return optf(i)
}
