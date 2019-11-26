package parser

import (
	"errors"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
)

type schema struct {
	Query        *rootOperation
	Mutation     *rootOperation
	Subscription *rootOperation
}

func analyzeSchema(schemaNode *ast.SchemaDefinition) schema {
	var s schema
	for _, op := range schemaNode.OperationTypes {
		switch op.Operation {
		case "query":
			s.Query = (*rootOperation)(op)
		case "mutation":
			s.Mutation = (*rootOperation)(op)
		case "subscription":
			s.Subscription = (*rootOperation)(op)
		}
	}
	return s
}

func (s schema) parse(p *Parser) (graphql.Schema, error) {
	if p.gqlTypeMap == nil {
		p.gqlTypeMap = graphql.TypeMap{}
	}
	if s.Query == nil {
		return graphql.Schema{}, errors.New("schema is missing root query")
	}
	o, err := s.Query.config(p)
	if err != nil {
		return graphql.Schema{}, err
	}
	sCfg := graphql.SchemaConfig{
		Query: o,
	}
	if s.Mutation != nil {
		o, err = s.Mutation.config(p)
		if err != nil {
			return graphql.Schema{}, err
		}
		sCfg.Mutation = o
	}
	if s.Subscription != nil {
		o, err = s.Subscription.config(p)
		if err != nil {
			return graphql.Schema{}, err
		}
		sCfg.Subscription = o
	}
	sCfg.Types = make([]graphql.Type, 0, len(p.gqlTypeMap))
	for _, t := range p.gqlTypeMap {
		sCfg.Types = append(sCfg.Types, t)
	}
	return graphql.NewSchema(sCfg)
}
