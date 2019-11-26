package parser

import (
	"errors"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
)

func objectDefintion(p *Parser, o *ast.ObjectDefinition) (t *graphql.Object, err error) {
	oCfg := graphql.ObjectConfig{
		Name: o.Name.Value,
	}
	setDescription(&oCfg.Description, o)
	t = graphql.NewObject(oCfg)
	p.gqlTypeMap[t.Name()] = t
	fields := graphql.Fields{}
	for _, f := range o.Fields {
		var field *graphql.Field
		field, err = makeField(p, o.Name.Value, f)
		if err != nil {
			return
		}
		fields[f.Name.Value] = field
	}
	var interfaces []*graphql.Interface
	for _, definition := range p.definitions {
		if idef, ok := definition.(*ast.InterfaceDefinition); ok {
			for _, iface := range o.Interfaces {
				if idef.Name.Value == iface.Name.Value {
					it, err := customDefinition(p, idef)
					if err != nil {
						return nil, err
					}
					gqlIface, ok := it.(*graphql.Interface)
					if !ok {
						return nil, errors.New("object can only implement interface")
					}
					interfaces = append(interfaces, gqlIface)
				}
			}
		}
	}
	oCfg.Fields = fields
	oCfg.Interfaces = interfaces
	*t = *graphql.NewObject(oCfg)
	return
}
