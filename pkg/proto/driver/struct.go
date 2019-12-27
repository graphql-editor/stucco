package protodriver

import (
	"reflect"
	"strings"
	"sync"

	"github.com/graphql-editor/stucco/pkg/proto"
	"k8s.io/klog"
)

func getTag(field *reflect.StructField) (string, []string) {
	tag := field.Tag.Get("stucco")
	if tag == "" {
		tag = field.Tag.Get("json")
		if tag == "" {
			return "", nil
		}
	}
	parts := strings.Split(tag, ",")
	return parts[0], parts[1:]
}

var fieldCache sync.Map

type field struct {
	typ    reflect.Type
	name   string
	index  []int
	tagged bool
	encode func(v reflect.Value) (*proto.Value, error)
}

func wrap(f func(reflect.Value, *proto.Value), mv func() *proto.Value) func(reflect.Value) (*proto.Value, error) {
	return func(v reflect.Value) (*proto.Value, error) {
		pv := mv()
		f(v, pv)
		return pv, nil
	}
}

func encodeFuncForType(t reflect.Type) func(reflect.Value) (*proto.Value, error) {
	if t.Implements(marshalerInterface) {
		return func(v reflect.Value) (*proto.Value, error) {
			return v.Interface().(ValueMarshaler).MarshalValue()
		}
	}
	switch t.Kind() {
	case reflect.Interface:
		// can't really do anything with interface, needs full check
		return anyToValueReflected
	case reflect.Ptr:
		t = t.Elem()
	}
	switch t.Kind() {
	case reflect.Ptr:
		klog.Warning("pointer to pointer types are not supported")
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return wrap(intToValue, func() *proto.Value {
			return &proto.Value{
				TestValue: &proto.Value_I{},
			}
		})
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return wrap(uintToValue, func() *proto.Value {
			return &proto.Value{
				TestValue: &proto.Value_U{},
			}
		})
	case reflect.Float32, reflect.Float64:
		return wrap(floatToValue, func() *proto.Value {
			return &proto.Value{
				TestValue: &proto.Value_F{},
			}
		})
	case reflect.String:
		return wrap(stringToValue, func() *proto.Value {
			return &proto.Value{
				TestValue: &proto.Value_S{},
			}
		})
	case reflect.Bool:
		return wrap(boolToValue, func() *proto.Value {
			return &proto.Value{
				TestValue: &proto.Value_B{},
			}
		})
	case reflect.Slice, reflect.Array:
		if t.Elem().Kind() == reflect.Uint8 {
			return func(v reflect.Value) (*proto.Value, error) {
				pv := &proto.Value{
					TestValue: &proto.Value_Any{},
				}
				bytesToValue(v, pv)
				return pv, nil
			}
		} else {
			return func(v reflect.Value) (*proto.Value, error) {
				pv := &proto.Value{
					TestValue: &proto.Value_A{},
				}
				return pv, sliceOrArrayToValue(v, pv)
			}
		}
	case reflect.Map:
		return func(v reflect.Value) (*proto.Value, error) {
			pv := &proto.Value{
				TestValue: &proto.Value_O{},
			}
			return pv, mapToValue(v, pv)
		}
	case reflect.Struct:
		return func(v reflect.Value) (*proto.Value, error) {
			pv := &proto.Value{
				TestValue: &proto.Value_O{},
			}
			return pv, structToValue(v, pv)
		}
	}
	klog.Warningf("kind %s is unsupported", t.Kind())
	return nil
}

func typeFields(t reflect.Type) []field {
	current := []field{}
	next := []field{{typ: t}}
	visited := map[reflect.Type]bool{}
	fieldAt := map[string]int{}
	var fields []field
	for len(next) > 0 {
		current, next = next, current[:0]
		nextCount := map[reflect.Type]bool{}
		for _, f := range current {
			if visited[f.typ] {
				continue
			}
			visited[f.typ] = true
			for i := 0; i < f.typ.NumField(); i++ {
				sf := f.typ.Field(i)
				isUnexported := sf.PkgPath != ""
				if sf.Anonymous {
					t := sf.Type
					if t.Kind() == reflect.Ptr {
						t = t.Elem()
					}
					if isUnexported && t.Kind() != reflect.Struct {
						continue
					}
				} else if isUnexported {
					continue
				}
				tag, _ := getTag(&sf)
				if tag == "-" {
					continue
				}
				index := make([]int, len(f.index)+1)
				copy(index, f.index)
				index[len(f.index)] = i
				name := tag
				ft := sf.Type
				encodeFunc := encodeFuncForType(ft)
				if ft.Name() == "" && ft.Kind() == reflect.Ptr {
					ft = ft.Elem()
				}
				if name != "" || !sf.Anonymous || ft.Kind() != reflect.Struct {
					if encodeFunc == nil {
						continue
					}
					tagged := name != ""
					if name == "" {
						name = sf.Name
					}
					// Anonymous structs define fields with the same
					// name/tag, rules for picking a best matching field:
					// 1. by depth - if there are names clashing in struct
					//              but they have different have different depth,
					//              pick the more shallow one
					// 2. by tag - if there are names clashing in struct
					//             on the same depth, choose the one that's tagged
					// 3. otherwise - in any other case the one that apeared first
					if fAt, ok := fieldAt[name]; ok {
						if len(index) < len(fields[fAt].index) ||
							len(index) == len(fields[fAt].index) &&
								tagged && !fields[fAt].tagged {
							fields = append(fields[:fAt], fields[fAt+1:]...)
						}
						continue
					}
					fieldAt[name] = len(fields)
					fields = append(fields, field{
						typ:    ft,
						tagged: tagged,
						name:   name,
						index:  index,
						encode: encodeFunc,
					})
					continue
				}
				if !nextCount[ft] {
					nextCount[ft] = true
					next = append(next, field{index: index, typ: ft})
				}
			}
		}
	}
	return fields
}

func cachedTypeFields(t reflect.Type) []field {
	if f, ok := fieldCache.Load(t); ok {
		return f.([]field)
	}
	f, _ := fieldCache.LoadOrStore(t, typeFields(t))
	return f.([]field)
}

func structToValue(v reflect.Value, pv *proto.Value) error {
	if v = getValue(v); v.IsValid() {
		fields := cachedTypeFields(v.Type())
		if len(fields) == 0 {
			// empty struct or struct with only unexported fields
			return nil
		}
		obj := &proto.ObjectValue{
			Props: make(map[string]*proto.Value, len(fields)),
		}
		for i := 0; i < len(fields); i++ {
			fv := v.Field(fields[i].index[0])
			for _, idx := range fields[i].index[1:] {
				fv = fv.Field(idx)
			}
			var err error
			obj.Props[fields[i].name], err = fields[i].encode(fv)
			if err != nil {
				return err
			}
		}
		pv.TestValue.(*proto.Value_O).O = obj
	}
	return nil
}
