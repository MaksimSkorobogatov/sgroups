package nft

import (
	"context"
	"time"

	"github.com/PRO-Robotech/sgroups/internal/shared/misc"
	"github.com/PRO-Robotech/sgroups/internal/shared/usecases"

	"github.com/PRO-Robotech/nftrace/pkg/watchers"
	agentv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/agent/v1"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

type (
	watchUseCase struct {
		cache        nftCache
		stream       grpc.ServerStreamingServer[agentv1.NftablesResp_Watch]
		nftCacheTTL  time.Duration
		syncInterval time.Duration
	}
)

// NewWatchUseCase creates a new instance of WatchUseCase
func NewWatchUseCase(stream grpc.ServerStreamingServer[agentv1.NftablesResp_Watch], nftCacheTTL time.Duration, syncInterval time.Duration) *watchUseCase {
	return &watchUseCase{
		stream:       stream,
		nftCacheTTL:  nftCacheTTL,
		syncInterval: syncInterval,
	}
}

// Watch -
func (uc *watchUseCase) Watch(ctx context.Context, _ *agentv1.NftablesReq_Watch) (err error) {
	defer func() {
		if err != nil {
			err = usecases.InternalError{Err: err}
		}
	}()
	ctx, cancel := misc.AnyContext(ctx, uc.stream.Context())
	defer cancel()
	if err = uc.cache.reload(uc.nftCacheTTL); err != nil {
		return errors.WithMessage(err, "reload nft cache")
	}
	nftWatcher, e := watchers.NftWatcher()
	if e != nil {
		return errors.WithMessage(e, "create nft watcher")
	}
	defer func() { _ = nftWatcher.Close() }()

	t := time.NewTicker(uc.syncInterval)
	defer t.Stop()

	for stm := nftWatcher.Stream(ctx); ; {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t.C:
			if !uc.cache.hasNewData(uc.nftCacheTTL) {
				continue
			}
			tbls, e := uc.cache.extractFormats()
			if e != nil {
				return e
			}
			err = uc.stream.Send(&agentv1.NftablesResp_Watch{
				Nftables: []*agentv1.Nftables{
					{
						Text: tbls.Text,
						Json: tbls.JSON,
					},
				},
			})
			if err != nil {
				return errors.WithMessage(err, "send nftables response")
			}
			uc.cache.markAsUsed()
		case msg, ok := <-stm:
			if !ok {
				return nil
			}
			if msg.Err != nil {
				return errors.WithMessage(msg.Err, "nft watcher stream")
			}
			uc.handleEvents(msg.Evt)
		}
	}
}

func (uc *watchUseCase) handleEvents(msg watchers.NftEvent) {
	var ok bool
	switch t := msg.Val.(type) {
	case watchers.TableEvent:
		ok = handleEvent(t.Action, t.Val, uc.cache.updTbl, uc.cache.delTbl)
	case watchers.ChainEvent:
		ok = handleEvent(t.Action, t.Val, uc.cache.updChain, uc.cache.delChain)
	case watchers.RuleEvent:
		ok = handleEvent(t.Action, t.Val.Rule, uc.cache.updRule, uc.cache.delRule)
	case watchers.SetEvent:
		ok = handleEvent(t.Action, t.Val, uc.cache.updSet, uc.cache.delSet)
	case watchers.SetElementEvent:
		ok = handleEvent(t.Action, t.Val, uc.cache.updSetElem, uc.cache.delSetElem)
	}
	if ok {
		uc.cache.markAsUpd()
	}
}

func handleEvent[T any](a watchers.Action, val T, upd, del func(T) bool) (ok bool) {
	switch a {
	case watchers.AddAction:
		ok = upd(val)
	case watchers.RmAction:
		ok = del(val)
	}
	return ok
}
