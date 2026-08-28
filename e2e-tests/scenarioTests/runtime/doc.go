// Package runtime holds the behavioural assertions for the scenario suite.
//
// Every test file is behind the `e2e` build tag, so `go build ./...`, `go vet ./...`
// and the acceptance-test run never compile them. This file carries no tag, so the
// package always has at least one Go file and those commands do not fail with
// "build constraints exclude all Go files".
//
// Run via ./e2e-test.sh, which applies the Terraform stages first and passes their
// outputs in as KESTRA_E2E_* environment variables.
package runtime
