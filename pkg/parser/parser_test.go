package parser_test

import (
	"sort"
	"testing"

	"github.com/graphql-editor/stucco/pkg/parser"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func starwarsSchema() graphql.Schema {
	// type Query
	queryType := graphql.NewObject(graphql.ObjectConfig{
		Name:        "Query",
		Description: "The query type, represents all of the entry points into our object graph",
		Fields:      graphql.Fields{},
	})

	// type Mutation
	mutationType := graphql.NewObject(graphql.ObjectConfig{
		Name:        "Mutation",
		Description: "The mutation type, represents all updates we can make to our data",
		Fields:      graphql.Fields{},
	})

	// enum Episode
	episodeEnum := graphql.NewEnum(graphql.EnumConfig{
		Name:        "Episode",
		Description: "The episodes in the Star Wars trilogy",
		Values: graphql.EnumValueConfigMap{
			"NEWHOPE": &graphql.EnumValueConfig{
				Description: "Star Wars Episode IV: A New Hope, released in 1977.",
			},
			"EMPIRE": &graphql.EnumValueConfig{
				Description: "Star Wars Episode V: The Empire Strikes Back, released in 1980.",
			},
			"JEDI": &graphql.EnumValueConfig{
				Description: "Star Wars Episode VI: Return of the Jedi, released in 1983.",
			},
		},
	})

	// interface Character
	characterInterface := graphql.NewInterface(graphql.InterfaceConfig{
		Name:        "Character",
		Description: "A character from the Star Wars universe",
		Fields:      graphql.Fields{},
	})

	// enum LengthUnit
	lengthUnitEnum := graphql.NewEnum(graphql.EnumConfig{
		Name:        "LengthUnit",
		Description: "Units of height",
		Values: graphql.EnumValueConfigMap{
			"METER": &graphql.EnumValueConfig{
				Description: "The standard unit around the world",
			},
			"FOOT": &graphql.EnumValueConfig{
				Description: "Primarily used in the United States",
			},
		},
	})

	// type Human
	humanType := graphql.NewObject(graphql.ObjectConfig{
		Name:        "Human",
		Description: "A humanoid creature from the Star Wars universe",
		Fields:      graphql.Fields{},
		Interfaces:  []*graphql.Interface{characterInterface},
	})

	// type Droid
	droidType := graphql.NewObject(graphql.ObjectConfig{
		Name:        "Droid",
		Description: "An autonomous mechanical character in the Star Wars universe",
		Fields:      graphql.Fields{},
		Interfaces:  []*graphql.Interface{characterInterface},
	})

	// type FriendsConnection
	friendsConnectionType := graphql.NewObject(graphql.ObjectConfig{
		Name:        "FriendsConnection",
		Description: "A connection object for a character's friends",
		Fields:      graphql.Fields{},
	})

	// type FriendsEdge
	friendsEdgeType := graphql.NewObject(graphql.ObjectConfig{
		Name:        "FriendsEdge",
		Description: "An edge object for a character's friends",
		Fields:      graphql.Fields{},
	})

	// type PageInfo
	pageInfoType := graphql.NewObject(graphql.ObjectConfig{
		Name:        "PageInfo",
		Description: "Information for paginating this connection",
		Fields:      graphql.Fields{},
	})

	// type Review
	reviewType := graphql.NewObject(graphql.ObjectConfig{
		Name:        "Review",
		Description: "Represents a review for a movie",
		Fields:      graphql.Fields{},
	})

	// input ReviewInput
	reviewInput := graphql.NewInputObject(graphql.InputObjectConfig{
		Name:        "ReviewInput",
		Description: "The input object sent when someone is creating a new review",
		Fields:      graphql.InputObjectConfigFieldMap{},
	})

	// type Starship
	starshipType := graphql.NewObject(graphql.ObjectConfig{
		Name:   "Starship",
		Fields: graphql.Fields{},
	})

	// union SearchResult
	searchResultUnion := graphql.NewUnion(graphql.UnionConfig{
		Name: "SearchResult",
		ResolveType: func(params graphql.ResolveTypeParams) *graphql.Object {
			return nil
		},
		Types: []*graphql.Object{humanType, droidType, starshipType},
	})

	// scalar Time
	timeScalar := graphql.NewScalar(graphql.ScalarConfig{
		Name:         "Time",
		Serialize:    func(v interface{}) interface{} { return v },
		ParseValue:   func(v interface{}) interface{} { return v },
		ParseLiteral: func(v ast.Value) interface{} { return v.GetValue() },
	})

	// Query fields
	queryType.AddFieldConfig("hero", &graphql.Field{
		Name: "hero",
		Args: map[string]*graphql.ArgumentConfig{
			"episode": &graphql.ArgumentConfig{
				Type:         episodeEnum,
				DefaultValue: "NEWHOPE",
			},
		},
		Type: characterInterface,
	})
	queryType.AddFieldConfig("reviews", &graphql.Field{
		Name: "reviews",
		Args: graphql.FieldConfigArgument{
			"episode": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(episodeEnum),
			},
			"since": &graphql.ArgumentConfig{
				Type: timeScalar,
			},
		},
		Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(reviewType))),
	})
	queryType.AddFieldConfig("search", &graphql.Field{
		Name: "search",
		Args: graphql.FieldConfigArgument{
			"text": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.String),
			},
		},
		Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(searchResultUnion))),
	})
	queryType.AddFieldConfig("character", &graphql.Field{
		Name: "character",
		Args: graphql.FieldConfigArgument{
			"id": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.ID),
			},
		},
		Type: characterInterface,
	})
	queryType.AddFieldConfig("droid", &graphql.Field{
		Name: "droid",
		Args: graphql.FieldConfigArgument{
			"id": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.ID),
			},
		},
		Type: droidType,
	})
	queryType.AddFieldConfig("human", &graphql.Field{
		Name: "human",
		Args: graphql.FieldConfigArgument{
			"id": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.ID),
			},
		},
		Type: humanType,
	})
	queryType.AddFieldConfig("starship", &graphql.Field{
		Name: "starship",
		Args: graphql.FieldConfigArgument{
			"id": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.ID),
			},
		},
		Type: starshipType,
	})

	// Mutation fields
	mutationType.AddFieldConfig("createReview", &graphql.Field{
		Name: "createReview",
		Args: graphql.FieldConfigArgument{
			"episode": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(episodeEnum),
			},
			"review": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(reviewInput),
			},
		},
		Type: reviewType,
	})

	// Character fields
	characterInterface.AddFieldConfig("id", &graphql.Field{
		Name:        "id",
		Description: "The ID of the character",
		Type:        graphql.NewNonNull(graphql.ID),
	})
	characterInterface.AddFieldConfig("name", &graphql.Field{
		Name:        "name",
		Description: "The name of the character",
		Type:        graphql.NewNonNull(graphql.String),
	})
	characterInterface.AddFieldConfig("friends", &graphql.Field{
		Name:        "friends",
		Description: "The friends of the character, or an empty list if they have none",
		Type:        graphql.NewList(graphql.NewNonNull(characterInterface)),
	})
	characterInterface.AddFieldConfig("friendsConnection", &graphql.Field{
		Name: "friendsConnection",
		Args: graphql.FieldConfigArgument{
			"first": &graphql.ArgumentConfig{
				Type: graphql.Int,
			},
			"after": &graphql.ArgumentConfig{
				Type: graphql.ID,
			},
		},
		Description: "The friends of the character exposed as a connection with edges",
		Type:        graphql.NewNonNull(friendsConnectionType),
	})
	characterInterface.AddFieldConfig("appearsIn", &graphql.Field{
		Name:        "appearsIn",
		Description: "The movies this character appears in",
		Type:        graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(episodeEnum))),
	})

	// Human fields
	humanType.AddFieldConfig("id", &graphql.Field{
		Name:        "id",
		Description: "The ID of the human",
		Type:        graphql.NewNonNull(graphql.ID),
	})
	humanType.AddFieldConfig("name", &graphql.Field{
		Name:        "name",
		Description: "What this human calls themselves",
		Type:        graphql.NewNonNull(graphql.String),
	})
	humanType.AddFieldConfig("height", &graphql.Field{
		Name: "height",
		Args: graphql.FieldConfigArgument{
			"unit": &graphql.ArgumentConfig{
				DefaultValue: "METER",
				Type:         lengthUnitEnum,
			},
		},
		Description: "Height in the preferred unit, default is meters",
		Type:        graphql.NewNonNull(graphql.Float),
	})
	humanType.AddFieldConfig("mass", &graphql.Field{
		Name:        "mass",
		Description: "Mass in kilograms, or null if unknown",
		Type:        graphql.Float,
	})
	humanType.AddFieldConfig("friends", &graphql.Field{
		Name:        "friends",
		Description: "This human's friends, or an empty list if they have none",
		Type:        graphql.NewList(graphql.NewNonNull(characterInterface)),
	})
	humanType.AddFieldConfig("friendsConnection", &graphql.Field{
		Name: "friendsConnection",
		Args: graphql.FieldConfigArgument{
			"first": &graphql.ArgumentConfig{
				Type: graphql.Int,
			},
			"after": &graphql.ArgumentConfig{
				Type: graphql.ID,
			},
		},
		Description: "The friends of the human exposed as a connection with edges",
		Type:        graphql.NewNonNull(friendsConnectionType),
	})
	humanType.AddFieldConfig("appearsIn", &graphql.Field{
		Name:        "appearsIn",
		Description: "The movies this human appears in",
		Type:        graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(episodeEnum))),
	})
	humanType.AddFieldConfig("starships", &graphql.Field{
		Name:        "starships",
		Description: "A list of starships this person has piloted, or an empty list if none",
		Type:        graphql.NewList(graphql.NewNonNull(starshipType)),
	})

	// Droid fields
	droidType.AddFieldConfig("id", &graphql.Field{
		Name:        "id",
		Description: "The ID of the droid",
		Type:        graphql.NewNonNull(graphql.ID),
	})
	droidType.AddFieldConfig("name", &graphql.Field{
		Name:        "name",
		Description: "What others call this droid",
		Type:        graphql.NewNonNull(graphql.String),
	})
	droidType.AddFieldConfig("friends", &graphql.Field{
		Name:        "friends",
		Description: "This droid's friends, or an empty list if they have none",
		Type:        graphql.NewList(graphql.NewNonNull(characterInterface)),
	})
	droidType.AddFieldConfig("friendsConnection", &graphql.Field{
		Name: "friendsConnection",
		Args: graphql.FieldConfigArgument{
			"first": &graphql.ArgumentConfig{
				Type: graphql.Int,
			},
			"after": &graphql.ArgumentConfig{
				Type: graphql.ID,
			},
		},
		Description: "The friends of the droid exposed as a connection with edges",
		Type:        graphql.NewNonNull(friendsConnectionType),
	})
	droidType.AddFieldConfig("appearsIn", &graphql.Field{
		Name:        "appearsIn",
		Description: "The movies this droid appears in",
		Type:        graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(episodeEnum))),
	})
	droidType.AddFieldConfig("primaryFunction", &graphql.Field{
		Name:        "primaryFunction",
		Description: "This droid's primary function",
		Type:        graphql.String,
	})

	// FirendsConnection fields
	friendsConnectionType.AddFieldConfig("totalCount", &graphql.Field{
		Name:        "totalCount",
		Description: "The total number of friends",
		Type:        graphql.NewNonNull(graphql.Int),
	})
	friendsConnectionType.AddFieldConfig("edges", &graphql.Field{
		Name:        "edges",
		Description: "The edges for each of the character's friends.",
		Type:        graphql.NewList(graphql.NewNonNull(friendsEdgeType)),
	})
	friendsConnectionType.AddFieldConfig("friends", &graphql.Field{
		Name:        "friends",
		Description: "A list of the friends, as a convenience when edges are not needed.",
		Type:        graphql.NewList(graphql.NewNonNull(characterInterface)),
	})
	friendsConnectionType.AddFieldConfig("pageInfo", &graphql.Field{
		Name:        "pageInfo",
		Description: "Information for paginating this connection",
		Type:        graphql.NewNonNull(pageInfoType),
	})

	// FriendsEdge fields
	friendsEdgeType.AddFieldConfig("cursor", &graphql.Field{
		Name:        "cursor",
		Description: "A cursor used for pagination",
		Type:        graphql.NewNonNull(graphql.ID),
	})
	friendsEdgeType.AddFieldConfig("node", &graphql.Field{
		Name:        "node",
		Description: "The character represented by this friendship edge",
		Type:        characterInterface,
	})

	// PageInfo fields
	pageInfoType.AddFieldConfig("startCursor", &graphql.Field{
		Name: "startCursor",
		Type: graphql.NewNonNull(graphql.ID),
	})
	pageInfoType.AddFieldConfig("endCursor", &graphql.Field{
		Name: "endCursor",
		Type: graphql.NewNonNull(graphql.ID),
	})
	pageInfoType.AddFieldConfig("hasNextPage", &graphql.Field{
		Name: "hasNextPage",
		Type: graphql.NewNonNull(graphql.Boolean),
	})

	// Review fields
	reviewType.AddFieldConfig("stars", &graphql.Field{
		Name:        "stars",
		Description: "The number of stars this review gave, 1-5",
		Type:        graphql.NewNonNull(graphql.Int),
	})
	reviewType.AddFieldConfig("commentary", &graphql.Field{
		Name:        "commentary",
		Description: "Comment about the movie",
		Type:        graphql.String,
	})
	reviewType.AddFieldConfig("time", &graphql.Field{
		Name:        "time",
		Description: "when the review was posted",
		Type:        timeScalar,
	})

	// ReviewInput fields
	reviewInput.AddFieldConfig("stars", &graphql.InputObjectFieldConfig{
		Description: "0-5 stars",
		Type:        graphql.NewNonNull(graphql.Int),
	})
	reviewInput.AddFieldConfig("commentary", &graphql.InputObjectFieldConfig{
		Description: "Comment about the movie, optional",
		Type:        graphql.String,
	})
	reviewInput.AddFieldConfig("time", &graphql.InputObjectFieldConfig{
		Description: "when the review was posted",
		Type:        timeScalar,
	})

	// Starship fields
	starshipType.AddFieldConfig("id", &graphql.Field{
		Name:        "id",
		Description: "The ID of the starship",
		Type:        graphql.NewNonNull(graphql.ID),
	})
	starshipType.AddFieldConfig("name", &graphql.Field{
		Name:        "name",
		Description: "The name of the starship",
		Type:        graphql.NewNonNull(graphql.String),
	})
	starshipType.AddFieldConfig("length", &graphql.Field{
		Name: "length",
		Args: graphql.FieldConfigArgument{
			"unit": &graphql.ArgumentConfig{
				DefaultValue: "METER",
				Type:         lengthUnitEnum,
			},
		},
		Description: "Length of the starship, along the longest axis",
		Type:        graphql.NewNonNull(graphql.Float),
	})
	starshipType.AddFieldConfig("history", &graphql.Field{
		Name:        "history",
		Description: "coordinates tracking this ship",
		Type:        graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(graphql.Int))))),
	})

	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Types: []graphql.Type{
			queryType,
			mutationType,
			episodeEnum,
			characterInterface,
			lengthUnitEnum,
			humanType,
			droidType,
			friendsConnectionType,
			friendsEdgeType,
			pageInfoType,
			reviewType,
			reviewInput,
			starshipType,
			searchResultUnion,
			timeScalar,
		},
		Query:    queryType,
		Mutation: mutationType,
	})
	if err != nil {
		panic(err)
	}
	return schema
}

func gqlTypeRef(t *testing.T, expected graphql.Type, actual graphql.Type) {
	assert.IsType(t, expected, actual)
	switch expectedTyped := expected.(type) {
	case *graphql.List:
		actualTyped := actual.(*graphql.List)
		gqlTypeRef(t, expectedTyped.OfType, actualTyped.OfType)
	case *graphql.NonNull:
		actualTyped := actual.(*graphql.NonNull)
		gqlTypeRef(t, expectedTyped.OfType, actualTyped.OfType)
	default:
		assert.Equal(t, expected.Name(), actual.Name())
	}
}

type typeSorter []graphql.Type

func (t typeSorter) Len() int {
	return len(t)
}

func (t typeSorter) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func (t typeSorter) Less(i, j int) bool {
	return t[i].Name() < t[j].Name()
}

type interfaces []*graphql.Interface

func (i interfaces) toTypeList() []graphql.Type {
	typeList := make([]graphql.Type, 0, len(i))
	for _, iface := range i {
		typeList = append(typeList, iface)
	}
	return typeList
}

type objects []*graphql.Object

func (o objects) toTypeList() []graphql.Type {
	typeList := make([]graphql.Type, 0, len(o))
	for _, obj := range o {
		typeList = append(typeList, obj)
	}
	return typeList
}

func gqlTypeRefs(t *testing.T, expected []graphql.Type, actual []graphql.Type) {
	if assert.Len(t, actual, len(expected)) {
		sort.Sort(typeSorter(expected))
		sort.Sort(typeSorter(actual))
		for i, tp := range expected {
			gqlTypeRef(t, tp, actual[i])
		}
	}
}

func gqlInputFieldEqual(t *testing.T, expected *graphql.InputObjectField, actual *graphql.InputObjectField) {
	assert.Equal(t, expected.Name(), actual.Name())
	assert.Equal(t, expected.Description(), actual.Description())
	assert.Equal(t, expected.String(), actual.String())
	assert.Equal(t, expected.Error(), actual.Error())
	gqlTypeRef(t, expected.Type, actual.Type)
}

func gqlInputObjectFieldsEqual(t *testing.T, expected graphql.InputObjectFieldMap, actual graphql.InputObjectFieldMap) {
	if assert.Len(t, actual, len(expected)) {
		for k, field := range expected {
			gqlInputFieldEqual(t, field, actual[k])
		}
	}
}

type argsSorter []*graphql.Argument

func (a argsSorter) Len() int {
	return len(a)
}

func (a argsSorter) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a argsSorter) Less(i, j int) bool {
	return a[i].Name() < a[j].Name()
}

func gqlFieldEqual(t *testing.T, expected *graphql.FieldDefinition, actual *graphql.FieldDefinition) {
	assert.Equal(t, expected.Name, actual.Name)
	assert.Equal(t, expected.DeprecationReason, actual.DeprecationReason)
	assert.Equal(t, expected.Description, actual.Description)
	gqlTypeRef(t, expected.Type, actual.Type)
	if assert.Len(t, expected.Args, len(expected.Args)) {
		// sort args as order does not matter
		sort.Sort(argsSorter(expected.Args))
		sort.Sort(argsSorter(actual.Args))
		for i, v := range expected.Args {
			assert.Equal(t, v.Name(), actual.Args[i].Name())
			assert.Equal(t, v.Description(), actual.Args[i].Description())
			assert.Equal(t, v.DefaultValue, actual.Args[i].DefaultValue)
			gqlTypeRef(t, v.Type, actual.Args[i].Type)
		}
	}
}

func gqlFieldsEqual(t *testing.T, expected graphql.FieldDefinitionMap, actual graphql.FieldDefinitionMap) {
	if assert.Len(t, actual, len(expected)) {
		for k, field := range expected {
			gqlFieldEqual(t, field, actual[k])
		}
	}
}

func gqlTypeEqual(t *testing.T, expected graphql.Type, actual graphql.Type) {
	assert.Equal(t, expected.Name(), actual.Name())
	assert.Equal(t, expected.Description(), actual.Description())
	assert.Equal(t, expected.Error(), actual.Error())
	assert.Equal(t, expected.String(), actual.String())
	if assert.IsType(t, expected, actual) {
		switch expectedTyped := expected.(type) {
		case *graphql.Object:
			actualTyped := actual.(*graphql.Object)
			gqlFieldsEqual(t, expectedTyped.Fields(), actualTyped.Fields())
			gqlTypeRefs(t, interfaces(expectedTyped.Interfaces()).toTypeList(), interfaces(actualTyped.Interfaces()).toTypeList())
		case *graphql.Interface:
			actualTyped := actual.(*graphql.Interface)
			gqlFieldsEqual(t, expectedTyped.Fields(), actualTyped.Fields())
		case *graphql.Union:
			actualTyped := actual.(*graphql.Union)
			gqlTypeRefs(t, objects(expectedTyped.Types()).toTypeList(), objects(actualTyped.Types()).toTypeList())
		case *graphql.InputObject:
			actualTyped := actual.(*graphql.InputObject)
			gqlInputObjectFieldsEqual(t, expectedTyped.Fields(), actualTyped.Fields())
		}
	}
}

func schemaEqual(t *testing.T, expected graphql.Schema, actual graphql.Schema) {
	for _, obj := range []struct {
		expected *graphql.Object
		actual   *graphql.Object
	}{
		{expected.QueryType(), actual.QueryType()},
		{expected.MutationType(), actual.MutationType()},
		{expected.SubscriptionType(), actual.SubscriptionType()},
	} {
		assert.True(t, obj.expected != nil || obj.expected == obj.actual)
		if obj.expected != nil {
			gqlTypeEqual(t, obj.expected, obj.actual)
		}
	}
	if assert.Len(t, expected.TypeMap(), len(actual.TypeMap())) {
		for k := range expected.TypeMap() {
			gqlTypeEqual(t, expected.Type(k), actual.Type(k))
		}
	}
}

const starwarsSchemaString = `# The query type, represents all of the entry points into our object graph
type Query {
    hero(episode: Episode = NEWHOPE): Character
    reviews(episode: Episode!, since: Time): [Review!]!
    search(text: String!): [SearchResult!]!
    character(id: ID!): Character
    droid(id: ID!): Droid
    human(id: ID!): Human
    starship(id: ID!): Starship
}
"""The mutation type, represents all updates we can make to our data"""
type Mutation {
    createReview(episode: Episode!, review: ReviewInput!): Review
}
"""The episodes in the Star Wars trilogy"""
enum Episode {
    """Star Wars Episode IV: A New Hope, released in 1977."""
    NEWHOPE
    """Star Wars Episode V: The Empire Strikes Back, released in 1980."""
    EMPIRE
    """Star Wars Episode VI: Return of the Jedi, released in 1983."""
    JEDI
}
"""A character from the Star Wars universe"""
interface Character {
    """The ID of the character"""
    id: ID!
    """The name of the character"""
    name: String!
    """The friends of the character, or an empty list if they have none"""
    friends: [Character!]
    """The friends of the character exposed as a connection with edges"""
    friendsConnection(first: Int, after: ID): FriendsConnection!
    """The movies this character appears in"""
    appearsIn: [Episode!]!
}
"""Units of height"""
enum LengthUnit {
    """The standard unit around the world"""
    METER
    """Primarily used in the United States"""
    FOOT
}
"""A humanoid creature from the Star Wars universe"""
type Human implements Character {
    """The ID of the human"""
    id: ID!
    """What this human calls themselves"""
    name: String!
    """Height in the preferred unit, default is meters"""
    height(unit: LengthUnit = METER): Float!
    """Mass in kilograms, or null if unknown"""
    mass: Float
    """This human's friends, or an empty list if they have none"""
    friends: [Character!]
    """The friends of the human exposed as a connection with edges"""
    friendsConnection(first: Int, after: ID): FriendsConnection!
    """The movies this human appears in"""
    appearsIn: [Episode!]!
    """A list of starships this person has piloted, or an empty list if none"""
    starships: [Starship!]
}
"""An autonomous mechanical character in the Star Wars universe"""
type Droid implements Character {
    """The ID of the droid"""
    id: ID!
    """What others call this droid"""
    name: String!
    """This droid's friends, or an empty list if they have none"""
    friends: [Character!]
    """The friends of the droid exposed as a connection with edges"""
    friendsConnection(first: Int, after: ID): FriendsConnection!
    """The movies this droid appears in"""
    appearsIn: [Episode!]!
    """This droid's primary function"""
    primaryFunction: String
}
"""A connection object for a character's friends"""
type FriendsConnection {
    """The total number of friends"""
    totalCount: Int!
    """The edges for each of the character's friends."""
    edges: [FriendsEdge!]
    """A list of the friends, as a convenience when edges are not needed."""
    friends: [Character!]
    """Information for paginating this connection"""
    pageInfo: PageInfo!
}
"""An edge object for a character's friends"""
type FriendsEdge {
    """A cursor used for pagination"""
    cursor: ID!
    """The character represented by this friendship edge"""
    node: Character
}
"""Information for paginating this connection"""
type PageInfo {
    startCursor: ID!
    endCursor: ID!
    hasNextPage: Boolean!
}
"""Represents a review for a movie"""
type Review {
    """The number of stars this review gave, 1-5"""
    stars: Int!
    """Comment about the movie"""
    commentary: String
    """when the review was posted"""
    time: Time
}
"""The input object sent when someone is creating a new review"""
input ReviewInput {
    """0-5 stars"""
    stars: Int!
    """Comment about the movie, optional"""
    commentary: String
    """when the review was posted"""
    time: Time
}
type Starship {
    """The ID of the starship"""
    id: ID!
    """The name of the starship"""
    name: String!
    """Length of the starship, along the longest axis"""
    length(unit: LengthUnit = METER): Float!
    """coordinates tracking this ship"""
    history: [[Int!]!]!
}
union SearchResult = Human | Droid | Starship
scalar Time

schema {
	query: Query
	mutation: Mutation
}
`

func TestParserParse(t *testing.T) {
	data := []struct {
		title       string
		config      parser.Config
		input       string
		expected    graphql.Schema
		expectedErr assert.ErrorAssertionFunc
	}{
		{
			title:       "Star wars schema",
			config:      parser.Config{},
			input:       starwarsSchemaString,
			expected:    starwarsSchema(),
			expectedErr: assert.NoError,
		},
	}
	for _, tt := range data {
		t.Run(tt.title, func(t *testing.T) {
			parser := parser.NewParser(tt.config)
			schema, err := parser.Parse(tt.input)
			tt.expectedErr(t, err)
			schemaEqual(t, tt.expected, schema)
		})
	}
}

type scalarParseMock struct {
	mock.Mock
}

func (m *scalarParseMock) Parse(v interface{}) interface{} {
	return m.Called(v).Get(0)
}

type scalarSerializeMock struct {
	mock.Mock
}

func (m *scalarSerializeMock) Serialize(v interface{}) interface{} {
	return m.Called(v).Get(0)
}

func TestScalarParseSerialize(t *testing.T) {
	p := parser.NewParser(parser.Config{})
	schema, err := p.Parse(starwarsSchemaString)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"value",
		schema.Type("Time").(*graphql.Scalar).ParseValue("value"),
	)
	assert.Equal(
		t,
		"value",
		schema.Type("Time").(*graphql.Scalar).ParseLiteral(&ast.StringValue{Value: "value"}),
	)
	assert.Equal(
		t,
		"value",
		schema.Type("Time").(*graphql.Scalar).Serialize("value"),
	)
	scalarParseMock := new(scalarParseMock)
	scalarSerializeMock := new(scalarSerializeMock)
	scalarParseMock.On("Parse", "value").Return("value")
	scalarSerializeMock.On("Serialize", "value").Return("value")
	p = parser.NewParser(parser.Config{
		Scalars: map[string]parser.ScalarFunctions{
			"Time": parser.ScalarFunctions{
				Parse:     scalarParseMock.Parse,
				Serialize: scalarSerializeMock.Serialize,
			},
		},
	})
	schema, err = p.Parse(starwarsSchemaString)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"value",
		schema.Type("Time").(*graphql.Scalar).ParseValue("value"),
	)
	assert.Equal(
		t,
		"value",
		schema.Type("Time").(*graphql.Scalar).ParseLiteral(&ast.StringValue{Value: "value"}),
	)
	assert.Equal(
		t,
		"value",
		schema.Type("Time").(*graphql.Scalar).Serialize("value"),
	)
	scalarParseMock.AssertNumberOfCalls(t, "Parse", 2)
	scalarSerializeMock.AssertNumberOfCalls(t, "Serialize", 1)
}

func TestDefaultInterfaceResolveType(t *testing.T) {
	p := parser.NewParser(parser.Config{})
	schema, err := p.Parse(starwarsSchemaString)
	assert.NoError(t, err)
	data := []struct {
		title    string
		input    interface{}
		expected *graphql.Object
	}{
		{title: "FailNil"},
		{title: "FailArbitrary", input: map[string]interface{}{
			"id": "XYZ",
		}},
		{
			title: "MatchHuman",
			input: map[string]interface{}{
				"__typename": "Human",
			},
			expected: schema.Type("Human").(*graphql.Object),
		},
		{
			title: "MatchDroid",
			input: map[string]interface{}{
				"__typename": "Droid",
			},
			expected: schema.Type("Droid").(*graphql.Object),
		},
	}
	for _, tt := range data {
		t.Run(tt.title, func(t *testing.T) {
			assert.Equal(
				t,
				tt.expected,
				schema.Type("Character").(*graphql.Interface).ResolveType(graphql.ResolveTypeParams{
					Value: tt.input,
				}),
			)
		})
	}
}

type interfaceResolveTypeMock struct {
	mock.Mock
}

func (m *interfaceResolveTypeMock) ResolveType(p graphql.ResolveTypeParams) *graphql.Object {
	return m.Called(p).Get(0).(*graphql.Object)
}

func TestInterfaceCallsResolveType(t *testing.T) {
	interfaceResolveTypeMock := new(interfaceResolveTypeMock)
	p := parser.NewParser(parser.Config{
		Interfaces: map[string]graphql.ResolveTypeFn{
			"Character": interfaceResolveTypeMock.ResolveType,
		},
	})
	schema, err := p.Parse(starwarsSchemaString)
	assert.NoError(t, err)
	data := []struct {
		title    string
		input    graphql.ResolveTypeParams
		expected *graphql.Object
	}{
		{
			title:    "UserResolveHuman",
			input:    graphql.ResolveTypeParams{Value: "Human"},
			expected: schema.Type("Human").(*graphql.Object),
		},
		{
			title:    "UserResolveDroid",
			input:    graphql.ResolveTypeParams{Value: "Droid"},
			expected: schema.Type("Droid").(*graphql.Object),
		},
	}
	for _, tt := range data {
		t.Run(tt.title, func(t *testing.T) {
			interfaceResolveTypeMock.On("ResolveType", tt.input).Return(tt.expected)
			assert.Equal(
				t,
				tt.expected,
				schema.Type("Character").(*graphql.Interface).ResolveType(tt.input),
			)
			interfaceResolveTypeMock.AssertCalled(t, "ResolveType", tt.input)
		})
	}
}

func TestDefaultUnionResolveType(t *testing.T) {
	p := parser.NewParser(parser.Config{})
	schema, err := p.Parse(starwarsSchemaString)
	assert.NoError(t, err)
	data := []struct {
		title    string
		input    interface{}
		expected *graphql.Object
	}{
		{title: "FailNil"},
		{title: "FailArbitrary", input: map[string]interface{}{
			"id": "XYZ",
		}},
		{
			title: "MatchHuman",
			input: map[string]interface{}{
				"__typename": "Human",
			},
			expected: schema.Type("Human").(*graphql.Object),
		},
		{
			title: "MatchDroid",
			input: map[string]interface{}{
				"__typename": "Droid",
			},
			expected: schema.Type("Droid").(*graphql.Object),
		},
		{
			title: "MatchStarship",
			input: map[string]interface{}{
				"__typename": "Starship",
			},
			expected: schema.Type("Starship").(*graphql.Object),
		},
	}
	for _, tt := range data {
		t.Run(tt.title, func(t *testing.T) {
			assert.Equal(
				t,
				tt.expected,
				schema.Type("SearchResult").(*graphql.Union).ResolveType(graphql.ResolveTypeParams{
					Value: tt.input,
				}),
			)
		})
	}
}

type unionResolveTypeMock struct {
	mock.Mock
}

func (m *unionResolveTypeMock) ResolveType(p graphql.ResolveTypeParams) *graphql.Object {
	return m.Called(p).Get(0).(*graphql.Object)
}

func TestUnionCallsResolveType(t *testing.T) {
	unionResolveTypeMock := new(unionResolveTypeMock)
	p := parser.NewParser(parser.Config{
		Unions: map[string]graphql.ResolveTypeFn{
			"SearchResult": unionResolveTypeMock.ResolveType,
		},
	})
	schema, err := p.Parse(starwarsSchemaString)
	assert.NoError(t, err)
	data := []struct {
		title    string
		input    graphql.ResolveTypeParams
		expected *graphql.Object
	}{
		{
			title:    "UserResolveHuman",
			input:    graphql.ResolveTypeParams{Value: "Human"},
			expected: schema.Type("Human").(*graphql.Object),
		},
		{
			title:    "UserResolveDroid",
			input:    graphql.ResolveTypeParams{Value: "Droid"},
			expected: schema.Type("Droid").(*graphql.Object),
		},
		{
			title:    "UserResolveStarship",
			input:    graphql.ResolveTypeParams{Value: "Starship"},
			expected: schema.Type("Starship").(*graphql.Object),
		},
	}
	for _, tt := range data {
		t.Run(tt.title, func(t *testing.T) {
			unionResolveTypeMock.On("ResolveType", tt.input).Return(tt.expected)
			assert.Equal(
				t,
				tt.expected,
				schema.Type("SearchResult").(*graphql.Union).ResolveType(tt.input),
			)
			unionResolveTypeMock.AssertCalled(t, "ResolveType", tt.input)
		})
	}
}

type fieldResolveMock struct {
	mock.Mock
}

func (m *fieldResolveMock) Resolve(p graphql.ResolveParams) (interface{}, error) {
	called := m.Called(p)
	return called.Get(0), called.Error(1)
}

func TestSetsUserFieldResolve(t *testing.T) {
	fieldResolveMock := new(fieldResolveMock)
	p := parser.NewParser(parser.Config{
		Resolvers: map[string]graphql.FieldResolveFn{
			"Query.hero": fieldResolveMock.Resolve,
		},
	})
	schema, err := p.Parse(starwarsSchemaString)
	assert.NoError(t, err)
	fieldResolveMock.On("Resolve", graphql.ResolveParams{
		Source: "source",
	}).Return("data", nil)
	i, err := schema.Type("Query").(*graphql.Object).Fields()["hero"].Resolve(graphql.ResolveParams{
		Source: "source",
	})
	assert.NoError(t, err)
	assert.Equal(
		t,
		"data",
		i,
	)
	fieldResolveMock.AssertCalled(t, "Resolve", graphql.ResolveParams{
		Source: "source",
	})
}
