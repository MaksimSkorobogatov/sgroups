package misc

import (
	"bytes"
	"fmt"
)

// StringerF -
type StringerF func() string

// StringerF -
func (s StringerF) String() string {
	return s()
}

// Slice2Stringer -
func Slice2Stringer[t any](ar ...t) StringerF {
	return func() string {
		b := bytes.NewBuffer(nil)
		_ = b.WriteByte('[')
		for i, o := range ar {
			if i > 0 {
				_ = b.WriteByte(',')
			}
			s, _ := any(o).(fmt.Stringer)
			if s == nil {
				s, _ = any(&o).(fmt.Stringer) //nolint:gosec
			}
			_, _ = fmt.Fprintf(b,
				Tern(s != nil, "%s", "%v"),
				TernAny(s != nil, s, o),
			)
		}
		_ = b.WriteByte(']')
		return b.String()
	}
}
