package parser

import (
	"errors"

	"github.com/graphql-go/graphql"

	"github.com/graphql-go/graphql/language/ast"
)

type rootOperation ast.OperationTypeDefinition

func (r rootOperation) config(p *Parser) (o *graphql.Object, err error) {
	switch v := p.definitions[r.Type.Name.Value].(type) {
	case *ast.ObjectDefinition:
		return objectDefintion(p, v)
	default:
		err = errors.New("root operation must be an object")
	}
	return
}
