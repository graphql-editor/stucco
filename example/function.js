// Fake database
const database = {
	humans: [
		{
			id: "luke_skywalker",
			name: "Luke Skywalker",
			height: 1.72,
			mass: 73,
			friends: ["han_solo", "r2_d2", "c_3po"],
			appearsIn: ["NEWHOPE", "EMPIRE", "JEDI"],
			starships: ["x_wing"],
		},
		{
			id: "han_solo",
			name: "Han Solo",
			height: 1.8,
			mass: 80,
			friends: ["luke_skywalker"],
			appearsIn: ["NEWHOPE", "EMPIRE", "JEDI"],
			starships: ["millenium_falcon"]
		},
		{
			id: "darth_vader",
			name: "Darth Vader",
			height: 2.03,
			mass: 120,
			friends: [],
			appearsIn: ["NEWHOPE", "EMPIRE", "JEDI"],
			starships: ["tie_advanced_x1"],
		}
	].map(v => ({__typename: "Human", ...v})), // field __typename allows for automatic type deduction without implementing ResolveType functions on interfaces and unions. It is optional, as long as user adds ResolveType functions to all of interfaces and unions that include Human type. In this example, SearchResults union does not have ResolveType implemented to ilustrate it.
	droids: [
		{
			id: "r2_d2",
			name: "R2-D2",
			friends: ["luke_skywalker", "c_3po"],
			appearsIn: ["NEWHOPE", "EMPIRE", "JEDI"],
			primaryFunction: "engineering",
		},
		{
			id: "c_3po",
			name: "C-3PO",
			friends: ["luke_skywalker", "r2_d2"],
			appearsIn: ["NEWHOPE", "EMPIRE", "JEDI"],
			primaryFunction: "protocol",
		},
	].map(v => ({__typename: "Droid", ...v})),
	starships: [
		{
			id: "x_wing",
			name: "X-Wing starfigher",
			length: 12.5,
			history: [],
		},
		{
			id: "millenium_falcon",
			name: "Millenium Falcon",
			length: 34.52,
			history: [],
		},
		{
			id: "tie_advanced_x1",
			name: "TIE Advanced x2",
			length: 9.2,
			history: [],
		},
	].map(v => ({__typename: "Starship", ...v})),
	reviews: [],
}

let episodeToHero = {
	JEDI: 2,
	EMPIRE: 1,
}

const starship = id => database
	.starships
	.find(v => v.id === id)

const character = id => database
	.humans
	.concat(database.droids)
	.find(v => v.id === id)

const friends = input => input.map(v => character(v))

function listenRandomGreet(emitter) {
	const handle = setInterval(() => {
		emitter.emit();
	}, 1000);
	emitter.on('close', (err) => {
		if (err) {
			console.error(err);
		}
		clearInterval(handle);
	});
}

function listenTickAt(emitter) {
	const handle = setInterval(() => {
		emitter.emit((new Date()).toString());
	}, 1000);
	emitter.on('close', (err) => {
		if (err) {
			console.error(err);
		}
		clearInterval(handle);
	});
}

// Here we are implementing functions as defined by stucco.json
module.exports = {
	// Query.hero
	hero: input => database.humans[episodeToHero[input.arguments.episode] || 0],

	// Query.reviews
	reviews: input => database.reviews.filter(
		v => v.episode === input.arguments.episode &&
			(!input.arguments.since || input.arguments.since < v.time
	), []),

	// Query.search
	search: input => database
		.humans
		.concat(database.droids)
		.concat(database.starships)
		.filter(v => v.name.match(input.arguments.text)),

	// Query.character
	character: input => character(input.arguments.id),

	// Character interface
	Character: input => database.humans.find(v => input.value.id === v.id) ? "Human" : "Droid",

	// Query.droid
	droid: input => database.droids.find(v => v.id === input.arguments.id),

	// Query.human
	human: input => database.humans.find(v => v.id === input.arguments.id),

	// Query.starship
	starship: input => starship(input.arguments.id),

	// Mutation.createReview
	createReview: input => {
		database.reviews.push({
			episode: input.arguments.episode,
			...input.arguments.review,
		})
		return input.arguments.review
	},

	// Human/Droid.friends
	friends: input => friends(input.source.friends),

	// Human/Droid.friendsConnection
	friendsConnection: input => {
		const after = input.arguments && input.arguments.after ? 
			input.source.friends.findIndex(f => f === input.arguments.after) + 1 : 0
		const first = input.arguments && input.arguments.first ?
			after+input.arguments.first : input.source.friends.length
		const friendsSlice = friends(input.source.friends.slice(after, first))
		const pageInfo = friendsSlice.length > 0 ? {
			startCursor: friendsSlice[0].id,
			endCursor: friendsSlice.slice(-1)[0].id,
			hasNextPage: input.source.friends.slice(-1)[0].id !== friendsSlice.slice(-1)[0].id,
		} : {startCursor: "", endCursor: "", hasNextPage: false}
		return {
			totalCount: friendsSlice.length,
			edges: friendsSlice.map(f => ({
				cursor: f.id,
				node: f,
			})),
			friends: friendsSlice,
			pageInfo,
		}
	},

	// Human.starships
	starships: input => input.source.starships.map(v => starship(v)),

	// Subscription example
	randomGreet: () => `Hey, ${
		database.humans.concat(database.droids)[Math.floor(
			(database.humans.length + database.droids.length) * Math.random()
		)].name
	}`,

	// Subscription example with payload
	tickAt: (input) => `Listener had a tick at ${input.info.rootValue.payload}`,

	listen: (input, emitter) => {
	  const findSelection = (sel) => input.operation.selectionSet.find((v) => v.name === sel);
	  if (findSelection('randomGreet')) {
		listenRandomGreet(emitter);
	  } else if (findSelection('tickAt')) {
		listenTickAt(emitter);
	  }
	},
}
