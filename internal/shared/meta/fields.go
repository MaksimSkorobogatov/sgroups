package meta

import (
	"fmt"
	"iter"
	"reflect"
	"unsafe"

	"github.com/H-BF/corlib/pkg/dict"
)

type (
	// TagName is a type alias for a tag name in a struct.
	TagName = string
	// TagValue is a type alias for a tag value in a struct.
	TagValue = string
	// FieldInfo contains information about a field in a struct.
	FieldInfo struct {
		Type   reflect.Type
		Name   string
		Tags   map[TagName]TagValue
		Offset uintptr
		Index  []int
		Field  reflect.Value
	}
	// StructInspector provides introspection capabilities for a struct.
	StructInspector[T any] struct {
		structValue *T
		fieldsInfo  dict.RBDict[uintptr, *FieldInfo]
	}
)

// Value returns the value of the field represented by FieldInfo.
func (f *FieldInfo) Value() any {
	if f == nil || !f.Field.IsValid() || !f.Field.CanInterface() {
		return nil
	}
	return f.Field.Interface()
}

// NewStructInspector - creates a new StructInspector for the given struct pointer
func NewStructInspector[T any](obj *T, tags ...string) *StructInspector[T] {
	field := &StructInspector[T]{structValue: obj}
	StructIntrospect(obj, tags, func(info *FieldInfo) {
		if info != nil {
			field.fieldsInfo.Put(info.Offset, info)
		}
	})
	return field
}

// Field - returns the FieldInfo for the given field pointer
func (fi *StructInspector[T]) Field(fieldPtr uintptr) *FieldInfo {
	if fieldPtr == 0 {
		panic("Field: nil passed instead of a pointer")
	}

	base := ToUPtr(fi.structValue)
	return fi.fieldsInfo.At(fieldPtr - base)
}

// Iterator - iterates through the fields of the struct and yields FieldInfo
func (fi *StructInspector[T]) Iterator() iter.Seq[*FieldInfo] {
	return func(yield func(*FieldInfo) bool) {
		fi.fieldsInfo.Iterate(func(_ uintptr, v *FieldInfo) bool {
			return yield(v)
		})
	}
}

// StructIntrospect - iterate through struct fields and list by tags
func StructIntrospect(val any, discoverTags []string, cb func(info *FieldInfo)) {
	walkStruct(val, 0, func(field reflect.StructField, fv reflect.Value, off uintptr) {
		tagsMap := make(map[string]string, len(discoverTags))
		for _, tag := range discoverTags {
			if v, ok := field.Tag.Lookup(tag); ok {
				tagsMap[tag] = v
			}
		}
		cb(&FieldInfo{
			Type:   field.Type,
			Name:   field.Name,
			Tags:   tagsMap,
			Offset: off,
			Index:  field.Index,
			Field:  fv,
		})
	})
}

// GetFieldTag - return object field tag
func GetFieldTag[T any](obj *T, objFieldPtr any, tagName string) string {
	tags := make(map[uintptr]string)

	if reflect.TypeOf(objFieldPtr).Kind() != reflect.Pointer {
		panic("field of object must be a pointer")
	}

	StructIntrospect(obj, []string{tagName}, func(info *FieldInfo) {
		if info != nil && info.Tags[tagName] != "" {
			tags[info.Offset] = info.Tags[tagName]
		}
	})
	return tags[reflect.ValueOf(objFieldPtr).Pointer()-(uintptr)(unsafe.Pointer(obj))]
}

// ToUPtr - converts a pointer to uintptr
func ToUPtr[T any](p *T) uintptr {
	if p == nil {
		return 0
	}
	return uintptr(unsafe.Pointer(p))
}

func walkStruct(val any, baseOffset uintptr, onField func(sf reflect.StructField, fv reflect.Value, fieldOffset uintptr)) {
	rv, ok := derefValue(reflect.ValueOf(val))
	if !ok {
		return
	}
	if rv.Kind() != reflect.Struct {
		panic(fmt.Sprintf("StructIntrospect: expected struct, got %T", val))
	}

	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		if !field.IsExported() {
			continue
		}
		fv := rv.Field(i)

		base := field.Type
		for base.Kind() == reflect.Ptr {
			base = base.Elem()
		}

		if base.Kind() == reflect.Struct {
			nestedVal := fv
			if fv.Kind() == reflect.Ptr && fv.IsNil() {
				nestedVal = reflect.New(base).Elem()
			}
			if nestedVal.CanInterface() {
				walkStruct(
					nestedVal.Interface(),
					baseOffset+field.Offset,
					onField,
				)
			}
		}

		if fv.CanInterface() {
			onField(field, fv, baseOffset+field.Offset)
		}
	}
}

func derefValue(v reflect.Value) (reflect.Value, bool) {
	if !v.IsValid() {
		return v, false
	}
	for v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return v, false
		}
		v = v.Elem()
		if !v.IsValid() {
			return v, false
		}
	}
	return v, true
}
