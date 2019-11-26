package parser

import (
	"errors"

	"github.com/graphql-go/graphql"

	"github.com/graphql-go/graphql/language/ast"
)

func toGraphQLType(p *Parser, tt ast.Type) (gt graphql.Type, err error) {
	switch t := tt.(type) {
	case *ast.Named:
		n := t.Name.Value
		switch n {
		case graphql.Int.Name():
			gt = graphql.Int
		case graphql.Float.Name():
			gt = graphql.Float
		case graphql.String.Name():
			gt = graphql.String
		case graphql.Boolean.Name():
			gt = graphql.Boolean
		case graphql.ID.Name():
			gt = graphql.ID
		default:
			d, ok := p.definitions[n]
			if !ok {
				err = errors.New("undefined type " + n)
				break
			}
			gt, err = customDefinition(p, d)
		}
	case *ast.NonNull:
		gt, err = toGraphQLType(p, t.Type)
		if err != nil {
			break
		}
		gt = graphql.NewNonNull(gt)
	case *ast.List:
		gt, err = toGraphQLType(p, t.Type)
		if err != nil {
			break
		}
		gt = graphql.NewList(gt)
	default:
		err = errors.New("type not supported")
	}
	return
}
