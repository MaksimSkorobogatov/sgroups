package resources

import (
	"context"

	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"

	"github.com/PRO-Robotech/sgroups-proto/pkg/api/common"
	sgv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
	"github.com/samber/lo"
)

func makeRuleFieldSelectorsByNames(ids []domain.ResourceIdentifier) []*sgv1.RuleReq_Selectors {
	sel := make([]*sgv1.RuleReq_Selectors, 0, len(ids))
	for _, id := range lo.Uniq(ids) {
		sel = append(sel, &sgv1.RuleReq_Selectors{
			FieldSelector: &sgv1.RuleReq_Selectors_FieldSelector{
				Name:      id.Name.String(),
				Namespace: id.Namespace.String(),
			},
		})
	}
	return sel
}

func makeFieldSelectorsByNames(ids []domain.ResourceIdentifier) []*common.ResSelector {
	sel := make([]*common.ResSelector, 0, len(ids))
	for _, id := range lo.Uniq(ids) {
		sel = append(sel, &common.ResSelector{
			FieldSelector: &common.FieldSelector{
				Name:      id.Name.String(),
				Namespace: id.Namespace.String(),
			},
		})
	}
	return sel
}

func makeFieldSelectorsByRefs(ids []domain.ResourceIdentifier, resType domain.ResourceType) []*common.ResSelector {
	sel := make([]*common.ResSelector, 0, len(ids))
	for _, id := range lo.Uniq(ids) {
		sel = append(sel, &common.ResSelector{
			FieldSelector: &common.FieldSelector{
				Refs: []*common.ResourceRef{
					{
						Name:      id.Name.String(),
						Namespace: id.Namespace.String(),
						ResType:   string(resType),
					},
				},
			},
		})
	}
	return sel
}

func loader[selT, pbT any, domainT interface {
	ResourceID() domain.ResourceIdentifier
}](
	ctx context.Context,
	selectors []selT,
	list func(context.Context, []selT) ([]pbT, error),
	conv func(pbT) (domainT, error),
	cache func(k domain.ResourceIdentifier, v domainT),
) error {
	if len(selectors) == 0 {
		return nil
	}
	items, err := list(ctx, selectors)
	if err != nil {
		return err
	}
	for _, pb := range items {
		item, err := conv(pb)
		if err != nil {
			return err
		}
		cache(item.ResourceID(), item)
	}
	return nil
}
