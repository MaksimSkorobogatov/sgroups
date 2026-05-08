package misc

import (
	"github.com/H-BF/corlib/pkg/option"
)

// MatchOption -
func MatchOption[T any](v option.ValueOf[T]) (ret optMatcher[T]) {
	ret.v = v
	return ret
}

type (
	optMatcher[T any] struct {
		v option.ValueOf[T]
	}
	optNone[T any] optMatcher[T]
	optSome[T any] optMatcher[T]
)

// Some -
func (x optMatcher[T]) Some(f func(T)) optNone[T] {
	if obj, ok := x.v.Maybe(); ok {
		f(obj)
	}
	return optNone[T](x)
}

// None -
func (x optMatcher[T]) None(f func()) optSome[T] {
	if x.v.IsNone() {
		f()
	}
	return optSome[T](x)
}

// Some -
func (x optSome[T]) Some(f func(T)) {
	_ = optMatcher[T](x).Some(f)
}

// None -
func (x optNone[T]) None(f func()) {
	_ = optMatcher[T](x).None(f)
}
