package protodriver

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"

	"github.com/graphql-editor/stucco/pkg/types"
	protoMessages "github.com/graphql-editor/stucco_proto/go/messages"
	"github.com/graphql-go/graphql/language/ast"
)

var (
	mapStringInterfaceReflectType = reflect.TypeOf((map[string]interface{})(nil))
	nilValue                      = &protoMessages.Value{
		TestValue: &protoMessages.Value_Nil{
			Nil: true,
		},
	}
)

func valueForType(t reflect.Type) *protoMessages.Value {
	var pv *protoMessages.Value
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		pv = &protoMessages.Value{
			TestValue: &protoMessages.Value_I{},
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		pv = &protoMessages.Value{
			TestValue: &protoMessages.Value_U{},
		}
	case reflect.Float32, reflect.Float64:
		pv = &protoMessages.Value{
			TestValue: &protoMessages.Value_F{},
		}
	case reflect.String:
		pv = &protoMessages.Value{
			TestValue: &protoMessages.Value_S{},
		}
	case reflect.Bool:
		pv = &protoMessages.Value{
			TestValue: &protoMessages.Value_B{},
		}
	case reflect.Slice, reflect.Array:
		if t.Elem().Kind() == reflect.Uint8 {
			pv = &protoMessages.Value{
				TestValue: &protoMessages.Value_Any{},
			}
		} else {
			pv = &protoMessages.Value{
				TestValue: &protoMessages.Value_A{},
			}
		}
	case reflect.Map, reflect.Struct:
		pv = &protoMessages.Value{
			TestValue: &protoMessages.Value_O{},
		}
	}
	return pv
}

func getValue(v reflect.Value) reflect.Value {
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return reflect.Value{}
		}
		v = v.Elem()
	}
	return v
}

func intToValue(v reflect.Value, pv *protoMessages.Value) {
	pv.TestValue.(*protoMessages.Value_I).I = v.Int()
}

func uintToValue(v reflect.Value, pv *protoMessages.Value) {
	pv.TestValue.(*protoMessages.Value_U).U = v.Uint()
}

func floatToValue(v reflect.Value, pv *protoMessages.Value) {
	pv.TestValue.(*protoMessages.Value_F).F = v.Float()
}

func stringToValue(v reflect.Value, pv *protoMessages.Value) {
	pv.TestValue.(*protoMessages.Value_S).S = v.String()
}

func boolToValue(v reflect.Value, pv *protoMessages.Value) {
	pv.TestValue.(*protoMessages.Value_B).B = v.Bool()
}

func bytesToValue(v reflect.Value, pv *protoMessages.Value) {
	bytesCopy := reflect.MakeSlice(
		reflect.SliceOf(v.Type().Elem()),
		v.Len(),
		v.Len(),
	)
	reflect.Copy(bytesCopy, v)
	pv.TestValue.(*protoMessages.Value_Any).Any = bytesCopy.Interface().([]byte)
}

func sliceOrArrayToValue(v reflect.Value, pv *protoMessages.Value) error {
	if v.Type().Elem().Kind() == reflect.Uint8 {
		bytesToValue(v, pv)
		return nil
	}
	arr := new(protoMessages.ArrayValue)
	if v.Len() != 0 {
		arr.Items = make([]*protoMessages.Value, 0, v.Len())
		for i := 0; i < v.Len(); i++ {
			item, err := anyToValueReflected(v.Index(i))
			if err != nil {
				return err
			}
			arr.Items = append(arr.Items, item)
		}
	}
	pv.TestValue.(*protoMessages.Value_A).A = arr
	return nil
}

func mapToValue(v reflect.Value, pv *protoMessages.Value) error {
	if v.Type().Key().Kind() != reflect.String {
		return fmt.Errorf("map key must be of string type")
	}
	obj := new(protoMessages.ObjectValue)
	if v.Len() > 0 {
		obj.Props = make(map[string]*protoMessages.Value)
		for _, k := range v.MapKeys() {
			v, err := anyToValueReflected(v.MapIndex(k))
			if err != nil {
				return err
			}
			obj.Props[k.String()] = v
		}
	}
	pv.TestValue.(*protoMessages.Value_O).O = obj
	return nil
}

// ValueMarshaler interface for client types that can return it's own protoMessages.Value
type ValueMarshaler interface {
	MarshalValue() (*protoMessages.Value, error)
}

type variable string

var variableType = reflect.TypeOf(variable(""))

var marshalerInterface = reflect.TypeOf((*ValueMarshaler)(nil)).Elem()

func anyToValueReflected(v reflect.Value) (*protoMessages.Value, error) {
	if !v.IsValid() {
		// Zero value, possibly nil
		return nilValue, nil
	}
	// short path for ValueMarshaler interface
	if v.Type().Implements(marshalerInterface) {
		return v.Interface().(ValueMarshaler).MarshalValue()
	}
	if v.Type().Kind() == reflect.Interface {
		if v.IsNil() {
			// empty value
			return nilValue, nil
		}
		v = v.Elem()
	}
	// Flatten GraphQL value types to an actual value
	v, err := flattenValue(v)
	if err != nil {
		return nil, err
	}
	t := v.Type()
	if t == variableType {
		return &protoMessages.Value{
			TestValue: &protoMessages.Value_Variable{
				Variable: v.String(),
			},
		}, nil
	}
	v = getValue(v)
	if !v.IsValid() {
		return nilValue, nil
	}
	t = v.Type()
	pv := valueForType(t)
	switch t.Kind() {
	case reflect.Ptr:
		// explicit error on pointer to pointer
		return nil, errors.New("pointer to pointer not supported")
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intToValue(v, pv)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintToValue(v, pv)
	case reflect.Float32, reflect.Float64:
		floatToValue(v, pv)
	case reflect.String:
		stringToValue(v, pv)
	case reflect.Bool:
		boolToValue(v, pv)
	case reflect.Slice, reflect.Array:
		err = sliceOrArrayToValue(v, pv)
	case reflect.Map:
		err = mapToValue(v, pv)
	case reflect.Struct:
		err = structToValue(v, pv)
	case reflect.Interface:
		if v.IsNil() {
			return nilValue, nil
		}
		fallthrough
	default:
		fmt.Printf("%v\n", v.Interface())
		return nil, fmt.Errorf("kind %s not supported", v.Kind())
	}
	return pv, err
}

func anyToValue(v interface{}) (*protoMessages.Value, error) {
	return anyToValueReflected(reflect.ValueOf(v))
}

func mapOfValueToMapOfAny(variables map[string]*protoMessages.Value, m map[string]*protoMessages.Value) (many map[string]interface{}, err error) {
	if len(m) == 0 {
		return
	}
	many = make(map[string]interface{}, len(m))
	for k, v := range m {
		many[k], err = valueToAny(variables, v)
		if err != nil {
			return
		}
	}
	return
}

func mapOfAnyToMapOfValue(m map[string]interface{}) (mval map[string]*protoMessages.Value, err error) {
	if len(m) == 0 {
		return
	}
	mval = make(map[string]*protoMessages.Value, len(m))
	for k, v := range m {
		mval[k], err = anyToValue(v)
		if err != nil {
			return
		}
	}
	return
}

func valueToAny(variables map[string]*protoMessages.Value, pv *protoMessages.Value) (v interface{}, err error) {
	if pv == nil || pv.GetTestValue() == nil {
		return
	}
	switch tv := pv.GetTestValue().(type) {
	case *protoMessages.Value_I:
		v = tv.I
	case *protoMessages.Value_U:
		v = tv.U
	case *protoMessages.Value_F:
		v = tv.F
	case *protoMessages.Value_S:
		v = tv.S
	case *protoMessages.Value_B:
		v = tv.B
	case *protoMessages.Value_A:
		var arr []interface{}
		if tv.A.GetItems() != nil {
			arr = make([]interface{}, 0, len(tv.A.GetItems()))
			for _, av := range tv.A.GetItems() {
				v, err := valueToAny(variables, av)
				if err != nil {
					return nil, err
				}
				arr = append(arr, v)
			}
		}
		v = arr
	case *protoMessages.Value_O:
		// There is no way to provide an output type for value decoding,
		// so just treat all always as map. Type information is lost.
		var m map[string]interface{}
		if tv.O.GetProps() != nil {
			m = make(map[string]interface{}, len(tv.O.GetProps()))
			for k, v := range tv.O.GetProps() {
				prop, err := valueToAny(variables, v)
				if err != nil {
					return nil, err
				}
				m[k] = prop
			}
		}
		v = m
	case *protoMessages.Value_Any:
		v = tv.Any
	case *protoMessages.Value_Variable:
		if variables != nil {
			if variableValue, ok := variables[tv.Variable]; ok {
				v, err = valueToAny(variables, variableValue)
			}
		}
	case *protoMessages.Value_Nil:
		v = nil
	}
	return
}

func makeProtoVariable(v types.Variable) *protoMessages.Variable {
	return &protoMessages.Variable{Name: v.Name}
}

func makeDriverVariable(v *protoMessages.Variable) types.Variable {
	return types.Variable{Name: v.GetName()}
}

func makeProtoVariableDefinition(v types.VariableDefinition) (vd *protoMessages.VariableDefinition, err error) {
	dv, err := anyToValue(v.DefaultValue)
	if err != nil {
		return
	}
	vd = &protoMessages.VariableDefinition{
		Variable:     makeProtoVariable(v.Variable),
		DefaultValue: dv,
	}
	return
}

func makeDriverVariableDefinition(v *protoMessages.VariableDefinition) (vd types.VariableDefinition, err error) {
	dv, err := valueToAny(nil, v.GetDefaultValue())
	if err != nil {
		return
	}
	vd = types.VariableDefinition{
		Variable:     makeDriverVariable(v.GetVariable()),
		DefaultValue: dv,
	}
	return
}

func makeProtoVariableDefinitions(v []types.VariableDefinition) (vd []*protoMessages.VariableDefinition, err error) {
	if len(v) == 0 {
		return nil, nil
	}
	r := make([]*protoMessages.VariableDefinition, 0, len(v))
	for _, vv := range v {
		var d *protoMessages.VariableDefinition
		d, err = makeProtoVariableDefinition(vv)
		if err != nil {
			return
		}
		r = append(r, d)
	}
	vd = r
	return
}

func makeDriverVariableDefinitions(v []*protoMessages.VariableDefinition) (vd []types.VariableDefinition, err error) {
	if len(v) == 0 {
		return nil, nil
	}
	vd = make([]types.VariableDefinition, 0, len(v))
	for _, vv := range v {
		var lv types.VariableDefinition
		lv, err = makeDriverVariableDefinition(vv)
		if err != nil {
			vd = nil
			return
		}
		vd = append(vd, lv)
	}
	return
}

func makeProtoDirective(v types.Directive) (dd *protoMessages.Directive, err error) {
	args, err := mapOfAnyToMapOfValue(v.Arguments)
	if err != nil {
		return
	}
	dd = &protoMessages.Directive{
		Name:      v.Name,
		Arguments: args,
	}
	return
}

func makeDriverDirective(variables map[string]*protoMessages.Value, v *protoMessages.Directive) (d types.Directive, err error) {
	args, err := mapOfValueToMapOfAny(variables, v.GetArguments())
	if err != nil {
		return
	}
	d = types.Directive{
		Name:      v.GetName(),
		Arguments: args,
	}
	return
}

func makeProtoDirectives(v types.Directives) (dd []*protoMessages.Directive, err error) {
	if v == nil {
		return nil, nil
	}
	r := make([]*protoMessages.Directive, 0, len(v))
	for _, dir := range v {
		var d *protoMessages.Directive
		d, err = makeProtoDirective(dir)
		if err != nil {
			return
		}
		r = append(r, d)
	}
	dd = r
	return
}

func makeDriverDirectives(variables map[string]*protoMessages.Value, v []*protoMessages.Directive) (dd types.Directives, err error) {
	if len(v) == 0 {
		return nil, nil
	}
	for _, dir := range v {
		var tdir types.Directive
		tdir, err = makeDriverDirective(variables, dir)
		if err != nil {
			dd = nil
			return
		}
		dd = append(dd, tdir)
	}
	return
}

func makeProtoFragmentDefinition(v *types.FragmentDefinition) (fd *protoMessages.FragmentDefinition, err error) {
	if v == nil {
		return
	}
	dirs, err := makeProtoDirectives(v.Directives)
	if err != nil {
		return
	}
	ss, err := makeProtoSelectionSet(v.SelectionSet)
	if err != nil {
		return
	}
	vd, err := makeProtoVariableDefinitions(v.VariableDefinitions)
	if err != nil {
		return
	}
	fd = &protoMessages.FragmentDefinition{
		Directives:          dirs,
		TypeCondition:       makeProtoTypeRef(&v.TypeCondition),
		SelectionSet:        ss,
		VariableDefinitions: vd,
	}
	return
}

func makeDriverFragmentDefinition(variables map[string]*protoMessages.Value, v *protoMessages.FragmentDefinition) (fd *types.FragmentDefinition, err error) {
	if v == nil {
		return
	}
	dirs, err := makeDriverDirectives(variables, v.GetDirectives())
	if err != nil {
		return
	}
	variableDefinitions, err := makeDriverVariableDefinitions(v.GetVariableDefinitions())
	if err != nil {
		return
	}
	selectionSet, err := makeDriverSelectionSet(variables, v.GetSelectionSet())
	if err != nil {
		return
	}
	fd = &types.FragmentDefinition{
		Directives:          dirs,
		TypeCondition:       mustMakeDriverTypeRef(v.GetTypeCondition()),
		SelectionSet:        selectionSet,
		VariableDefinitions: variableDefinitions,
	}
	return
}

func makeProtoSelection(v types.Selection) (s *protoMessages.Selection, err error) {
	args, err := mapOfAnyToMapOfValue(v.Arguments)
	if err != nil {
		return
	}
	dirs, err := makeProtoDirectives(v.Directives)
	if err != nil {
		return
	}
	ss, err := makeProtoSelectionSet(v.SelectionSet)
	if err != nil {
		return
	}
	fd, err := makeProtoFragmentDefinition(v.Definition)
	if err != nil {
		return
	}
	s = &protoMessages.Selection{
		Name:         v.Name,
		Arguments:    args,
		Directives:   dirs,
		SelectionSet: ss,
		Definition:   fd,
	}
	return
}

func makeDriverSelection(variables map[string]*protoMessages.Value, v *protoMessages.Selection) (s types.Selection, err error) {
	args, err := mapOfValueToMapOfAny(variables, v.GetArguments())
	if err != nil {
		return
	}
	dirs, err := makeDriverDirectives(variables, v.GetDirectives())
	if err != nil {
		return
	}
	fd, err := makeDriverFragmentDefinition(variables, v.GetDefinition())
	if err != nil {
		return
	}
	selectionSet, err := makeDriverSelectionSet(variables, v.GetSelectionSet())
	if err != nil {
		return
	}
	s = types.Selection{
		Name:         v.GetName(),
		Arguments:    args,
		Directives:   dirs,
		SelectionSet: selectionSet,
		Definition:   fd,
	}
	return
}

func makeProtoSelectionSet(v types.Selections) (ss []*protoMessages.Selection, err error) {
	if v == nil {
		return nil, nil
	}
	r := make([]*protoMessages.Selection, 0, len(v))
	for _, sel := range v {
		var s *protoMessages.Selection
		s, err = makeProtoSelection(sel)
		if err != nil {
			return
		}
		r = append(r, s)
	}
	ss = r
	return
}

func makeDriverSelectionSet(variables map[string]*protoMessages.Value, v []*protoMessages.Selection) (ss types.Selections, err error) {
	if len(v) == 0 {
		return nil, nil
	}
	for _, sel := range v {
		var s types.Selection
		s, err = makeDriverSelection(variables, sel)
		if err != nil {
			return
		}
		ss = append(ss, s)
	}
	return
}

func makeProtoOperationDefinition(v *types.OperationDefinition) (o *protoMessages.OperationDefinition, err error) {
	if v == nil {
		return nil, nil
	}
	vd, err := makeProtoVariableDefinitions(v.VariableDefinitions)
	if err != nil {
		return
	}
	dd, err := makeProtoDirectives(v.Directives)
	if err != nil {
		return
	}

	ss, err := makeProtoSelectionSet(v.SelectionSet)
	if err != nil {
		return
	}
	o = &protoMessages.OperationDefinition{
		Operation:           v.Operation,
		Name:                v.Name,
		VariableDefinitions: vd,
		Directives:          dd,
		SelectionSet:        ss,
	}
	return
}

func makeDriverOperationDefinition(variables map[string]*protoMessages.Value, v *protoMessages.OperationDefinition) (od *types.OperationDefinition, err error) {
	if v == nil {
		return
	}
	variableDefinitions, err := makeDriverVariableDefinitions(v.GetVariableDefinitions())
	if err != nil {
		return
	}
	dirs, err := makeDriverDirectives(variables, v.GetDirectives())
	if err != nil {
		return
	}
	selectionSet, err := makeDriverSelectionSet(variables, v.GetSelectionSet())
	if err != nil {
		return
	}
	od = &types.OperationDefinition{
		Operation:           v.GetOperation(),
		Name:                v.GetName(),
		VariableDefinitions: variableDefinitions,
		Directives:          dirs,
		SelectionSet:        selectionSet,
	}
	return
}

func makeProtoTypeRef(v *types.TypeRef) *protoMessages.TypeRef {
	if v == nil {
		return nil
	}
	var tt protoMessages.TypeRef
	switch {
	case v.Name != "":
		tt.TestTyperef = &protoMessages.TypeRef_Name{
			Name: v.Name,
		}
	case v.NonNull != nil:
		tt.TestTyperef = &protoMessages.TypeRef_NonNull{
			NonNull: makeProtoTypeRef(v.NonNull),
		}
	case v.List != nil:
		tt.TestTyperef = &protoMessages.TypeRef_List{
			List: makeProtoTypeRef(v.List),
		}
	}
	return &tt
}

func makeDriverTypeRef(v *protoMessages.TypeRef) *types.TypeRef {
	if v == nil {
		return nil
	}
	if name := v.GetName(); name != "" {
		return &types.TypeRef{Name: name}
	}
	if nonNull := v.GetNonNull(); nonNull != nil {
		return &types.TypeRef{NonNull: makeDriverTypeRef(nonNull)}
	}
	if list := v.GetList(); list != nil {
		return &types.TypeRef{List: makeDriverTypeRef(list)}
	}
	return &types.TypeRef{}
}

func mustMakeDriverTypeRef(v *protoMessages.TypeRef) types.TypeRef {
	return types.TypeRef{Name: v.GetName()}
}

func makeProtoResponsePath(v *types.ResponsePath) (*protoMessages.ResponsePath, error) {
	if v == nil {
		return nil, nil
	}
	prev, err := makeProtoResponsePath(v.Prev)
	var k *protoMessages.Value
	if err == nil {
		k, err = anyToValue(v.Key)
	}
	var rp *protoMessages.ResponsePath
	if err == nil {
		rp = &protoMessages.ResponsePath{
			Prev: prev,
			Key:  k,
		}
	}
	return rp, err
}

func makeDriverResponsePath(variables map[string]*protoMessages.Value, v *protoMessages.ResponsePath) (*types.ResponsePath, error) {
	if v == nil {
		return nil, nil
	}
	k, err := valueToAny(variables, v.GetKey())
	var rp *types.ResponsePath
	if err == nil {
		prev, err := makeDriverResponsePath(variables, v.GetPrev())
		if err == nil {
			rp = &types.ResponsePath{
				Key:  k,
				Prev: prev,
			}
		}
	}
	return rp, err
}

var (
	valueInterface         = reflect.TypeOf((*ast.Value)(nil)).Elem()
	astVariableType        = reflect.TypeOf((*ast.Variable)(nil))
	astIntValueType        = reflect.TypeOf((*ast.IntValue)(nil))
	astFloatValueType      = reflect.TypeOf((*ast.FloatValue)(nil))
	astBooleanValueType    = reflect.TypeOf((*ast.BooleanValue)(nil))
	astStringValueType     = reflect.TypeOf((*ast.StringValue)(nil))
	astEnumValueType       = reflect.TypeOf((*ast.EnumValue)(nil))
	astListValueType       = reflect.TypeOf((*ast.ListValue)(nil))
	astObjectValueType     = reflect.TypeOf((*ast.ObjectValue)(nil))
	astObjectFieldListType = reflect.TypeOf([]*ast.ObjectField{})
)

func flattenValue(v reflect.Value) (rv reflect.Value, err error) {
	rv = v
	if !v.IsValid() {
		return
	}
	if v.Kind() == reflect.Interface {
		if v.IsNil() {
			return
		}
		v = v.Elem()
	}
	if v.Type().Implements(valueInterface) || v.Type() == astObjectFieldListType {
		switch v.Type() {
		case astVariableType:
			rv = v.Elem().FieldByName("Name").Elem().FieldByName("Value").Convert(variableType)
		case astIntValueType:
			var i int64
			i, err = strconv.ParseInt(v.Elem().FieldByName("Value").String(), 10, 32)
			if err == nil {
				rv = reflect.ValueOf(i)
			}
		case astFloatValueType:
			var f float64
			f, err = strconv.ParseFloat(v.Elem().FieldByName("Value").String(), 64)
			if err == nil {
				rv = reflect.ValueOf(f)
			}
		case astBooleanValueType, astStringValueType, astEnumValueType:
			rv = v.Elem().FieldByName("Value")
		case astListValueType:
			rv = v.Elem().FieldByName("Values")
		case astObjectValueType:
			// Handle list of object fields like a map
			v = v.Elem().FieldByName("Fields")
			fallthrough
		case astObjectFieldListType:
			rv = reflect.MakeMap(mapStringInterfaceReflectType)
			for i := 0; i < v.Len(); i++ {
				f := v.Index(i).Elem()
				rv.SetMapIndex(
					f.FieldByName("Name").Elem().FieldByName("Value"),
					f.FieldByName("Value"),
				)
			}
		default:
			rv = v.MethodByName("GetValue").Call([]reflect.Value{})[0]
		}
	}
	return
}

func initVariablesWithDefaults(variables map[string]*protoMessages.Value, opDef *protoMessages.OperationDefinition) map[string]*protoMessages.Value {
	nv := make(map[string]*protoMessages.Value)
	if opDef != nil {
		varDefs := opDef.VariableDefinitions
		for _, vd := range varDefs {
			if v, ok := variables[vd.Variable.Name]; ok {
				nv[vd.Variable.Name] = v
			} else {
				nv[vd.Variable.Name] = vd.DefaultValue
			}
		}
	}
	return nv
}
