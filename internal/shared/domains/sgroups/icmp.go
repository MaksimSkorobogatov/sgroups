package sgroups

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/H-BF/corlib/pkg/dict"
)

// IcmpTypes represents set of ICMP types
type IcmpTypes dict.RBSet[uint8]

// Eq -
func (o IcmpTypes) Eq(other IcmpTypes) bool {
	return (*dict.RBSet[uint8])(&o).Eq((*dict.RBSet[uint8])(&other))
}

// String -
func (o IcmpTypes) String() string {
	b := bytes.NewBuffer(nil)
	_, _ = b.WriteString("ICMP")

	if i := 0; o.Len() > 0 {
		for k := range o.Iterate {
			if i++; i > 1 {
				_ = b.WriteByte(',')
			}
			_, _ = fmt.Fprintf(b, "%v", k)
		}
	}

	return b.String()
}

// MarshalJSON -
func (ty IcmpTypes) MarshalJSON() ([]byte, error) {
	b := bytes.NewBuffer(nil)
	e := b.WriteByte('[')
	if e != nil {
		return nil, e
	}
	var i int
	ty.Iterate(func(k uint8) bool {
		if i++; i > 1 {
			e = b.WriteByte(',')
			if e != nil {
				return false
			}
		}
		_, e = fmt.Fprintf(b, "%v", k)
		return e == nil
	})
	if e != nil {
		return nil, e
	}
	if e = b.WriteByte(']'); e != nil {
		return nil, e
	}
	return b.Bytes(), nil
}

// UnmarshalJSON -
func (ty *IcmpTypes) UnmarshalJSON(data []byte) error {
	var x []uint8
	e := json.Unmarshal(data, &x)
	if e == nil {
		ty.Clear()
		ty.PutMany(x...)
	}
	return e
}
