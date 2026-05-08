package dto

import (
	"reflect"

	"github.com/H-BF/corlib/pkg/dict"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

// ErrDTO -
var ErrDTO = errors.New("data transfer failure")

var registry dict.HDict[reflect.Type, any]

type (
	// Converter -
	Converter[From, To any] func(From) (To, error)

	// Pair - dto pair from -> to types.
	Pair[From, To any] struct {
		From From
		To   *To
	}

	tag[From, To any] struct{}
)

// Convert -
func (p *Pair[From, To]) Convert() error {
	return Convert(p.From, p.To)
}

// MakePair -
func MakePair[tFrom any, tTo any](a tFrom, b *tTo) *Pair[tFrom, tTo] {
	return &Pair[tFrom, tTo]{From: a, To: b}
}

// Register -
func Register[From, To any](fn Converter[From, To]) {
	registry.Put(reflect.TypeFor[tag[From, To]](), fn)
}

// Convert -
func Convert[From, To any](src From, dest *To) (err error) {
	fn := lookup[From, To]()
	if fn == nil {
		var from From
		var to To
		return errors.WithMessagef(
			ErrDTO,
			"no converter registered: %T -> %T", from, to,
		)
	}
	*dest, err = fn(src)
	if err != nil && !errors.Is(err, ErrDTO) {
		err = multierr.Append(err, ErrDTO)
	}
	return err
}

func lookup[From, To any]() Converter[From, To] {
	v, ok := registry.Get(reflect.TypeFor[tag[From, To]]())
	if !ok || v == nil {
		return nil
	}
	return v.(Converter[From, To])
}
