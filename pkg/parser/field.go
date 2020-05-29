package parser

import (
	"errors"

	"github.com/graphql-editor/stucco/pkg/utils"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
)

func makeFieldArgs(p *Parser, fd *ast.FieldDefinition) (args graphql.FieldConfigArgument, err error) {
	if len(fd.Arguments) != 0 {
		args = make(graphql.FieldConfigArgument)
		for i := 0; err == nil && i < len(fd.Arguments); i++ {
			arg := fd.Arguments[i]
			var t graphql.Type
			if t, err = toGraphQLType(p, arg.Type); err == nil {
				args[arg.Name.Value] = &graphql.ArgumentConfig{
					Type: t,
				}
				if arg.DefaultValue != nil {
					args[arg.Name.Value].DefaultValue = arg.DefaultValue.GetValue()
				}
				setDescription(&args[arg.Name.Value].Description, arg)
			}
		}
	}
	return
}

func makeField(p *Parser, tn string, fd *ast.FieldDefinition) (field *graphql.Field, err error) {
	astType, ok := fd.Type.(ast.Type)
	if !ok {
		err = errors.New("could not find field type definition")
	}
	if err == nil {
		var t graphql.Type
		t, err = toGraphQLType(p, astType)
		if err == nil {
			var args graphql.FieldConfigArgument
			args, err = makeFieldArgs(p, fd)
			if err == nil {
				field = &graphql.Field{
					Name: fd.Name.Value,
					Args: args,
					Type: t,
				}
				if fn, ok := p.Resolvers[utils.FieldName(tn, fd.Name.Value)]; ok {
					field.Resolve = fn
				}
				setDescription(&field.Description, fd)
			}
		}
	}
	return
}

func makeInputObjectField(p *Parser, fd *ast.InputValueDefinition) (field *graphql.InputObjectFieldConfig, err error) {
	astType, ok := fd.Type.(ast.Type)
	if !ok {
		err = errors.New("could not find field type definition")
	}
	if err == nil {
		var t graphql.Type
		if t, err = toGraphQLType(p, astType); err == nil {
			field = &graphql.InputObjectFieldConfig{Type: t}
			setDescription(&field.Description, fd)
		}
	}
	return
}
