package parser

import (
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
)

func enumDefinition(p *Parser, e *ast.EnumDefinition) (t *graphql.Enum, err error) {
	eCfg := graphql.EnumConfig{
		Name: e.Name.Value,
	}
	setDescription(&eCfg.Description, e)
	t = graphql.NewEnum(eCfg)
	p.gqlTypeMap[t.Name()] = t
	for _, v := range e.Values {
		if eCfg.Values == nil {
			eCfg.Values = make(graphql.EnumValueConfigMap, len(e.GetVariableDefinitions()))
		}
		eCfg.Values[v.Name.Value] = &graphql.EnumValueConfig{
			Value: v.Name.Value,
		}
		if v.Description != nil {
			eCfg.Values[v.Name.Value].Description = v.Description.Value
		}
	}
	*t = *graphql.NewEnum(eCfg)
	return t, nil
}
