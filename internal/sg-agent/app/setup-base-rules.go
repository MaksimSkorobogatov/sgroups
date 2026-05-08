package app

import (
	"context"
	"encoding/json"
	"unsafe"

	conf "github.com/PRO-Robotech/sgroups/internal/sg-agent/config"
	rc "github.com/PRO-Robotech/sgroups/internal/sg-agent/resources"

	"github.com/pkg/errors"
)

// IfBaseRulesFromConfig -
func IfBaseRulesFromConfig(ctx context.Context, cons func(rc.BaseRuleList) error) (err error) {
	defer func() {
		err = errors.WithMessage(err, "load base rule list")
	}()
	var (
		data string
		br   rc.BaseRuleList
	)
	data, err = conf.BaseRulesConfig.Value(ctx)
	if err != nil && !errors.Is(err, conf.ErrNotFound) {
		return err
	} else if errors.Is(err, conf.ErrNotFound) {
		var (
			rl     rc.BaseRuleList
			sgAddr string
		)
		if sgAddr, err = conf.SGroupsAddress.Value(ctx); err != nil {
			return err
		}
		if rl, err = rc.MakeDefaultBaseRules(ctx, sgAddr); err != nil {
			return err
		}
		return cons(rl)
	}

	if err = json.Unmarshal(unsafe.Slice(unsafe.StringData(data), len(data)), &br); err != nil {
		return err
	}

	return cons(br)
}
