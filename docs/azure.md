# About

Stucco router that can be run by Azure Functions host.

# Usage

## Run locally

Currently router can only be ran locally on Linux and macOS because it depends on [azure-functions-golang-worker](https://github.com/graphql-editor/azure-functions-golang-worker/) which does not support running plugin functions on Windows.

### Dependencies
* [azure-functions-core-tools@v3](https://github.com/Azure/azure-functions-core-tools)
* [azure-functions-golang-worker](https://github.com/graphql-editor/azure-functions-golang-worker/)

### Run

```
$ STUCCO_SCHEMA=path/to/schema.graphql STUCCO_CONFIG=path/to/stucco.json STUCCO_WORKER_BASE_URL=http://worker.url func start
```

## Docker

### New image

To create new router image just add schema.graphql and stucco.json to base image

```
FROM gqleditor/stucco-router-azure-worker:latest

COPY schema.graphql /home/site/wwwroot/schema.graphql
COPY stucco.json /home/site/wwwroot/stucco.json
```

### Run using base image

```
$ docker run -p 8080:80 -e STUCCO_SCHEMA=path/to/schema.graphql -e STUCCO_CONFIG=path/to/stucco.json -e STUCCO_WORKER_BASE_URL=http://worker.url gqleditor/stucco-router-azure-worker:latest
```

# Notes

By default function.json has `function` auth level which makes them inaccessible locally. Edit `authLevel` field in graphql/function.json to make it debuggable locally.
