package parser

import (
	"github.com/graphql-go/graphql/language/ast"
)

func setDescription(d *string, descNode ast.DescribableNode) {
	if descNode.GetDescription() != nil {
		*d = descNode.GetDescription().Value
	}
}
