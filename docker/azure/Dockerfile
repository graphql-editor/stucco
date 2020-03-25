FROM gqleditor/azure-functions-golang-worker:v0.1.0 AS build

ARG VERSION
ENV HOME=/home
COPY . /home/site/wwwroot
WORKDIR /home/site/wwwroot

RUN go run hack/tasks.go build_azure --version=${VERSION}

FROM debian:buster-slim

RUN apt-get update && apt-get install -y \
    libicu63 libssl1.1 ca-certificates \
 && rm -rf /var/lib/apt/lists/*

ENV AzureWebJobsScriptRoot=/home/site/wwwroot \
    HOME=/home \
    FUNCTIONS_WORKER_RUNTIME=golang \
    DOTNET_USE_POLLING_FILE_WATCHER=true \
	ASPNETCORE_URLS=http://*:80 \
	AZURE_GOLANG_WORKER_PREBUILT_graphql=/home/site/wwwroot/graphql/function.so

COPY --from=build [ "/azure-functions-host", "/azure-functions-host" ]
COPY --from=build [ "/FuncExtensionBundles", "/FuncExtensionBundles" ]
COPY --from=build [ "/home/site/wwwroot/bin/azure/worker", "/azure-functions-host/workers/golang/worker" ]
COPY --from=build [ "/home/site/wwwroot/bin/azure/function.so", "/home/site/wwwroot/graphql/function.so" ]
COPY pkg/providers/azure/function /home/site/wwwroot
COPY pkg/providers/azure/function/worker.config.json /azure-functions-host/workers/golang/
WORKDIR /home/site/wwwroot
CMD [ "/azure-functions-host/Microsoft.Azure.WebJobs.Script.WebHost" ]
