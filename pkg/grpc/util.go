package grpc

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/graphql-editor/stucco/pkg/proto"
	"github.com/graphql-editor/stucco/pkg/types"
	"github.com/graphql-go/graphql/language/ast"
)

func intToValue(v reflect.Value) (pv *proto.Value) {
	iValue := &proto.Value_I{}
	pv = &proto.Value{TestValue: iValue}
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return
		}
		v = v.Elem()
	}
	iValue.I = v.Int()
	return
}

func uintToValue(v reflect.Value) (pv *proto.Value) {
	uValue := &proto.Value_U{}
	pv = &proto.Value{TestValue: uValue}
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return
		}
		v = v.Elem()
	}
	uValue.U = v.Uint()
	return
}

func floatToValue(v reflect.Value) (pv *proto.Value) {
	fValue := &proto.Value_F{}
	pv = &proto.Value{TestValue: fValue}
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return
		}
		v = v.Elem()
	}
	fValue.F = v.Float()
	return
}

func stringToValue(v reflect.Value) (pv *proto.Value) {
	sValue := &proto.Value_S{}
	pv = &proto.Value{TestValue: sValue}
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return
		}
		v = v.Elem()
	}
	sValue.S = v.String()
	return
}

func boolToValue(v reflect.Value) (pv *proto.Value) {
	bValue := &proto.Value_B{}
	pv = &proto.Value{TestValue: bValue}
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return
		}
		v = v.Elem()
	}
	bValue.B = v.Bool()
	return
}

func bytesToValue(v reflect.Value) *proto.Value {
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return &proto.Value{TestValue: &proto.Value_Any{}}
		}
		v = v.Elem()
	}
	protoValue := new(proto.Value)
	bytesCopy := reflect.MakeSlice(
		reflect.SliceOf(v.Type().Elem()),
		v.Len(),
		v.Len(),
	)
	reflect.Copy(bytesCopy, v)
	protoValue.TestValue = &proto.Value_Any{
		Any: bytesCopy.Interface().([]byte),
	}
	return protoValue
}

func sliceOrArrayToValue(v reflect.Value) (*proto.Value, error) {
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return &proto.Value{TestValue: &proto.Value_A{}}, nil
		}
		v = v.Elem()
	}
	if v.Type().Elem().Kind() == reflect.Uint8 {
		return bytesToValue(v), nil
	}
	protoValue := new(proto.Value)
	protoValue.TestValue = &proto.Value_A{
		A: &proto.ArrayValue{},
	}
	if v.Len() == 0 {
		return protoValue, nil
	}
	arr := protoValue.GetA()
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
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return &proto.Value{TestValue: &proto.Value_O{}}, nil
		}
		v = v.Elem()
	}
	protoValue := new(proto.Value)
	if v.Type().Key().Kind() != reflect.String {
		return nil, fmt.Errorf("map key must be of string type")
	}
	protoValue.TestValue = &proto.Value_O{
		O: &proto.ObjectValue{},
	}
	if v.Len() == 0 {
		return protoValue, nil
	}
	obj := protoValue.GetO()
	obj.Props = make(map[string]*proto.Value)
	for _, k := range v.MapKeys() {
		if v, err := anyToValueReflected(v.MapIndex(k)); err == nil {
			if v == nil {
				v = new(proto.Value)
			}
			obj.Props[k.String()] = v
		} else if err != nil {
			return nil, err
		}
	}
	return protoValue, nil
}

// ValueMarshaler interface for client types that can return it's own proto.Value
type ValueMarshaler interface {
	MarshalValue() (*proto.Value, error)
}

var marshalerInterface = reflect.TypeOf((*ValueMarshaler)(nil)).Elem()

func anyToValueReflected(v reflect.Value) (*proto.Value, error) {
	// short path for ValueMarshaler interface
	if v.Type().Implements(marshalerInterface) {
		return v.Interface().(ValueMarshaler).MarshalValue()
	}
	t := v.Type()
	if t.Kind() == reflect.Interface {
		if v.IsNil() {
			// empty value
			return new(proto.Value), nil
		}
		v = v.Elem()
		t = v.Type()
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
		var err error
		protoValue, err = sliceOrArrayToValue(v)
		if err != nil {
			return nil, err
		}
	case reflect.Map:
		var err error
		protoValue, err = mapToValue(v)
		if err != nil {
			return nil, err
		}
	case reflect.Struct:
		var err error
		protoValue, err = structToValue(v)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("kind %s not supported", v.Kind())
	}
	return protoValue, nil
}

func anyToValue(v interface{}) (*proto.Value, error) {
	return anyToValueReflected(reflect.ValueOf(flattenValue(v)))
}

func mapOfValueToMapOfAny(m map[string]*proto.Value) (many map[string]interface{}, err error) {
	if len(m) == 0 {
		return
	}
	many = make(map[string]interface{}, len(m))
	for k, v := range m {
		many[k], err = valueToAny(v)
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

func valueToAny(pv *proto.Value) (v interface{}, err error) {
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
		if tv.A != nil && len(tv.A.GetItems()) > 0 {
			arr = make([]interface{}, 0, len(tv.A.GetItems()))
			for _, av := range tv.A.GetItems() {
				v, err := valueToAny(av)
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
		if tv.O != nil && len(tv.O.GetProps()) > 0 {
			m = make(map[string]interface{}, len(tv.O.GetProps()))
			for k, v := range tv.O.GetProps() {
				prop, err := valueToAny(v)
				if err != nil {
					return nil, err
				}
				if prop != nil {
					m[k] = prop
				}
			}
		}
		v = m
	case *proto.Value_Any:
		v = tv.Any
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
	dv, err := valueToAny(v.GetDefaultValue())
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

func makeDriverDirective(v *proto.Directive) (d types.Directive, err error) {
	args, err := mapOfValueToMapOfAny(v.GetArguments())
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

func makeDriverDirectives(v []*proto.Directive) (dd types.Directives, err error) {
	for _, dir := range v {
		var tdir types.Directive
		tdir, err = makeDriverDirective(dir)
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

func makeDriverFragmentDefinition(v *proto.FragmentDefinition) (fd *types.FragmentDefinition, err error) {
	if v == nil {
		return
	}
	dirs, err := makeDriverDirectives(v.GetDirectives())
	if err != nil {
		return
	}
	variableDefinitions, err := makeDriverVariableDefinitions(v.GetVariableDefinitions())
	if err != nil {
		return
	}
	selectionSet, err := makeDriverSelectionSet(v.GetSelectionSet())
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

func makeDriverSelection(v *proto.Selection) (s types.Selection, err error) {
	args, err := mapOfValueToMapOfAny(v.GetArguments())
	if err != nil {
		return
	}
	dirs, err := makeDriverDirectives(v.GetDirectives())
	if err != nil {
		return
	}
	fd, err := makeDriverFragmentDefinition(v.GetDefinition())
	if err != nil {
		return
	}
	selectionSet, err := makeDriverSelectionSet(v.GetSelectionSet())
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

func makeDriverSelectionSet(v []*proto.Selection) (ss types.Selections, err error) {
	for _, sel := range v {
		var s types.Selection
		s, err = makeDriverSelection(sel)
		if err != nil {
			return
		}
		ss = append(ss, s)
	}
	return
}

func makeProtoOperationDefinition(v *types.OperationDefinition) (o *proto.OperationDefinition, err error) {
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

func makeDriverOperationDefinition(v *proto.OperationDefinition) (od *types.OperationDefinition, err error) {
	variableDefinitions, err := makeDriverVariableDefinitions(v.GetVariableDefinitions())
	if err != nil {
		return
	}
	dirs, err := makeDriverDirectives(v.GetDirectives())
	if err != nil {
		return
	}
	selectionSet, err := makeDriverSelectionSet(v.GetSelectionSet())
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

func makeProtoResponsePath(v *types.ResponsePath) *proto.ResponsePath {
	if v == nil {
		return nil
	}
	return &proto.ResponsePath{
		Key:  v.Key,
		Prev: makeProtoResponsePath(v.Prev),
	}
}

func makeDriverResponsePath(v *proto.ResponsePath) *types.ResponsePath {
	if v == nil {
		return nil
	}
	return &types.ResponsePath{
		Key:  v.GetKey(),
		Prev: makeDriverResponsePath(v.GetPrev()),
	}
}

func flattenValue(v interface{}) interface{} {
	switch astValue := v.(type) {
	case *ast.Variable:
		return astValue.Name.Value
	case *ast.IntValue, *ast.FloatValue, *ast.StringValue, *ast.BooleanValue, *ast.EnumValue:
		return astValue.(ast.Value).GetValue()
	case *ast.ListValue:
		arr := make([]interface{}, len(astValue.Values))
		for i := 0; i < len(arr); i++ {
			arr[i] = flattenValue(astValue.Values[i])
		}
		return arr
	case *ast.ObjectValue:
		obj := make(map[string]interface{})
		for _, f := range astValue.Fields {
			obj[f.Name.Value] = flattenValue(f.Value)
		}
		return obj
	case []*ast.ObjectField:
		// Handle list of object fields like a map
		obj := make(map[string]interface{})
		for _, f := range astValue {
			obj[f.Name.Value] = flattenValue(f.Value)
		}
		return obj
	case ast.Value:
		return astValue.GetValue()
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		arr := make([]interface{}, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			arr[i] = flattenValue(rv.Index(i).Interface())
		}
		return arr
	}
	return v
}
