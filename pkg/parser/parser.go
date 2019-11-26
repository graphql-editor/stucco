package parser

import (
	"errors"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"
)

type definitions map[string]ast.Definition

// Parser binds defined functions with fields/types defined
// in schema source
type Parser struct {
	Config
	document    *ast.Document
	schema      schema
	definitions definitions
	gqlTypeMap  graphql.TypeMap
}

func (p *Parser) addDefinition(n string, d ast.Definition) error {
	if p.definitions == nil {
		p.definitions = make(definitions)
	}
	if _, ok := p.definitions[n]; ok {
		return errors.New("more than one definition of type " + n)
	}
	p.definitions[n] = d
	return nil
}

type namedDefinition interface {
	ast.Definition
	GetName() *ast.Name
}

func (p *Parser) analyzeDocument() (err error) {
	for _, n := range p.document.Definitions {
		switch v := n.(type) {
		case *ast.SchemaDefinition:
			p.schema = analyzeSchema(v)
		case namedDefinition:
			err = p.addDefinition(v.GetName().Value, v)
		}
		if err != nil {
			return
		}
	}
	return
}

func (p *Parser) analyze() (graphql.Schema, error) {
	if err := p.analyzeDocument(); err != nil {
		return graphql.Schema{}, nil
	}
	return p.schema.parse(p)
}

// Parse schema source using parser
func (p *Parser) Parse(source string) (graphql.Schema, error) {
	var err error
	p.document, err = parser.Parse(parser.ParseParams{
		Source: source,
	})
	if err != nil {
		p.document = nil
		return graphql.Schema{}, err
	}
	return p.analyze()
}

// ScalarFunctions definitions
type ScalarFunctions struct {
	// Parse scalar function definition
	Parse graphql.ParseValueFn
	// Serialize scalar function definition
	Serialize graphql.SerializeFn
}

// Config definies implementation of functions to be called
// that parser will bind with types/fields of schema.
type Config struct {
	Interfaces map[string]graphql.ResolveTypeFn
	Resolvers  map[string]graphql.FieldResolveFn
	Scalars    map[string]ScalarFunctions
	Unions     map[string]graphql.ResolveTypeFn
}

// NewParser creates a schema Parser with a config
func NewParser(c Config) Parser {
	return Parser{
		Config: c,
	}
}
