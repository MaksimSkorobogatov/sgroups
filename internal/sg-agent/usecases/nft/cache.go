package nft

import (
	"fmt"
	"time"

	"github.com/PRO-Robotech/sgroups/internal/shared/misc"

	"github.com/H-BF/corlib/pkg/dict"
	"github.com/PRO-Robotech/nft-go/pkg/nftenc"
	"github.com/PRO-Robotech/nft-go/pkg/nftlist"
	"github.com/PRO-Robotech/nft-go/pkg/nlparser"
	nftLib "github.com/google/nftables"
	"github.com/samber/lo"
	"golang.org/x/sys/unix"
)

type (
	chainKey = string
	setKey   = string
	rlHandl  = uint64

	nftCache struct {
		dict.HDict[string, *tblItems]
		updAt  time.Time
		usedAt time.Time
	}
	tblItems struct {
		table  *nftLib.Table
		chains dict.HDict[chainKey, *chainItems]
		sets   dict.HDict[setKey, *setItems]
	}
	chainItems struct {
		chain *nftLib.Chain
		rules dict.HDict[rlHandl, *nftLib.Rule]
	}
	setItems struct {
		set   *nftLib.Set
		elems []nftLib.SetElement
	}
)

func (c *nftCache) updTbl(tbl *nftLib.Table) bool {
	if tbl == nil {
		return false
	}
	items := c.At(extractTblKey(*tbl))
	if items == nil {
		items = new(tblItems)
		defer c.Put(extractTblKey(*tbl), items)
	}
	items.table = tbl
	return true
}

func (c *nftCache) delTbl(tbl *nftLib.Table) bool {
	if tbl == nil {
		return false
	}
	c.Del(extractTblKey(*tbl))
	return true
}

func (c *nftCache) updChain(ch *nftLib.Chain) bool {
	if ch == nil || ch.Table == nil {
		return false
	}

	tblItems := c.At(extractTblKey(*ch.Table))
	if tblItems == nil {
		return false
	}
	chainItem := tblItems.chains.At(extractChKey(*ch))
	if chainItem == nil {
		chainItem = new(chainItems)
		defer tblItems.chains.Put(extractChKey(*ch), chainItem)
	}
	chainItem.chain = ch
	return true
}

func (c *nftCache) delChain(ch *nftLib.Chain) bool {
	if ch == nil || ch.Table == nil {
		return false
	}
	items := c.At(extractTblKey(*ch.Table))
	if items == nil {
		return false
	}
	items.chains.Del(extractChKey(*ch))
	return true
}

func (c *nftCache) updRule(rl *nftLib.Rule) bool {
	if rl == nil || rl.Chain == nil || rl.Table == nil {
		return false
	}
	tblKey := extractTblKey(*rl.Table)
	items := c.At(tblKey)
	if items == nil {
		return false
	}

	chainItem := items.chains.At(extractChKey(*rl.Chain))
	if chainItem == nil {
		return false
	}
	chainItem.rules.Put(rl.Handle, rl)
	return true
}

func (c *nftCache) delRule(rl *nftLib.Rule) bool {
	if rl == nil || rl.Chain == nil || rl.Table == nil {
		return false
	}

	tblItem := c.At(extractTblKey(*rl.Table))
	if tblItem == nil {
		return false
	}
	chainItem := tblItem.chains.At(extractChKey(*rl.Chain))
	if chainItem == nil {
		return false
	}
	chainItem.rules.Del(rl.Handle)
	return true
}

func (c *nftCache) updSet(st *nftLib.Set) bool {
	if st == nil || st.Table == nil {
		return false
	}

	tblItem := c.At(extractTblKey(*st.Table))
	if tblItem == nil {
		return false
	}
	setItem := tblItem.sets.At(extractSetKey(*st))
	if setItem == nil {
		setItem = new(setItems)
		defer tblItem.sets.Put(extractSetKey(*st), setItem)
	}
	setItem.set = st
	return true
}

func (c *nftCache) delSet(st *nftLib.Set) bool {
	if st == nil || st.Table == nil {
		return false
	}

	tblItem := c.At(extractTblKey(*st.Table))
	if tblItem == nil {
		return false
	}
	tblItem.sets.Del(extractSetKey(*st))
	return true
}

func (c *nftCache) updSetElem(st *nlparser.SetElems) bool {
	if st == nil || st.Table == nil {
		return false
	}

	tblItem := c.At(extractTblKey(*st.Table))
	if tblItem == nil {
		return false
	}
	setItem := tblItem.sets.At(extractSetElemKey(*st))
	if setItem == nil {
		setItem = new(setItems)
		defer tblItem.sets.Put(extractSetElemKey(*st), setItem)
	}
	setItem.elems = st.Elems
	return true
}

func (c *nftCache) delSetElem(st *nlparser.SetElems) bool {
	if st == nil || st.Table == nil {
		return false
	}

	tblItem := c.At(extractTblKey(*st.Table))
	if tblItem == nil {
		return false
	}
	setItem := tblItem.sets.At(extractSetElemKey(*st))
	if setItem == nil {
		return false
	}
	setItem.elems = nil
	return true
}

func (c *nftCache) reload(ttl time.Duration) error {
	tbls, err := nftLoadWithTTL(ttl)
	if err != nil {
		return err
	}
	c.Clear()
	for _, tbl := range tbls.Raw {
		if tbl.Table == nil {
			continue
		}
		items := c.At(extractTblKey(*tbl.Table))
		if items == nil {
			items = new(tblItems)
		}
		items.table = tbl.Table
		for _, ch := range tbl.Chains {
			if ch.Chain == nil {
				continue
			}
			chainItem := items.chains.At(extractChKey(*ch.Chain))
			if chainItem == nil {
				chainItem = new(chainItems)
			}
			chainItem.chain = ch.Chain

			for _, rl := range ch.Rules {
				if rl == nil {
					continue
				}
				chainItem.rules.Put(rl.Handle, rl)
			}
			items.chains.Put(extractChKey(*ch.Chain), chainItem)
		}
		for _, st := range tbl.Sets {
			if st.Set == nil {
				continue
			}
			setItem := items.sets.At(extractSetKey(*st.Set))
			if setItem == nil {
				setItem = new(setItems)
			}
			setItem.set = st.Set
			setItem.elems = lo.Map(st.Elems, func(el nftenc.SetElement, _ int) nftLib.SetElement {
				return nftLib.SetElement(el)
			})
			items.sets.Put(extractSetKey(*st.Set), setItem)
		}
		c.Put(extractTblKey(*tbl.Table), items)
	}
	c.markAsUpd()
	return nil
}

func (c *nftCache) extractFormats() (ret nftlist.TablesOutput, err error) {
	var tblEnc []*nftenc.TableEncoder
	for _, tblItems := range c.Iterate {
		if tblItems.table == nil {
			continue
		}
		if tblItems.table.Flags&uint32(unix.NFT_TABLE_F_DORMANT) != 0 {
			continue
		}

		var encs []nftenc.Encoder
		for _, chItems := range tblItems.chains.Iterate {
			var rlEncs []*nftenc.RuleEncoder
			for _, rl := range chItems.rules.Iterate {
				rlEncs = append(rlEncs, nftenc.NewRuleEncoder(rl))
			}
			encs = append(encs, nftenc.NewChainEncoder(chItems.chain, rlEncs...))
		}
		for _, setItems := range tblItems.sets.Iterate {
			encs = append(encs, nftenc.NewSetEncoder(setItems.set,
				nftenc.NewSetElemsEncoder(setItems.set.KeyType, setItems.set.Interval, setItems.elems)))
		}
		tblEnc = append(tblEnc, nftenc.NewTableEncoder(tblItems.table, encs...))
	}

	return nftlist.TablesFromEncoders(tblEnc, nftlist.WithoutRawOutput())
}

func (s *nftCache) hasNewData(delay time.Duration) bool {
	if !s.updAt.After(s.usedAt) {
		return false
	}
	if time.Since(s.updAt) < delay {
		return false
	}
	return s.hasReadyTables()
}

func (s *nftCache) hasReadyTables() bool {
	for _, tblItems := range s.Iterate {
		if tblItems.table == nil {
			continue
		}
		if tblItems.table.Flags&uint32(unix.NFT_TABLE_F_DORMANT) == 0 {
			return true
		}
	}
	return s.Len() == 0
}

func (s *nftCache) markAsUsed() {
	s.usedAt = time.Now()
}

func (s *nftCache) markAsUpd() {
	s.updAt = time.Now()
}

func extractTblKey(tbl nftLib.Table) string {
	return fmt.Sprintf("%s:%d", tbl.Name, tbl.Family)
}

func extractChKey(ch nftLib.Chain) string {
	return misc.Tern(ch.Table == nil, ch.Name, fmt.Sprintf("%s/%s", extractTblKey(*ch.Table), ch.Name))
}

func extractSetKey(st nftLib.Set) string {
	return misc.Tern(st.Table == nil, st.Name, fmt.Sprintf("%s/%s:%d", extractTblKey(*st.Table), st.Name, st.ID))
}

func extractSetElemKey(st nlparser.SetElems) string {
	return misc.Tern(st.Table == nil, st.SetName, fmt.Sprintf("%s/%s:%d", extractTblKey(*st.Table), st.SetName, st.SetId))
}
