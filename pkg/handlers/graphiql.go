package handlers

import (
	"net/http"

	"github.com/graphql-go/graphql"
)

func renderGraphiQL(rw http.ResponseWriter, params graphql.Params) {
	_, err := rw.Write(graphiql)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
	return
}

var graphiql = []byte(`<html>
	<head>
		<title>GraphiQL</title>
		<link href="https://unpkg.com/graphiql/graphiql.min.css" rel="stylesheet" />
	</head>
	<body style="margin: 0;">
		<div id="graphiql" style="height: 100vh;"></div>

		<script
	  crossorigin
	  src="https://unpkg.com/react/umd/react.production.min.js"
   ></script>
		<script
	  crossorigin
	  src="https://unpkg.com/react-dom/umd/react-dom.production.min.js"
   ></script>
		<script
	  crossorigin
	  src="https://unpkg.com/graphiql/graphiql.min.js"
   ></script>

		<script>
			const url = window.location.protocol + '//' + window.location.host + window.location.pathname;
			const params = new URLSearchParams(document.location.search.substring(1));
			const query = params.get('query');
			const variables = params.get('variables') ? JSON.parse(params.get('variables')) : undefined;
			const operationName = params.get('operationName');
			const graphQLFetcher = graphQLParams =>
				fetch(url, {
					method: 'POST',
					headers: { 'Content-Type': 'application/json' },
					body: JSON.stringify(graphQLParams),
				})
					.then(response => response.json())
					.catch(() => response.text());
			ReactDOM.render(
				React.createElement(GraphiQL, {
					fetcher: graphQLFetcher,
					query,
					operationName,
					variables,
				}),
				document.getElementById('graphiql'),
			);
		</script>
	</body>
</html>`)
