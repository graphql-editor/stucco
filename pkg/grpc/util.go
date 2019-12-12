package grpc

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"

	"github.com/graphql-editor/stucco/pkg/proto"
	"github.com/graphql-editor/stucco/pkg/types"
	"github.com/graphql-go/graphql/language/ast"
)

var sliceInterfaceReflectType = reflect.SliceOf(reflect.TypeOf((*interface{})(nil)).Elem())
var mapStringInterfaceReflectType = reflect.TypeOf((map[string]interface{})(nil))

func getValue(v reflect.Value) reflect.Value {
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return reflect.Value{}
		}
		v = v.Elem()
	}
	return v
}

func intToValue(v reflect.Value) (pv *proto.Value) {
	iValue := &proto.Value_I{}
	pv = &proto.Value{TestValue: iValue}
	if v = getValue(v); v.IsValid() {
		iValue.I = v.Int()
	}
	return
}

func uintToValue(v reflect.Value) (pv *proto.Value) {
	uValue := &proto.Value_U{}
	pv = &proto.Value{TestValue: uValue}
	if v = getValue(v); v.IsValid() {
		uValue.U = v.Uint()
	}
	return
}

func floatToValue(v reflect.Value) (pv *proto.Value) {
	fValue := &proto.Value_F{}
	pv = &proto.Value{TestValue: fValue}
	if v = getValue(v); v.IsValid() {
		fValue.F = v.Float()
	}
	return
}

func stringToValue(v reflect.Value) (pv *proto.Value) {
	sValue := &proto.Value_S{}
	pv = &proto.Value{TestValue: sValue}
	if v = getValue(v); v.IsValid() {
		sValue.S = v.String()
	}
	return
}

func boolToValue(v reflect.Value) (pv *proto.Value) {
	bValue := &proto.Value_B{}
	pv = &proto.Value{TestValue: bValue}
	if v = getValue(v); v.IsValid() {
		bValue.B = v.Bool()
	}
	return
}

func bytesToValue(v reflect.Value) *proto.Value {
	protoValue := new(proto.Value)
	any := &proto.Value_Any{}
	protoValue.TestValue = any
	if v = getValue(v); v.IsValid() {
		bytesCopy := reflect.MakeSlice(
			reflect.SliceOf(v.Type().Elem()),
			v.Len(),
			v.Len(),
		)
		reflect.Copy(bytesCopy, v)
		any.Any = bytesCopy.Interface().([]byte)
	}
	return protoValue
}

func sliceOrArrayToValue(v reflect.Value) (*proto.Value, error) {
	if v = getValue(v); !v.IsValid() {
		return &proto.Value{TestValue: &proto.Value_A{}}, nil
	}
	if v.Type().Elem().Kind() == reflect.Uint8 {
		return bytesToValue(v), nil
	}
	arr := new(proto.ArrayValue)
	protoValue := &proto.Value{
		TestValue: &proto.Value_A{
			A: arr,
		},
	}
	arr.Items = make([]*proto.Value, 0, v.Len())
	for i := 0; i < v.Len(); i++ {
		item, err := anyToValueReflected(v.Index(i))
		if err != nil {
			return nil, err
		}
		arr.Items = append(arr.Items, item)
	}
	return protoValue, nil
}

func mapToValue(v reflect.Value) (*proto.Value, error) {
	obj := new(proto.Value_O)
	protoValue := &proto.Value{
		TestValue: obj,
	}
	if v = getValue(v); v.IsValid() {
		obj.O = new(proto.ObjectValue)
		if v.Type().Key().Kind() != reflect.String {
			return nil, fmt.Errorf("map key must be of string type")
		}
		obj.O.Props = make(map[string]*proto.Value)
		for _, k := range v.MapKeys() {
			v, err := anyToValueReflected(v.MapIndex(k))
			if err != nil {
				return nil, err
			}
			obj.O.Props[k.String()] = v
		}
	}
	return protoValue, nil
}

// ValueMarshaler interface for client types that can return it's own proto.Value
type ValueMarshaler interface {
	MarshalValue() (*proto.Value, error)
}

type variable string

var variableType = reflect.TypeOf(variable(""))

var marshalerInterface = reflect.TypeOf((*ValueMarshaler)(nil)).Elem()

func anyToValueReflected(v reflect.Value) (*proto.Value, error) {
	if !v.IsValid() {
		// Zero value, possibly nil
		return new(proto.Value), nil
	}
	// short path for ValueMarshaler interface
	if v.Type().Implements(marshalerInterface) {
		return v.Interface().(ValueMarshaler).MarshalValue()
	}
	if v.Type().Kind() == reflect.Interface {
		if v.IsNil() {
			// empty value
			return new(proto.Value), nil
		}
		v = v.Elem()
	}
	// Flatten GraphQL value types to an actual value
	v, err := flattenValue(v)
	if err != nil {
		return nil, err
	}
	if !v.IsValid() {
		return new(proto.Value), nil
	}
	t := v.Type()
	if t == variableType {
		return &proto.Value{
			TestValue: &proto.Value_Variable{
				Variable: v.String(),
			},
		}, nil
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	var protoValue *proto.Value
	switch t.Kind() {
	case reflect.Ptr:
		// explicit error on pointer to pointer
		return nil, errors.New("pointer to pointer not supported")
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		protoValue = intToValue(v)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		protoValue = uintToValue(v)
	case reflect.Float32, reflect.Float64:
		protoValue = floatToValue(v)
	case reflect.String:
		protoValue = stringToValue(v)
	case reflect.Bool:
		protoValue = boolToValue(v)
	case reflect.Slice, reflect.Array:
		protoValue, err = sliceOrArrayToValue(v)
	case reflect.Map:
		protoValue, err = mapToValue(v)
	case reflect.Struct:
		protoValue, err = structToValue(v)
	default:
		fmt.Printf("%v\n", v.Interface())
		return nil, fmt.Errorf("kind %s not supported", v.Kind())
	}
	return protoValue, err
}

func anyToValue(v interface{}) (*proto.Value, error) {
	return anyToValueReflected(reflect.ValueOf(v))
}

func mapOfValueToMapOfAny(variables map[string]*proto.Value, m map[string]*proto.Value) (many map[string]interface{}, err error) {
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

func mapOfAnyToMapOfValue(m map[string]interface{}) (mval map[string]*proto.Value, err error) {
	if len(m) == 0 {
		return
	}
	mval = make(map[string]*proto.Value, len(m))
	for k, v := range m {
		mval[k], err = anyToValue(v)
		if err != nil {
			return
		}
	}
	return
}

func valueToAny(variables map[string]*proto.Value, pv *proto.Value) (v interface{}, err error) {
	if pv == nil || pv.GetTestValue() == nil {
		return
	}
	switch tv := pv.GetTestValue().(type) {
	case *proto.Value_I:
		v = tv.I
	case *proto.Value_U:
		v = tv.U
	case *proto.Value_F:
		v = tv.F
	case *proto.Value_S:
		v = tv.S
	case *proto.Value_B:
		v = tv.B
	case *proto.Value_A:
		var arr []interface{}
		if tv.A != nil {
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
	case *proto.Value_O:
		// There is no way to provide an output type for value decoding,
		// so just treat all always as map. Type information is lost.
		var m map[string]interface{}
		if tv.O != nil {
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
	case *proto.Value_Any:
		v = tv.Any
	case *proto.Value_Variable:
		if variables != nil {
			if variableValue, ok := variables[tv.Variable]; ok {
				v, err = valueToAny(variables, variableValue)
			}
		}
	}
	return
}

func makeProtoVariable(v types.Variable) *proto.Variable {
	return &proto.Variable{Name: v.Name}
}

func makeDriverVariable(v *proto.Variable) types.Variable {
	return types.Variable{Name: v.GetName()}
}

func makeProtoVariableDefinition(v types.VariableDefinition) (vd *proto.VariableDefinition, err error) {
	dv, err := anyToValue(v.DefaultValue)
	if err != nil {
		return
	}
	vd = &proto.VariableDefinition{
		Variable:     makeProtoVariable(v.Variable),
		DefaultValue: dv,
	}
	return
}

func makeDriverVariableDefinition(v *proto.VariableDefinition) (vd types.VariableDefinition, err error) {
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

func makeProtoVariableDefinitions(v []types.VariableDefinition) (vd []*proto.VariableDefinition, err error) {
	r := make([]*proto.VariableDefinition, 0, len(v))
	for _, vv := range v {
		var d *proto.VariableDefinition
		d, err = makeProtoVariableDefinition(vv)
		if err != nil {
			return
		}
		r = append(r, d)
	}
	vd = r
	return
}

func makeDriverVariableDefinitions(v []*proto.VariableDefinition) (vd []types.VariableDefinition, err error) {
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

func makeProtoDirective(v types.Directive) (dd *proto.Directive, err error) {
	args, err := mapOfAnyToMapOfValue(v.Arguments)
	if err != nil {
		return
	}
	dd = &proto.Directive{
		Name:      v.Name,
		Arguments: args,
	}
	return
}

func makeDriverDirective(variables map[string]*proto.Value, v *proto.Directive) (d types.Directive, err error) {
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

func makeProtoDirectives(v types.Directives) (dd []*proto.Directive, err error) {
	if v == nil {
		return nil, nil
	}
	r := make([]*proto.Directive, 0, len(v))
	for _, dir := range v {
		var d *proto.Directive
		d, err = makeProtoDirective(dir)
		if err != nil {
			return
		}
		r = append(r, d)
	}
	dd = r
	return
}

func makeDriverDirectives(variables map[string]*proto.Value, v []*proto.Directive) (dd types.Directives, err error) {
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

func makeProtoFragmentDefinition(v *types.FragmentDefinition) (fd *proto.FragmentDefinition, err error) {
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
	fd = &proto.FragmentDefinition{
		Directives:          dirs,
		TypeCondition:       makeProtoTypeRef(&v.TypeCondition),
		SelectionSet:        ss,
		VariableDefinitions: vd,
	}
	return
}

func makeDriverFragmentDefinition(variables map[string]*proto.Value, v *proto.FragmentDefinition) (fd *types.FragmentDefinition, err error) {
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

func makeProtoSelection(v types.Selection) (s *proto.Selection, err error) {
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
	s = &proto.Selection{
		Name:         v.Name,
		Arguments:    args,
		Directives:   dirs,
		SelectionSet: ss,
		Definition:   fd,
	}
	return
}

func makeDriverSelection(variables map[string]*proto.Value, v *proto.Selection) (s types.Selection, err error) {
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

func makeProtoSelectionSet(v types.Selections) (ss []*proto.Selection, err error) {
	if v == nil {
		return nil, nil
	}
	r := make([]*proto.Selection, 0, len(v))
	for _, sel := range v {
		var s *proto.Selection
		s, err = makeProtoSelection(sel)
		if err != nil {
			return
		}
		r = append(r, s)
	}
	ss = r
	return
}

func makeDriverSelectionSet(variables map[string]*proto.Value, v []*proto.Selection) (ss types.Selections, err error) {
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

func makeProtoOperationDefinition(v *types.OperationDefinition) (o *proto.OperationDefinition, err error) {
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
	o = &proto.OperationDefinition{
		Operation:           v.Operation,
		Name:                v.Name,
		VariableDefinitions: vd,
		Directives:          dd,
		SelectionSet:        ss,
	}
	return
}

func makeDriverOperationDefinition(variables map[string]*proto.Value, v *proto.OperationDefinition) (od *types.OperationDefinition, err error) {
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

func makeProtoTypeRef(v *types.TypeRef) *proto.TypeRef {
	if v == nil {
		return nil
	}
	var tt proto.TypeRef
	switch {
	case v.Name != "":
		tt.TestTyperef = &proto.TypeRef_Name{
			Name: v.Name,
		}
	case v.NonNull != nil:
		tt.TestTyperef = &proto.TypeRef_NonNull{
			NonNull: makeProtoTypeRef(v.NonNull),
		}
	case v.List != nil:
		tt.TestTyperef = &proto.TypeRef_List{
			List: makeProtoTypeRef(v.List),
		}
	}
	return &tt
}

func makeDriverTypeRef(v *proto.TypeRef) *types.TypeRef {
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

func mustMakeDriverTypeRef(v *proto.TypeRef) types.TypeRef {
	return types.TypeRef{Name: v.GetName()}
}

func makeProtoResponsePath(v *types.ResponsePath) (*proto.ResponsePath, error) {
	if v == nil {
		return nil, nil
	}
	prev, err := makeProtoResponsePath(v.Prev)
	var k *proto.Value
	if err == nil {
		k, err = anyToValue(v.Key)
	}
	var rp *proto.ResponsePath
	if err == nil {
		rp = &proto.ResponsePath{
			Prev: prev,
			Key:  k,
		}
	}
	return rp, err
}

func makeDriverResponsePath(variables map[string]*proto.Value, v *proto.ResponsePath) (*types.ResponsePath, error) {
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

func initVariablesWithDefaults(variables map[string]*proto.Value, opDef *proto.OperationDefinition) map[string]*proto.Value {
	nv := make(map[string]*proto.Value)
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
