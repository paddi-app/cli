build:
    go build -ldflags "-s -w" -o bin/paddi main.go

build-staging:
    go build -ldflags "-s -w -X github.com/paddi-app/cli/internal/config.defaultAPIBase=https://dev-api.paddi.app" -o bin/paddi-staging main.go
