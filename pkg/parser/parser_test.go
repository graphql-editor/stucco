package parser

//import (
//	"testing"
//
//	"github.com/graphql-go/graphql"
//	"github.com/stretchr/testify/assert"
//	"github.com/stretchr/testify/require"
//)
//
//var (
//	catId = &graphql.Field{
//		Type: graphql.NewNonNull(graphql.ID),
//	}
//	catColor = &graphql.Field{
//		Type: graphql.String,
//	}
//	catIdIn = &graphql.InputObjectFieldConfig{
//		Type: graphql.NewNonNull(graphql.ID),
//	}
//	catColorIn = &graphql.InputObjectFieldConfig{
//		Type: graphql.String,
//	}
//	catReadInput = graphql.NewInputObject(graphql.InputObjectConfig{
//		Name: "CatReadInput",
//		Fields: graphql.InputObjectConfigFieldMap{
//			"id": catIdIn,
//		},
//	})
//	catCreateInput = graphql.NewInputObject(graphql.InputObjectConfig{
//		Name: "CatCreateInput",
//		Fields: graphql.InputObjectConfigFieldMap{
//			"color": catColorIn,
//		},
//	})
//	catUpdateInput = graphql.NewInputObject(graphql.InputObjectConfig{
//		Name: "CatUpdateInput",
//		Fields: graphql.InputObjectConfigFieldMap{
//			"id":    catIdIn,
//			"color": catColorIn,
//		},
//	})
//	catDeleteInput = graphql.NewInputObject(graphql.InputObjectConfig{
//		Name: "CatDeleteInput",
//		Fields: graphql.InputObjectConfigFieldMap{
//			"id": catIdIn,
//		},
//	})
//	cat = graphql.NewObject(graphql.ObjectConfig{
//		Name: "Cat",
//		Fields: graphql.Fields{
//			"id":    catId,
//			"color": catColor,
//		},
//	})
//	listCat = &graphql.Field{
//		Type: graphql.NewNonNull(graphql.NewList(cat)),
//	}
//	readCat = &graphql.Field{
//		Type: graphql.NewNonNull(cat),
//		Args: graphql.FieldConfigArgument{
//			"cat": &graphql.ArgumentConfig{
//				Type: graphql.NewNonNull(catReadInput),
//			},
//		},
//	}
//	query = graphql.NewObject(graphql.ObjectConfig{
//		Name: "Query",
//		Fields: graphql.Fields{
//			"listCat": listCat,
//			"readCat": readCat,
//		},
//	})
//	createCat = &graphql.Field{
//		Type: graphql.NewNonNull(cat),
//		Args: graphql.FieldConfigArgument{
//			"cat": &graphql.ArgumentConfig{
//				Type: graphql.NewNonNull(catCreateInput),
//			},
//		},
//	}
//	updateCat = &graphql.Field{
//		Type: graphql.NewNonNull(cat),
//		Args: graphql.FieldConfigArgument{
//			"cat": &graphql.ArgumentConfig{
//				Type: graphql.NewNonNull(catUpdateInput),
//			},
//		},
//	}
//	deleteCat = &graphql.Field{
//		Type: cat,
//		Args: graphql.FieldConfigArgument{
//			"cat": &graphql.ArgumentConfig{
//				Type: graphql.NewNonNull(catDeleteInput),
//			},
//		},
//	}
//	mutation = graphql.NewObject(graphql.ObjectConfig{
//		Name: "Mutation",
//		Fields: graphql.Fields{
//			"createCat": createCat,
//			"updateCat": updateCat,
//			"deleteCat": deleteCat,
//		},
//	})
//	catSchema, _ = graphql.NewSchema(graphql.SchemaConfig{
//		Query:    query,
//		Mutation: mutation,
//	})
//	catSource = `input CatReadInput {
//    id: ID!
//}
//
//input CatCreateInput {
//    color: String
//}
//
//input CatUpdateInput {
//    id: ID!
//    color: String
//}
//
//input CatDeleteInput {
//    id: ID!
//}
//
//type Cat {
//    id: ID!
//    color: String
//}
//
//type Query {
//    listCat: [Cat]!
//    readCat(cat: CatReadInput!): Cat!
//}
//
//type Mutation {
//    createCat(cat: CatCreateInput!): Cat!
//    updateCat(cat: CatUpdateInput!): Cat!
//    deleteCat(cat: CatDeleteInput!): Cat
//}
//
//schema {
//    query: Query
//    mutation: Mutation
//}`
//)
//
//func TestParse(t *testing.T) {
//	assert := assert.New(t)
//	data := []struct {
//		source string
//		out    graphql.Schema
//		err    ErrorAssertion
//	}{
//		{
//			source: catSource,
//			out:    catSchema,
//			err:    NoError(assert),
//		},
//	}
//
//	for _, tt := range data {
//		p := NewParser(Config{})
//		schema, err := p.Parse(tt.source)
//		tt.err(err)
//		assert.Len(schema.TypeMap(), len(tt.out.TypeMap()))
//		for k := range tt.out.TypeMap() {
//			assert.Contains(schema.TypeMap(), k)
//		}
//	}
//}
//
//type parserParseTestCase struct {
//	in  string
//	out graphql.SchemaConfig
//	err func(a *assert.Assertions) ErrorAssertion
//}
//
//func (e parserParseTestCase) test(t *testing.T) {
//	assert := assert.New(t)
//	require := require.New(t)
//	if e.err == nil {
//		e.err = NoError
//	}
//	expected, err := graphql.NewSchema(e.out)
//	require.NoError(err)
//	p := NewParser(Config{})
//	s, err := p.Parse(e.in)
//	e.err(assert)(err)
//	for k, t := range expected.TypeMap() {
//		assert.Contains(s.TypeMap(), k)
//		ts := s.TypeMap()[k]
//		assert.Equal(t.Name(), ts.Name())
//		assert.Equal(t.Description(), ts.Description())
//		assert.Equal(t.String(), ts.String())
//		switch tt := t.(type) {
//		case *graphql.Object:
//			tts, ok := ts.(*graphql.Object)
//			assert.True(ok, "graphql type mismatch")
//			if !ok {
//				continue
//			}
//			for fk, fv := range tt.Fields() {
//				assert.Contains(tts.Fields(), fk)
//				tf := tts.Fields()[fk]
//				assert.Equal(fv.Args, tf.Args)
//				assert.Equal(fv.DeprecationReason, tf.DeprecationReason)
//				assert.Equal(fv.Description, tf.Description)
//				assert.Equal(fv.Name, tf.Name)
//				assert.Equal(fv.Type.Name(), tf.Type.Name())
//			}
//		}
//	}
//}
//
//func TestParserParseTestCase(t *testing.T) {
//	data := []parserParseTestCase{
//		{
//			in: `enum Foo{
//    Fooo
//    Foooo
//    Fooooo
//}
//
//scalar Bar
//
//type UnionFoo {
//    foo: String
//}
//
//type UnionBar {
//    bar: String
//}
//
//union UnionFoobar = UnionFoo | UnionBar
//
//type Query {
//    foo: Foo
//    bar: Bar
//    foobar: UnionFoobar
//}
//
//schema {
//    query: Query
//}
//`,
//			out: graphql.SchemaConfig{
//				Query: graphql.NewObject(graphql.ObjectConfig{
//					Name: "Query",
//					Fields: graphql.Fields{
//						"foo": &graphql.Field{
//							Name: "foo",
//							Type: graphql.NewEnum(graphql.EnumConfig{
//								Name: "Foo",
//								Values: graphql.EnumValueConfigMap{
//									"Fooo": &graphql.EnumValueConfig{
//										Value: "Fooo",
//									},
//									"Foooo": &graphql.EnumValueConfig{
//										Value: "Foooo",
//									},
//									"Fooooo": &graphql.EnumValueConfig{
//										Value: "Fooooo",
//									},
//								},
//							}),
//						},
//						"bar": &graphql.Field{
//							Name: "bar",
//							Type: graphql.NewScalar(graphql.ScalarConfig{
//								Name: "Bar",
//								Serialize: func(interface{}) interface{} {
//									return nil
//								},
//							}),
//						},
//						"foobar": &graphql.Field{
//							Name: "foobar",
//							Type: graphql.NewUnion(graphql.UnionConfig{
//								Name: "UnionFoobar",
//								ResolveType: func(graphql.ResolveTypeParams) *graphql.Object {
//									return nil
//								},
//								Types: []*graphql.Object{
//									graphql.NewObject(graphql.ObjectConfig{
//										Name: "UnionFoo",
//										Fields: graphql.Fields{
//											"foo": &graphql.Field{
//												Name: "foo",
//												Type: graphql.String,
//											},
//										},
//									}),
//									graphql.NewObject(graphql.ObjectConfig{
//										Name: "UnionBar",
//										Fields: graphql.Fields{
//											"bar": &graphql.Field{
//												Name: "bar",
//												Type: graphql.String,
//											},
//										},
//									}),
//								},
//							}),
//						},
//					},
//				}),
//			},
//		},
//	}
//	for _, tt := range data {
//		tt.test(t)
//	}
//}
