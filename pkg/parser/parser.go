package parser

import (
	"github.com/pkg/errors"

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

func (p *Parser) schemaExtension(ext *ast.SchemaExtensionDefinition) error {
	var def *ast.SchemaDefinition
	for i := 0; def == nil && i < len(p.document.Definitions); i++ {
		var ok bool
		if def, ok = p.document.Definitions[i].(*ast.SchemaDefinition); !ok {
			def = nil
		}
	}
	if def == nil {
		return errors.New("found extension for schema, but schema is not defined")
	}
	for _, op := range ext.Definition.OperationTypes {
		for _, baseop := range def.OperationTypes {
			if op.Operation == baseop.Operation {
				return errors.Errorf("operation %s already defined in schema", op.Operation)
			}
		}
		def.OperationTypes = append(def.OperationTypes, op)
	}
	for _, dir := range ext.Definition.Directives {
		for _, basedir := range def.Directives {
			if dir.Name.Value == basedir.Name.Value {
				return errors.New("directive already present in definition of schema")
			}
		}
		def.Directives = append(def.Directives, dir)
	}
	return nil
}

func (p *Parser) scalarExtension(ext *ast.ScalarExtensionDefinition) error {
	var def *ast.ScalarDefinition
	for i := 0; def == nil && i < len(p.document.Definitions); i++ {
		var ok bool
		if def, ok = p.document.Definitions[i].(*ast.ScalarDefinition); !ok || def.Name.Value != ext.Definition.Name.Value {
			def = nil
		}
	}
	if def == nil {
		return errors.Errorf("found extension for type %s, which either does not exist or is not an scalar definition", ext.Definition.Name.Value)
	}
	for _, dir := range ext.Definition.Directives {
		for _, basedir := range def.Directives {
			if dir.Name.Value == basedir.Name.Value {
				return errors.Errorf("directive %s already present in definition of scalar %s", dir.Name.Value, def.Name.Value)
			}
		}
		def.Directives = append(def.Directives, dir)
	}
	return nil
}

func (p *Parser) objectExtension(ext *ast.ObjectExtensionDefinition) error {
	var def *ast.ObjectDefinition
	for i := 0; def == nil && i < len(p.document.Definitions); i++ {
		var ok bool
		if def, ok = p.document.Definitions[i].(*ast.ObjectDefinition); !ok || def.Name.Value != ext.Definition.Name.Value {
			def = nil
		}
	}
	if def == nil {
		return errors.Errorf("found extension for type %s, which either does not exist or is not an object definition", ext.Definition.Name.Value)
	}
	for _, dir := range ext.Definition.Directives {
		for _, basedir := range def.Directives {
			if dir.Name.Value == basedir.Name.Value {
				return errors.Errorf("directive %s already present in definition of object %s", dir.Name.Value, def.Name.Value)
			}
		}
		def.Directives = append(def.Directives, dir)
	}
	for _, field := range ext.Definition.Fields {
		for _, basefield := range def.Fields {
			if field.Name.Value == basefield.Name.Value {
				return errors.Errorf("field %s already present in definition of object %s", field.Name.Value, def.Name.Value)
			}
		}
		def.Fields = append(def.Fields, field)
	}
	for _, iface := range ext.Definition.Interfaces {
		for _, baseiface := range def.Interfaces {
			if iface.Name.Value == baseiface.Name.Value {
				return errors.Errorf("interface %s already present in definition of object %s", iface.Name.Value, def.Name.Value)
			}
		}
		def.Interfaces = append(def.Interfaces, iface)
	}
	return nil
}

func (p *Parser) interfaceExtension(ext *ast.InterfaceExtensionDefinition) error {
	var def *ast.InterfaceDefinition
	for i := 0; def == nil && i < len(p.document.Definitions); i++ {
		var ok bool
		if def, ok = p.document.Definitions[i].(*ast.InterfaceDefinition); !ok || def.Name.Value != ext.Definition.Name.Value {
			def = nil
		}
	}
	if def == nil {
		return errors.Errorf("found extension for type %s, which either does not exist or is not an interface definition", ext.Definition.Name.Value)
	}
	for _, dir := range ext.Definition.Directives {
		for _, basedir := range def.Directives {
			if dir.Name.Value == basedir.Name.Value {
				return errors.Errorf("directive %s already present in definition of interface %s", dir.Name.Value, def.Name.Value)
			}
		}
		def.Directives = append(def.Directives, dir)
	}
	for _, field := range ext.Definition.Fields {
		for _, basefield := range def.Fields {
			if field.Name.Value == basefield.Name.Value {
				return errors.Errorf("field %s already present in definition of interface %s", field.Name.Value, def.Name.Value)
			}
		}
		def.Fields = append(def.Fields, field)
	}
	return nil
}

func (p *Parser) unionExtension(ext *ast.UnionExtensionDefinition) error {
	var def *ast.UnionDefinition
	for i := 0; def == nil && i < len(p.document.Definitions); i++ {
		var ok bool
		if def, ok = p.document.Definitions[i].(*ast.UnionDefinition); !ok || def.Name.Value != ext.Definition.Name.Value {
			def = nil
		}
	}
	if def == nil {
		return errors.Errorf("found extension for type %s, which either does not exist or is not an union definition", ext.Definition.Name.Value)
	}
	for _, dir := range ext.Definition.Directives {
		for _, basedir := range def.Directives {
			if dir.Name.Value == basedir.Name.Value {
				return errors.Errorf("directive %s already present in definition of union %s", dir.Name.Value, def.Name.Value)
			}
		}
		def.Directives = append(def.Directives, dir)
	}
	for _, tp := range ext.Definition.Types {
		for _, basetp := range def.Types {
			if tp.Name.Value == basetp.Name.Value {
				return errors.Errorf("type %s already present in definition of type %s", tp.Name.Value, def.Name.Value)
			}
		}
		def.Types = append(def.Types, tp)
	}
	return nil
}

func (p *Parser) enumExtension(ext *ast.EnumExtensionDefinition) error {
	var def *ast.EnumDefinition
	for i := 0; def == nil && i < len(p.document.Definitions); i++ {
		var ok bool
		if def, ok = p.document.Definitions[i].(*ast.EnumDefinition); !ok || def.Name.Value != ext.Definition.Name.Value {
			def = nil
		}
	}
	if def == nil {
		return errors.Errorf("found extension for type %s, which either does not exist or is not an enum definition", ext.Definition.Name.Value)
	}
	for _, dir := range ext.Definition.Directives {
		for _, basedir := range def.Directives {
			if dir.Name.Value == basedir.Name.Value {
				return errors.Errorf("directive %s already present in definition of union %s", dir.Name.Value, def.Name.Value)
			}
		}
		def.Directives = append(def.Directives, dir)
	}
	for _, val := range ext.Definition.Values {
		for _, baseval := range def.Values {
			if val.Name.Value == baseval.Name.Value {
				return errors.Errorf("value %s already present in definition of type %s", val.Name.Value, def.Name.Value)
			}
		}
		def.Values = append(def.Values, val)
	}
	return nil
}

func (p *Parser) inputObjectExtension(ext *ast.InputObjectExtensionDefinition) error {
	var def *ast.InputObjectDefinition
	for i := 0; def == nil && i < len(p.document.Definitions); i++ {
		var ok bool
		if def, ok = p.document.Definitions[i].(*ast.InputObjectDefinition); !ok || def.Name.Value != ext.Definition.Name.Value {
			def = nil
		}
	}
	if def == nil {
		return errors.Errorf("found extension for type %s, which either does not exist or is not an input object definition", ext.Definition.Name.Value)
	}
	for _, dir := range ext.Definition.Directives {
		for _, basedir := range def.Directives {
			if dir.Name.Value == basedir.Name.Value {
				return errors.Errorf("directive %s already present in definition of union %s", dir.Name.Value, def.Name.Value)
			}
		}
		def.Directives = append(def.Directives, dir)
	}
	for _, field := range ext.Definition.Fields {
		for _, basefield := range def.Fields {
			if field.Name.Value == basefield.Name.Value {
				return errors.Errorf("field %s already present in definition of type %s", field.Name.Value, def.Name.Value)
			}
		}
		def.Fields = append(def.Fields, field)
	}
	return nil
}

func (p *Parser) mergeExtensions() (err error) {
	for i := 0; err == nil && i < len(p.document.Definitions); i++ {
		switch v := p.document.Definitions[i].(type) {
		case *ast.SchemaExtensionDefinition:
			err = p.schemaExtension(v)
		case *ast.ScalarExtensionDefinition:
			err = p.scalarExtension(v)
		case *ast.ObjectExtensionDefinition:
			err = p.objectExtension(v)
		case *ast.InterfaceExtensionDefinition:
			err = p.interfaceExtension(v)
		case *ast.UnionExtensionDefinition:
			err = p.unionExtension(v)
		case *ast.EnumExtensionDefinition:
			err = p.enumExtension(v)
		case *ast.InputObjectExtensionDefinition:
			err = p.inputObjectExtension(v)
		}
	}
	return
}

func (p *Parser) analyzeDocument() (err error) {
	// first pass, merge extensions
	if err = p.mergeExtensions(); err == nil {
		// second pass, gather definitions
		for i := 0; err == nil && i < len(p.document.Definitions); i++ {
			switch v := p.document.Definitions[i].(type) {
			case *ast.SchemaDefinition:
				p.schema = analyzeSchema(v)
			case namedDefinition:
				err = p.addDefinition(v.GetName().Value, v)
			}
			if err != nil {
				return
			}
		}
	}
	return
}

func (p *Parser) analyze() (graphql.Schema, error) {
	if err := p.analyzeDocument(); err != nil {
		return graphql.Schema{}, err
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
