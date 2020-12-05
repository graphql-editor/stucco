package parser

import (
	"errors"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
)

func getGraphQLObjectDefinition(p *Parser, odef *ast.ObjectDefinition) (o *graphql.Object, err error) {
	var ot graphql.Type
	ot, err = customDefinition(p, odef)
	if err == nil {
		var ok bool
		o, ok = ot.(*graphql.Object)
		if !ok {
			err = errors.New("only object can implement interface")
		}
	}
	return
}

func getObjectsImplementingInterface(p *Parser, name string) (objects []*graphql.Object, err error) {
	for _, definition := range p.definitions {
		odef, ok := definition.(*ast.ObjectDefinition)
		if !ok {
			continue
		}
		var i *ast.Named
		interfaces := odef.Interfaces
		for i == nil && len(interfaces) > 0 {
			if interfaces[0].Name.Value == name {
				i = interfaces[0]
			}
			interfaces = interfaces[1:]
		}
		if i == nil {
			continue
		}
		var gqlObj *graphql.Object
		gqlObj, err = getGraphQLObjectDefinition(p, odef)
		if err != nil {
			return
		}
		objects = append(objects, gqlObj)
	}
	return
}

func getInterfaceResolveTypeFunction(p *Parser, name string) (fn graphql.ResolveTypeFn, err error) {
	var ok bool
	if fn, ok = p.Interfaces[name]; !ok {
		var types []*graphql.Object
		types, err = getObjectsImplementingInterface(p, name)
		if err != nil {
			return
		}
		fn = defaultResolveFunc(types)
	} else {
		_, err = getObjectsImplementingInterface(p, name)
	}
	return
}

func interfaceDefinition(p *Parser, i *ast.InterfaceDefinition) (t *graphql.Interface, err error) {
	iCfg := graphql.InterfaceConfig{
		Name: i.Name.Value,
	}
	setDescription(&iCfg.Description, i)
	t = graphql.NewInterface(iCfg)
	p.gqlTypeMap[t.Name()] = t
	fields := graphql.Fields{}
	for _, f := range i.Fields {
		var field *graphql.Field
		field, err = makeField(p, i.Name.Value, f)
		if err != nil {
			return
		}
		fields[f.Name.Value] = field
	}
	if iCfg.ResolveType, err = getInterfaceResolveTypeFunction(p, i.Name.Value); err != nil {
		return
	}
	iCfg.Fields = fields
	*t = *graphql.NewInterface(iCfg)
	return
}
