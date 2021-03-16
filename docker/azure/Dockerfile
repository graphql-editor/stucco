FROM golang:1.16 as build

ENV CGO_ENABLED=0
ADD . /app
WORKDIR /app
RUN go run hack/tasks.go bin/cli/linux/amd64/stucco --version=${VERSION}

FROM mcr.microsoft.com/azure-functions/dotnet:3.0

ENV AzureFunctionsJobHost__Logging__Console__IsEnabled=true

COPY --from=build /app/pkg/providers/azure/function /home/site/wwwroot
COPY --from=build /app/bin/cli/linux/amd64/stucco /home/site/wwwroot/stucco
