package nft

import (
	"bytes"
	"fmt"

	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"
	"github.com/PRO-Robotech/sgroups/internal/shared/misc"

	nftlib "github.com/google/nftables"
	util "github.com/google/nftables/binaryutil"
)

type (
	accports [][2]domain.PortNumber
)

// String -
func (accp accports) String() string {
	b := bytes.NewBuffer(nil)
	if len(accp) == 0 {
		b.WriteByte('*')
	} else {
		for i := range accp {
			if i > 0 {
				b.WriteByte(',')
			}
			x := accp[i]
			fmt.Fprintf(b, "%v", x[0])
			if x[0] != x[1] {
				fmt.Fprintf(b, "-%v", x[1])
			}
		}
	}
	return b.String()
}

// S - as SPorts
func (accp accports) S(rb ruleBuilder) ruleBuilder {
	return accp.imbueRuleBuilder(rb, true)
}

// D as DPorts
func (accp accports) D(rb ruleBuilder) ruleBuilder {
	return accp.imbueRuleBuilder(rb, false)
}

func (accp accports) imbueRuleBuilder(rb ruleBuilder, isSPorts bool) ruleBuilder {
	if n := len(accp); n == 1 {
		rb = misc.Tern(isSPorts, rb.SPort, rb.DPort)()
		if p := accp[0]; p[0] == p[1] {
			rb = rb.EqU16(p[0])
		} else {
			rb = rb.GeU16(p[0]).LeU16(p[1])
		}
	} else if n > 1 { //add anonimous port set
		set := &nftlib.Set{
			ID:        nextSetID(),
			Name:      "__set%d",
			Interval:  true,
			Anonymous: true,
			Constant:  true,
			KeyType:   nftlib.TypeInetService,
		}
		elements := make([]nftlib.SetElement, 0, 2*n)
		for _, p := range accp {
			elements = append(elements,
				nftlib.SetElement{
					Key: util.BigEndian.PutUint16(p[0]),
				},
				nftlib.SetElement{
					Key:         util.BigEndian.PutUint16(p[1] + 1),
					IntervalEnd: true,
				},
			)
		}
		rb = misc.Tern(isSPorts, rb.SPort, rb.DPort)().InSet(set)
		rb.PutSet(set.ID, NfSet{Set: set, Elements: elements})
	}
	return rb
}
