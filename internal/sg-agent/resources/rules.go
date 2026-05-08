package resources

import (
	"context"

	dto "github.com/PRO-Robotech/sgroups/internal/sg-agent/dto/rules"
	sg "github.com/PRO-Robotech/sgroups/internal/sg-agent/transport/sgroups/grpc"
	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"

	"github.com/H-BF/corlib/pkg/dict"
	sgv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

type (
	// RuleID -
	RuleID = domain.ResourceIdentifier

	// Rules - list of rules with metadata
	Rules struct {
		dict.HDict[domain.ResourceType, *rules]
	}

	rules struct {
		dict.HDict[RuleID, domain.Rule]
	}
)

// GetAgRefs returns address group references of the services
func (r *Rules) GetAgRefs() (res []domain.ResourceIdentifier) {
	for _, rule := range r.Iterate {
		for _, rl := range rule.Iterate {
			res = append(res, rl.GetAddressGroupRefs()...)
		}
	}
	return res
}

// GetSvcRefs returns service references of the services
func (r *Rules) GetSvcRefs() (res []domain.ResourceIdentifier) {
	for _, rule := range r.Iterate {
		for _, rl := range rule.Iterate {
			res = append(res, rl.GetServiceRefs()...)
		}
	}
	return res
}

// IsEq -
func (r *Rules) IsEq(other Rules) bool {
	return r.Eq(&other, func(vL, vR *rules) bool {
		return vL.IsEq(vR)
	})
}

// LoadFromRefs loads rules from SGroups server by rules references
func (r *Rules) LoadFromRefs(ctx context.Context, sgClients sg.Clients, refs []domain.ResourceIdentifier) error {
	return r.load(ctx, sgClients, makeRuleFieldSelectorsByNames(refs))
}

func (r *Rules) load(ctx context.Context, sgClients sg.Clients, selectors []*sgv1.RuleReq_Selectors) error {
	client, err := sgClients.Rules()
	if err != nil {
		return err
	}
	return loader(ctx, selectors,
		func(ctx context.Context, selectors []*sgv1.RuleReq_Selectors) ([]*sgv1.Rule, error) {
			resp, e := client.List(ctx, &sgv1.RuleReq_List{
				Selectors: selectors,
			})
			return resp.GetRules(), errors.WithMessage(e, "list rules")
		},
		func(pb *sgv1.Rule) (domain.Rule, error) {
			var rl domain.Rule
			e := dto.Proto2Domain(dto.DTO(pb, &rl))
			return rl, e
		},
		func(k domain.ResourceIdentifier, v domain.Rule) {
			rt := v.Type()
			item := r.At(rt)
			if item == nil {
				item = new(rules)
				_ = r.Insert(rt, item)
			}
			item.Put(k, v)
		},
	)
}

// IsEq -
func (r *rules) IsEq(other *rules) bool {
	return r.Eq(other, func(lRl, rRl domain.Rule) bool {
		return lRl.IsEq(rRl)
	})
}

// In -
func (r *rules) In(to domain.ResourceIdentifier) (ret []domain.Rule, err error) { //nolint:dupl
	var remote domain.EpRemote
	for _, v := range r.Iterate {
		if remote, err = domain.RemoteFromSpec(v.Spec.Remote); err != nil {
			return nil, err
		}
		if remote.ResourceIdentifier == to {
			ret = append(ret, v)
		}
	}

	return ret, nil
}

// Out -
func (r *rules) Out(from domain.ResourceIdentifier) (ret []domain.Rule, err error) { //nolint:dupl
	var local domain.EpRemote
	for _, v := range r.Iterate {
		if local, err = domain.RemoteFromSpec(v.Spec.Local); err != nil {
			return nil, err
		}
		if local.ResourceIdentifier == from {
			ret = append(ret, v)
		}
	}

	return ret, nil
}

// BothTraffic -
func (r *rules) BothTraffic() *rules {
	if r == nil {
		return nil
	}
	return r.Traffic(domain.BOTH)
}

// IngressTraffic -
func (r *rules) IngressTraffic() *rules {
	if r == nil {
		return nil
	}
	return r.Traffic(domain.INGRESS)
}

// EgressTraffic -
func (r *rules) EgressTraffic() *rules {
	if r == nil {
		return nil
	}
	return r.Traffic(domain.EGRESS)
}

// Traffic returns rules with specified traffic
func (r *rules) Traffic(dir domain.Traffic) *rules {
	if r == nil {
		return nil
	}
	var ret *rules
	for _, rl := range r.Iterate {
		if rl.Spec.Traffic == dir {
			if ret == nil {
				ret = new(rules)
			}
			ret.Put(rl.ResourceID(), rl)
		}
	}
	return ret
}

// Len returns number of rules in the list
func (r *rules) Len() int {
	if r == nil {
		return 0
	}
	return r.HDict.Len()
}

// Values returns list of rules
func (r *rules) Values() []domain.Rule {
	if r == nil {
		return nil
	}
	return lo.Map(r.HDict.Items(), func(item dict.KV[RuleID, domain.Rule], index int) domain.Rule {
		return item.V
	})
}
