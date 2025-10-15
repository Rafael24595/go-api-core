# Go API Core

This repository provides the core implementation for a Go-based API client library. It enables making requests to APIs using parameterized requests, returning structured responses. The library also supports data access and persistence, including requests, responses, collections with their contexts, and user data.  
By default, it uses a custom CSVT format to manage data storage.

## Project Structure

- `.env`, `.env.template`: Environment configuration files.
- `go.mod`, `go.sum`: Go module dependencies.
- `src/`: Main source code
  - `commons/`: Utilities and configuration
  - `domain/`: Domain models (e.g., cookie, openapi)
  - `infrastructure/`: Repositories and data access
- `test/`: Unit and integration tests
- `.github/workflows/`: CI/CD workflows for build and release
- `db/`: Example data files

## Build & Test

To build and test the project, run:

```sh
go build -v ./...
go test -v ./...
```

## Environment Setup

Copy `.env.template` to `.env` and adjust variables as needed.

## Release Workflow

GitHub Actions are configured for automated build and release. See [.github/workflows/publish-release.yml](.github/workflows/publish-release.yml).

## Key Features

- Parameterized API requests and structured responses
- Data access and persistence for requests, responses, collections, contexts, and user data
- Default storage management using custom CSVT format
- OpenAPI import and schema handling ([src/domain/openapi/FactoryCollection.go](src/domain/openapi/FactoryCollection.go))
- Collection and request management ([src/infrastructure/repository/ManagerCollection.go](src/infrastructure/repository/ManagerCollection.go))
- Utilities for version parsing ([src/commons/utils/Version.go](src/commons/utils/Version.go))
- Extensive unit tests ([test/domain/openapi/openapi_test.go](test/domain/openapi/openapi_test.go), [test/domain/context/context_test.go](test/domain/context/context_test.go))
