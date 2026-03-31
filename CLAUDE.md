# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is the **Kestra Terraform Provider** — a Terraform provider for managing Kestra workflow orchestrator resources via Infrastructure as Code. The provider supports Kestra 1.0.x and above.

Key architectural note: The provider runs both an old (terraform-plugin-sdk v2) and new (terraform-plugin-framework) implementation simultaneously using a mux server pattern in `main.go`. This allows gradual migration to the modern framework while maintaining backward compatibility.

## Common Development Commands

### Building and Testing

```bash
# Build the provider
go install

# Generate/update documentation (required before release)
go generate

# Run acceptance tests (requires running Kestra instance)
# First, start the test environment:
./init-tests-env.sh
# Then run tests with Kestra credentials:
TF_ACC=1 KESTRA_URL=http://127.0.0.1:8088 KESTRA_USERNAME=root@root.com KESTRA_PASSWORD='Root!1234' go test -v -cover ./internal/provider/

# Run a single test
TF_ACC=1 KESTRA_URL=http://127.0.0.1:8088 KESTRA_USERNAME=root@root.com KESTRA_PASSWORD='Root!1234' go test -v -run TestAccResourceFlow ./internal/provider/

# Generate test coverage report
go test -v -coverprofile=test-coverage-result.out ./internal/provider/
# View in browser:
go tool cover -html=test-coverage-result.out
# View in terminal:
go tool cover -func=test-coverage-result.out
```

### Dependency Management

```bash
# Add a new dependency
go get github.com/author/dependency

# Tidy dependencies
go mod tidy

# Commit changes to go.mod and go.sum when done
```

### Release

- Ensure documentation is up to date: `go generate`
- Create and push a tag on the main branch (goreleaser handles the rest)

## Architecture

### Directory Structure

- **`internal/provider/`** — SDK v2 implementation (deprecated, being phased out)
  - `provider.go` — Provider configuration and resource/data source registration
  - `client.go` — HTTP client for Kestra API communication
  - `resource_*.go` — Resource implementations
  - `data_source_*.go` — Data source implementations
  - `utils_*.go` — Helper functions for each resource type

- **`internal/provider_v2/`** — Modern framework implementation (new resources go here)
  - `provider.go` — Provider configuration using terraform-plugin-framework
  - `resource_*.go` — Framework-based resource implementations
  - `data_source_*.go` — Framework-based data source implementations
  - `sdk_client/` — Wrapper around Kestra Go SDK client

- **`examples/`** — Example Terraform configurations for resources and data sources (used for documentation)

### Provider Mux Pattern

`main.go` uses `tf5muxserver.NewMuxServer()` to combine:
1. The new framework-based provider (preferred for new work)
2. The old SDK v2 provider (for backward compatibility)

When adding new resources, prefer `provider_v2/` implementation. The mux server automatically routes requests to the appropriate implementation.

### Client and Authentication

The `Client` struct in `internal/provider/client.go` handles all HTTP communication with Kestra. It supports:
- Basic auth (username/password)
- JWT tokens (Enterprise)
- API tokens (Enterprise)
- Custom headers
- Configurable timeouts

Provider authentication is configured via environment variables or Terraform config:
- `KESTRA_URL` — Kestra API endpoint
- `KESTRA_USERNAME` / `KESTRA_PASSWORD` — Basic auth
- `KESTRA_JWT` or `KESTRA_API_TOKEN` — Enterprise authentication
- `KESTRA_TENANT_ID` — Multi-tenant deployments (defaults to "main")
- `KESTRA_TIMEOUT` — HTTP request timeout in seconds

## Testing

### Acceptance Tests

Acceptance tests require a running Kestra instance. The test environment setup is in `docker-compose-ci.yml` and `init-tests-env.sh`. Tests use the `TF_ACC` environment variable to distinguish acceptance tests from unit tests.

Most test files follow the pattern `*_test.go` and test both CRUD operations and data sources. Key test patterns:
- `TestAcc*Create` — Validates resource creation
- `TestAcc*Update` — Validates resource updates
- `TestAcc*Delete` — Validates resource deletion
- `TestAccDataSource*` — Validates data source queries

### Unit Tests

Provider-level configuration tests exist in `provider_test.go`.

## Documentation

Documentation is auto-generated from:
- Provider schema descriptions
- Resource/data source attribute descriptions
- Example configurations in `examples/`

Running `go generate` (via `terraform-plugin-docs`) creates markdown files in `docs/`. These are synced to the main Kestra docs repo via GitHub Actions (see `.github/workflows/docs.yml`).

## Key Dependencies

- `github.com/hashicorp/terraform-plugin-framework` — Modern Terraform plugin framework
- `github.com/hashicorp/terraform-plugin-sdk/v2` — Legacy SDK (being phased out)
- `github.com/kestra-io/client-sdk/go-sdk` — Kestra API client
- `gopkg.in/yaml.v2` — YAML parsing (Kestra flows are YAML)
- `github.com/json-iterator/go` — JSON parsing performance library

## Release Process

1. Ensure `go generate` has been run to update documentation
2. Create a git tag (e.g., `v1.2.3`) on the main branch
3. Push the tag — goreleaser will automatically build and publish binaries
4. GPG signing is configured in `.goreleaser.yml` (requires `GPG_FINGERPRINT` env var)
