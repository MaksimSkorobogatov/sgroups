package meta

import (
	"reflect"
)

// DedupCollections - deduplicates slices/arrays; recursively traverses structs and
// deduplicates their slice/array fields; recursively traverses maps ONLY through values and deduplicates their slice/array fields
func DedupCollections(objPtr any) {
	visited := make(map[uintptr]struct{})
	dedupValues(reflect.ValueOf(objPtr), visited)
}

func dedupValues(v reflect.Value, visited map[uintptr]struct{}) {
	if !v.IsValid() {
		return
	}

	if id, has := refIdentity(v); has {
		if _, seen := visited[id]; seen {
			return
		}
		visited[id] = struct{}{}
	}

	switch v.Kind() {
	case reflect.Pointer:
		if v.IsNil() {
			return
		}
		dedupValues(v.Elem(), visited)
		return

	case reflect.Interface:
		if v.IsNil() {
			return
		}
		inner := v.Elem()

		if v.CanSet() {
			tmp := reflect.New(inner.Type()).Elem()
			tmp.Set(inner)
			dedupValues(tmp, visited)
			v.Set(tmp)
			return
		}

		dedupValues(inner, visited)
		return

	case reflect.Struct:
		rt := v.Type()
		for i := 0; i < v.NumField(); i++ {
			sf := rt.Field(i)
			if !sf.IsExported() {
				continue
			}
			dedupValues(v.Field(i), visited)
		}
		return

	case reflect.Map:
		dedupMapValues(v, visited)
		return

	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			dedupValues(v.Index(i), visited)
		}
		dedupSeq(v)
		return
	}
}

func dedupMapValues(m reflect.Value, visited map[uintptr]struct{}) {
	if !m.IsValid() || m.Kind() != reflect.Map || m.IsNil() {
		return
	}

	iter := m.MapRange()
	for iter.Next() {
		k := iter.Key()
		val := iter.Value()

		switch val.Kind() {
		case reflect.Interface:
			if val.IsNil() {
				continue
			}
			inner := val.Elem()
			tmp := reflect.New(inner.Type()).Elem()
			tmp.Set(inner)
			dedupValues(tmp, visited)

			newIface := reflect.New(val.Type()).Elem()
			newIface.Set(tmp)
			m.SetMapIndex(k, newIface)
			continue

		case reflect.Pointer:
			dedupValues(val, visited)
			continue
		}

		switch val.Kind() {
		case reflect.Map:
			dedupMapValues(val, visited)

		case reflect.Struct:
			tmp := reflect.New(val.Type()).Elem()
			tmp.Set(val)
			dedupValues(tmp, visited)
			m.SetMapIndex(k, tmp)

		case reflect.Slice, reflect.Array:
			tmp := reflect.New(val.Type()).Elem()
			tmp.Set(val)
			for i := 0; i < tmp.Len(); i++ {
				dedupValues(tmp.Index(i), visited)
			}
			dedupSeq(tmp)
			m.SetMapIndex(k, tmp)
		}
	}
}

func dedupSeq(v reflect.Value) {
	if !v.IsValid() || v.Len() == 0 || !v.CanSet() {
		return
	}

	k := v.Kind()
	if k != reflect.Slice && k != reflect.Array {
		return
	}

	elemType := v.Type().Elem()
	if !elemType.Comparable() {
		return
	}

	emptyStructType := reflect.TypeOf(struct{}{})
	seen := reflect.MakeMapWithSize(reflect.MapOf(elemType, emptyStructType), v.Len())
	empty := reflect.Zero(emptyStructType)
	out := reflect.MakeSlice(reflect.SliceOf(elemType), 0, v.Len())

	for i := 0; i < v.Len(); i++ {
		e := v.Index(i)
		if seen.MapIndex(e).IsValid() {
			continue
		}
		seen.SetMapIndex(e, empty)
		out = reflect.Append(out, e)
	}

	switch k {
	case reflect.Slice:
		v.Set(out)
	case reflect.Array:
		n := out.Len()
		for i := 0; i < n; i++ {
			v.Index(i).Set(out.Index(i))
		}
		zero := reflect.Zero(elemType)
		for i := n; i < v.Len(); i++ {
			v.Index(i).Set(zero)
		}
	}
}

func refIdentity(v reflect.Value) (uintptr, bool) {
	switch v.Kind() {
	case reflect.Pointer, reflect.Map:
		if v.IsNil() {
			return 0, false
		}
		return v.Pointer(), true
	default:
		return 0, false
	}
}
