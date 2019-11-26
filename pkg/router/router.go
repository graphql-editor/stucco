package router

import (
	"errors"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-editor/stucco/pkg/driver"
	"github.com/graphql-editor/stucco/pkg/parser"
	"github.com/graphql-editor/stucco/pkg/types"
)

type Router struct {
	Interfaces map[string]InterfaceConfig // Interfaces is a map of FaaS function configs used in determining concrete type of an interface
	Resolvers  map[string]ResolverConfig  // Resolvers is a map of FaaS function configs used in resolution
	Scalars    map[string]ScalarConfig    // Scalars is a map of FaaS function configs used in parsing and serializing custom scalars
	Unions     map[string]UnionConfig     // Unions is a map of FaaS function configs used in determining concrete type of an union
	Schema     graphql.Schema             // Parsed schema
	Secrets    SecretsConfig              // Secrets is a map of secret references
}

func assertTypeRef(t *types.TypeRef) types.TypeRef {
	if t == nil {
		panic("assertion failed, TypeRef cannot be null here")
	}
	return *t
}

func makeTypeRefFromNamed(named *ast.Named) *types.TypeRef {
	if named == nil {
		return nil
	}
	return &types.TypeRef{Name: named.Name.Value}
}

func mustMakeTypeRefFromNamed(named *ast.Named) types.TypeRef {
	return assertTypeRef(makeTypeRefFromNamed(named))
}

func makeTypeRefFromType(t graphql.Type) *types.TypeRef {
	if t == nil {
		return nil
	}
	switch tt := t.(type) {
	case *graphql.NonNull:
		return &types.TypeRef{
			NonNull: makeTypeRefFromType(tt.OfType),
		}
	case *graphql.List:
		return &types.TypeRef{
			List: makeTypeRefFromType(tt.OfType),
		}
	}
	return &types.TypeRef{Name: t.Name()}
}

func mustMakeTypeRefFromType(t graphql.Type) types.TypeRef {
	return assertTypeRef(makeTypeRefFromType(t))
}

func newReponsePath(p *graphql.ResponsePath) (*types.ResponsePath, error) {
	if p == nil || p.Key == nil {
		return nil, nil
	}
	k, ok := p.Key.(string)
	if !ok {
		return nil, errors.New("could not evaluate response path")
	}
	prev, err := newReponsePath(p.Prev)
	if err != nil {
		return nil, err
	}
	return &types.ResponsePath{
		Prev: prev,
		Key:  k,
	}, nil
}

func makeArguments(args []*ast.Argument) types.Arguments {
	if len(args) == 0 {
		return nil
	}
	o := make(types.Arguments)
	for _, arg := range args {
		if v := arg.Value.GetValue(); v != nil {
			switch vt := v.(type) {
			case []*ast.ObjectField:
				o[arg.Name.Value] = flattenArgObjectField(vt)
			case ast.Value:
				o[arg.Name.Value] = flattenArgValue(vt)
			default:
				o[arg.Name.Value] = v
			}
		}
	}
	return o
}

func makeVariableDefinition(v *ast.VariableDefinition) types.VariableDefinition {
	if v == nil {
		return types.VariableDefinition{}
	}
	return types.VariableDefinition{
		Variable: types.Variable{
			Name: v.Variable.Name.Value,
		},
		DefaultValue: v.DefaultValue.GetValue(),
	}
}

func makeVariableDefintions(v []*ast.VariableDefinition) []types.VariableDefinition {
	if len(v) == 0 {
		return nil
	}
	r := make([]types.VariableDefinition, 0, len(v))
	for _, vv := range v {
		r = append(r, makeVariableDefinition(vv))
	}
	return r
}

func makeDirective(dir *ast.Directive) (d types.Directive) {
	if dir == nil {
		return
	}
	d.Name = dir.Name.Value
	d.Arguments = makeArguments(dir.Arguments)
	return
}

func makeDirectives(dirs []*ast.Directive) types.Directives {
	if len(dirs) == 0 {
		return nil
	}
	o := make(types.Directives, 0, len(dirs))
	for _, dir := range dirs {
		o = append(o, makeDirective(dir))
	}
	return o
}

func makeSelection(sel ast.Selection, fragments map[string]ast.Definition) (s types.Selection) {
	switch st := sel.(type) {
	case *ast.Field:
		s = types.Selection{
			Name:         st.Name.Value,
			Arguments:    makeArguments(st.Arguments),
			Directives:   makeDirectives(st.Directives),
			SelectionSet: makeSelections(st.SelectionSet, fragments),
		}
	case *ast.FragmentSpread:
		fdef := fragments[st.Name.Value].(*ast.FragmentDefinition)
		s = types.Selection{
			Directives: makeDirectives(st.Directives),
			Definition: &types.FragmentDefinition{
				TypeCondition:       mustMakeTypeRefFromNamed(fdef.TypeCondition),
				SelectionSet:        makeSelections(fdef.SelectionSet, fragments),
				Directives:          makeDirectives(fdef.Directives),
				VariableDefinitions: makeVariableDefintions(fdef.VariableDefinitions),
			},
		}
	case *ast.InlineFragment:
		s = types.Selection{
			Directives: makeDirectives(st.Directives),
			Definition: &types.FragmentDefinition{
				TypeCondition: mustMakeTypeRefFromNamed(st.TypeCondition),
				SelectionSet:  makeSelections(st.SelectionSet, fragments),
			},
		}
	}
	return
}

func makeSelections(selectionSet *ast.SelectionSet, fragments map[string]ast.Definition) types.Selections {
	if selectionSet == nil {
		return nil
	}
	selections := make(types.Selections, 0, len(selectionSet.Selections))
	for _, sel := range selectionSet.Selections {
		selections = append(selections, makeSelection(sel, fragments))
	}
	return selections
}

func makeOperationDefinition(odef *ast.OperationDefinition, fragments map[string]ast.Definition) *types.OperationDefinition {
	if odef == nil {
		return nil
	}
	name := ""
	if odef.Name != nil {
		name = odef.Name.Value
	}
	return &types.OperationDefinition{
		Operation:           odef.Operation,
		Name:                name,
		Directives:          makeDirectives(odef.Directives),
		VariableDefinitions: makeVariableDefintions(odef.VariableDefinitions),
		SelectionSet:        makeSelections(odef.SelectionSet, fragments),
	}
}

func buildInterfaceInfoParams(params graphql.ResolveInfo) (info driver.InterfaceResolveTypeInfo, err error) {
	info = driver.InterfaceResolveTypeInfo{
		FieldName:      params.FieldName,
		ReturnType:     makeTypeRefFromType(params.ReturnType),
		ParentType:     makeTypeRefFromType(params.ParentType),
		VariableValues: params.VariableValues,
	}
	path, err := newReponsePath(params.Path)
	if err != nil {
		return
	}
	info.Path = path
	odef, ok := params.Operation.(*ast.OperationDefinition)
	if ok {
		info.Operation = makeOperationDefinition(odef, params.Fragments)
	}
	return
}

func (r *Router) bindInterfaces(c *parser.Config) error {
	for k, i := range r.Interfaces {
		dri, err := r.getDriver(driver.Config{
			Provider: i.Environment.Provider,
			Runtime:  i.Environment.Runtime,
		})
		if err != nil {
			return err
		}
		r.setDriverSecrets(dri)
		c.Interfaces[k] = driDispatch{
			Driver: dri,
			r:      r,
		}.InterfaceResolveType(i)
	}
	return nil
}

func buildFieldInfoParams(params graphql.ResolveInfo) (info driver.FieldResolveInfo, err error) {
	info = driver.FieldResolveInfo{
		FieldName:      params.FieldName,
		ReturnType:     makeTypeRefFromType(params.ReturnType),
		ParentType:     makeTypeRefFromType(params.ParentType),
		VariableValues: params.VariableValues,
	}
	path, err := newReponsePath(params.Path)
	if err != nil {
		return
	}
	info.Path = path
	odef, ok := params.Operation.(*ast.OperationDefinition)
	if ok {
		info.Operation = makeOperationDefinition(odef, params.Fragments)
	}
	return
}

func (r *Router) bindResolvers(c *parser.Config) error {
	for k, rs := range r.Resolvers {
		dri, err := r.getDriver(driver.Config{
			Provider: rs.Environment.Provider,
			Runtime:  rs.Environment.Runtime,
		})
		if err != nil {
			return err
		}
		c.Resolvers[k] = driDispatch{dri, r}.FieldResolve(rs)
	}
	return nil
}

func (r *Router) bindScalars(c *parser.Config) error {
	for k, s := range r.Scalars {
		dri, err := r.getDriver(driver.Config{
			Provider: s.Environment.Provider,
			Runtime:  s.Environment.Runtime,
		})
		if err != nil {
			return err
		}
		c.Scalars[k] = driDispatch{
			Driver: dri,
			r:      r,
		}.ScalarFunctions(s)
	}
	return nil
}

func buildUnionInfoParams(params graphql.ResolveInfo) (info driver.UnionResolveTypeInfo, err error) {
	info = driver.UnionResolveTypeInfo{
		FieldName:      params.FieldName,
		ReturnType:     makeTypeRefFromType(params.ReturnType),
		ParentType:     makeTypeRefFromType(params.ParentType),
		VariableValues: params.VariableValues,
	}
	path, err := newReponsePath(params.Path)
	if err != nil {
		return
	}
	info.Path = path
	odef, ok := params.Operation.(*ast.OperationDefinition)
	if ok {
		info.Operation = makeOperationDefinition(odef, params.Fragments)
	}
	return
}

func (r *Router) bindUnions(c *parser.Config) error {
	for k, u := range r.Unions {
		dri, err := r.getDriver(driver.Config{
			Provider: u.Environment.Provider,
			Runtime:  u.Environment.Runtime,
		})
		if err != nil {
			return err
		}
		c.Unions[k] = driDispatch{
			Driver: dri,
			r:      r,
		}.UnionResolveType(u)
	}
	return nil
}

func (r *Router) parserConfig() (parser.Config, error) {
	c := parser.Config{
		Interfaces: make(map[string]graphql.ResolveTypeFn, len(r.Interfaces)),
		Resolvers:  make(map[string]graphql.FieldResolveFn, len(r.Resolvers)),
		Scalars:    make(map[string]parser.ScalarFunctions, len(r.Scalars)),
		Unions:     make(map[string]graphql.ResolveTypeFn, len(r.Unions)),
	}
	for _, f := range []func(c *parser.Config) error{
		r.bindInterfaces,
		r.bindResolvers,
		r.bindScalars,
		r.bindUnions,
	} {
		if err := f(&c); err != nil {
			return parser.Config{}, err
		}
	}
	return c, nil
}

func (r *Router) parseSchema(c Config) error {
	source, err := c.rawSchema()
	if err != nil {
		return err
	}
	pConfig, err := r.parserConfig()
	if err != nil {
		return err
	}
	p := parser.NewParser(pConfig)
	schema, err := p.Parse(source)
	if err != nil {
		return err
	}
	r.Schema = schema
	return nil
}

func (r *Router) setDriverSecrets(dri driver.Driver) error {
	secrets, err := dri.SetSecrets(driver.SetSecretsInput{
		Secrets: r.Secrets.Secrets,
	})
	if err == nil && secrets.Error != nil {
		err = errors.New(secrets.Error.Message)
	}
	return err
}

func (r *Router) getDriver(cfg driver.Config) (dri driver.Driver, err error) {
	dri = driver.GetDriver(cfg)
	if dri == nil {
		err = errors.New("driver not found")
		return
	}
	err = r.setDriverSecrets(dri)
	return
}

func (r *Router) load(c Config) error {
	r.Secrets = c.Secrets
	for k, i := range c.Interfaces {
		i.Environment = newEnvironment(i.Environment, c.Environment)
		r.Interfaces[k] = i
	}
	for k, rs := range c.Resolvers {
		rs.Environment = newEnvironment(rs.Environment, c.Environment)
		r.Resolvers[k] = rs
	}
	for k, s := range c.Scalars {
		s.Environment = newEnvironment(s.Environment, c.Environment)
		r.Scalars[k] = s
	}
	for k, u := range c.Unions {
		u.Environment = newEnvironment(u.Environment, c.Environment)
		r.Unions[k] = u
	}
	return r.parseSchema(c)
}

func NewRouter(c Config) (Router, error) {
	// Load local-nodejs environment by default
	if c.RuntimeEnvironment == nil {
		c.RuntimeEnvironment = &defaultEnvironment
	}
	c.RuntimeEnvironment.Merge(defaultEnvironment)
	c.Environment.Merge(*c.RuntimeEnvironment)
	r := Router{
		Interfaces: make(map[string]InterfaceConfig, len(c.Interfaces)),
		Resolvers:  make(map[string]ResolverConfig, len(c.Resolvers)),
		Scalars:    make(map[string]ScalarConfig, len(c.Scalars)),
		Unions:     make(map[string]UnionConfig, len(c.Unions)),
	}
	err := r.load(c)
	return r, err
}
