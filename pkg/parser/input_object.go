package parser

import (
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
)

func inputObjectDefinition(p *Parser, i *ast.InputObjectDefinition) (t *graphql.InputObject, err error) {
	iCfg := graphql.InputObjectConfig{
		Name: i.Name.Value,
	}
	setDescription(&iCfg.Description, i)
	t = graphql.NewInputObject(iCfg)
	p.gqlTypeMap[t.Name()] = t
	fields := graphql.InputObjectConfigFieldMap{}
	for _, f := range i.Fields {
		var field *graphql.InputObjectFieldConfig
		field, err = makeInputObjectField(p, f)
		if err != nil {
			return
		}
		fields[f.Name.Value] = field
	}
	iCfg.Fields = fields
	*t = *graphql.NewInputObject(iCfg)
	return
}
