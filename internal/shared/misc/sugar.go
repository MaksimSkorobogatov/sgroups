package misc

import (
	"slices"

	"github.com/H-BF/corlib/pkg/dict"
	"github.com/pkg/errors"
)

// SliceToAny converts a slice of any type to a slice of empty interfaces.
func SliceToAny[In ~[]T, T any](s In) []any {
	r := Tern(len(s) > 0, make([]any, len(s)), nil)
	for i, v := range s {
		r[i] = any(v)
	}
	return r
}

// SliceToString converts a slice of string-like type to a slice of strings.
func SliceToString[In ~[]T, T ~string](s In) []string {
	r := Tern(len(s) > 0, make([]string, len(s)), nil)
	for i, v := range s {
		r[i] = string(v)
	}
	return r
}

// SliceToStringFunc converts a slice of any type to a slice of strings using a provided conversion function.
func SliceToStringFunc[In ~[]T, T any](s In, conv func(T) string) []string {
	r := Tern(len(s) > 0, make([]string, len(s)), nil)
	for i, v := range s {
		r[i] = conv(v)
	}
	return r
}

// Sli -
func Sli[T any](d ...T) []T {
	return d
}

// Tern -
func Tern[T any](cond bool, a1, a2 T) T {
	if cond {
		return a1
	}
	return a2
}

// TernAny -
func TernAny[t1 any, t2 any](cond bool, a1 t1, a2 t2) any {
	if cond {
		return a1
	}
	return a2
}

// Val2Ptr -
func Val2Ptr[T any](val T) *T {
	return &val
}

// MapContains checks if all key-value pairs in m2 are present in m1.
func MapContains[K comparable, V comparable](m1, m2 map[K]V) bool {
	for k, v2 := range m2 {
		v1, ok := m1[k]
		if !ok || v1 != v2 {
			return false
		}
	}
	return true
}

// IsIn checks if a value is present in a slice.
func IsIn[T comparable](test T, vals ...T) bool {
	return slices.Contains(vals, test)
}

// Send2ChanNoBlock -
func Send2ChanNoBlock[T any](ch chan<- T, a T) {
	select {
	case ch <- a:
	default:
	}
}

// ErrorIsInAny -
func ErrorIsInAny(e error, errs ...error) bool {
	if e == nil {
		return false
	}
	for _, test := range errs {
		if errors.Is(e, test) {
			return true
		}
	}
	return false
}

// HasSubset reports whether a contains every element of b.
func HasSubset[T comparable](a, b []T) bool {
	if len(b) == 0 {
		return true
	}
	set := dict.MakeHSet[T](a...)
	for _, x := range b {
		if !set.Contains(x) {
			return false
		}
	}
	return true
}
